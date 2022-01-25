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

func GinParseGroupAndUser(c *gin.Context) *CreditInfo {
	groupId, _ := strconv.ParseInt(c.Param("groupId"), 10, 64)
	userId, _ := strconv.ParseInt(c.Param("userId"), 10, 64)

	if groupId < 0 && userId > 0 {
		DLogf("API Gateway | Route credit group=%d user=%d", groupId, userId)

		if gc := GetGroupConfig(groupId); gc != nil {
			if ci := GetCredit(groupId, userId); ci != nil && ci.ID == userId {
				return ci
			} else {
				c.JSON(http.StatusNotFound, GinError("the user does not exist."))
			}
		} else {
			c.JSON(http.StatusNotFound, GinError("the group does not exist."))
		}
	} else {
		c.JSON(http.StatusBadRequest, GinError("either groupId or userId is invalid."))
	}

	return nil
}

func MiddlewareAuthorization(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" || c.GetHeader("Authorization") == token {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, GinError("authorization failed."))
			c.Abort()
		}
	}
}

func InitRESTServer(portNum int, token string) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	authorized := r.Group("/")
	authorized.Use(MiddlewareAuthorization(token))

	apiVersionOne := authorized.Group("/v1")
	{
		credit := apiVersionOne.Group("/credit")
		credit.GET("/:groupId/:userId", func(c *gin.Context) {
			if ci := GinParseGroupAndUser(c); ci != nil {
				c.JSON(http.StatusOK, GinData(ci))
			}
		})
		credit.POST("/:groupId/:userId/consume", func(c *gin.Context) {
			consumeLock.Lock()
			defer consumeLock.Unlock()

			if ci := GinParseGroupAndUser(c); ci != nil {
				consumeRequest := struct {
					Credit int64 `json:"credit,omitempty"`
				}{}
				c.BindJSON(&consumeRequest)
				if consumeRequest.Credit > 0 {
					if ci.Credit >= consumeRequest.Credit {
						ci = UpdateCredit(ci, UMAdd, -consumeRequest.Credit)
						c.JSON(http.StatusOK, GinData(ci))
					} else {
						c.JSON(http.StatusNotAcceptable, GinError("the user does not have enough credit."))
					}
				} else {
					c.JSON(http.StatusBadRequest, GinError("consume credit should be a positive number."))
				}
			}
		})
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
