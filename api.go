package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"github.com/gin-gonic/gin"
)

var consumeLock sync.Mutex

func GinError(err string) gin.H {
	return gin.H{
		"status": false,
		"error":  err,
	}
}

func GinData(data interface{}) gin.H {
	return gin.H{
		"status": true,
		"data":   data,
	}
}

func GinParseUser(c *gin.Context) *CreditInfo {
	userId, _ := strconv.ParseInt(c.Param("userId"), 10, 64)
	if groupId := c.GetInt64("gid"); groupId != 0 {
		if ci := GetCredit(groupId, userId); ci != nil && ci.ID == userId {
			return ci
		} else {
			c.JSON(http.StatusNotFound, GinError("the user does not exist."))
		}
	} else {
		c.JSON(http.StatusNotFound, GinError("unexpected error."))
	}

	return nil
}

func MiddlewareGroupAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupId, _ := strconv.ParseInt(c.Param("groupId"), 10, 64)
		if groupId < 0 {
			if gc := GetGroupConfig(groupId); gc != nil {
				if gc.GenerateSign(GST_API_SIGN) == c.GetHeader("Authorization") {
					c.Set("gc", gc)
					c.Set("gid", groupId)

					c.Next()
					return
				} else {
					c.JSON(http.StatusUnauthorized, GinError("authorization failed."))
				}
			} else {
				c.JSON(http.StatusNotFound, GinError("the group does not exist."))
			}
		} else {
			c.JSON(http.StatusBadRequest, GinError("the group id is not valid."))
		}

		c.Abort()
	}
}

func InitRESTServer(portNum int) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, GinError("endpoint not found"))
	})

	apiVersionOne := r.Group("/v1")
	{
		authorizedGroup := apiVersionOne.Group("/group/:groupId")
		authorizedGroup.Use(MiddlewareGroupAuthorization())
		{
			credit := authorizedGroup.Group("/credit")
			credit.GET("/:userId", func(c *gin.Context) {
				if ci := GinParseUser(c); ci != nil {
					c.JSON(http.StatusOK, GinData(ci))
				}
			})
			credit.POST("/:userId/consume", func(c *gin.Context) {
				consumeLock.Lock()
				defer consumeLock.Unlock()

				if ci := GinParseUser(c); ci != nil {
					consumeRequest := struct {
						Credit        int64 `json:"credit,omitempty"`
						AllowNegative bool  `json:"allowNegative,omitempty"`
					}{}
					c.BindJSON(&consumeRequest)
					if consumeRequest.Credit > 0 {
						if ci.Credit >= consumeRequest.Credit || (ci.Credit > 0 && consumeRequest.AllowNegative) {
							ci = UpdateCredit(ci, UMAdd, -consumeRequest.Credit, OPByAPIConsume)
							c.JSON(http.StatusOK, GinData(ci))
						} else {
							c.JSON(http.StatusNotAcceptable, GinError("the user does not have enough credit."))
						}
					} else {
						c.JSON(http.StatusBadRequest, GinError("consume credit should be a positive number."))
					}
				}
			})
			credit.POST("/:userId/bonus", func(c *gin.Context) {
				consumeLock.Lock()
				defer consumeLock.Unlock()

				if ci := GinParseUser(c); ci != nil {
					bonusRequest := struct {
						Credit        int64 `json:"credit,omitempty"`
					}{}
					c.BindJSON(&bonusRequest)
					if bonusRequest.Credit > 0 {
						ci = UpdateCredit(ci, UMAdd, bonusRequest.Credit, OPByAPIBonus)
						c.JSON(http.StatusOK, GinData(ci))
					} else {
						c.JSON(http.StatusBadRequest, GinError("added credit should be a positive number."))
					}
				}
			})
		}
	}

	go func() {
		err := r.Run(fmt.Sprintf(":%d", portNum))
		if err != nil {
			DErrorf("API Gateway Error | unexcepted happens: %s", err.Error())
			os.Exit(1)
		}
	}()

	DInfo("API Gateway | API server started")
}
