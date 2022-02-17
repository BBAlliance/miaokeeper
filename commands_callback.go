package main

import (
	"fmt"
	"math/rand"
	"strconv"

	tb "gopkg.in/tucnak/telebot.v2"
)

var callbackHandler *CallbackHandler

func CmdOnCallback(c *tb.Callback) {
	callbackHandler.Handle(c)
}

func InitCallback() {
	callbackHandler = &CallbackHandler{}

	callbackHandler.Add("vote", func(cp *CallbackParams) {
		gid, tuid := cp.GroupID(), cp.TriggerUserID()
		uid, _ := cp.GetUserId("u")
		secuid, _ := cp.GetUserId("s")

		vtToken := fmt.Sprintf("vt-%d,%d", gid, uid)
		userVtToken := fmt.Sprintf("vu-%d,%d,%d", gid, uid, tuid)

		if _, ok := votemap.Get(vtToken); ok {
			if votemap.Add(userVtToken) == 1 {
				votes := votemap.Add(vtToken)
				if votes >= 6 {
					Unban(gid, uid, 0)
					votemap.Unset(vtToken)
					SmartEdit(cp.Callback.Message, cp.Callback.Message.Text+Locale("cb.unblock.byvote", cp.Locale()))
					addCredit(gid, &tb.User{ID: uid}, 50, true, OPByAbuse)
					if secuid > 0 {
						addCredit(gid, &tb.User{ID: secuid}, -15, true, OPByAbuse)
					}
				} else {
					EditBtns(cp.Callback.Message, cp.Callback.Message.Text, "", GenVMBtns(votes, gid, uid, secuid))
				}
				cp.Response("cb.vote.success")
			} else {
				cp.Response("cb.vote.failure")
			}
		} else {
			cp.Response("cb.vote.notExists")
		}
	}).ShouldValidGroup(true).Should("u", "user")

	callbackHandler.Add("rp", func(cp *CallbackParams) {
		gid, tuid := cp.GroupID(), cp.TriggerUserID()
		rpKey, _ := cp.GetInt64("r")

		redpacketKey := fmt.Sprintf("%d-%d", gid, rpKey)
		credits, _ := redpacketmap.Get(redpacketKey)
		left, _ := redpacketnmap.Get(redpacketKey)
		if credits > 0 && left > 0 {
			redpacketBestKey := fmt.Sprintf("%d-%d:best", gid, rpKey)
			redpacketUserKey := fmt.Sprintf("%d-%d:%d", gid, rpKey, tuid)
			if redpacketmap.Add(redpacketUserKey) == 1 {
				amount := 0
				if left <= 1 {
					amount = credits
				} else if left == 2 {
					amount = rand.Intn(credits)
				} else {
					rate := 3
					if left <= 4 {
						rate = 2
					} else if left >= 12 {
						rate = 4
					}
					amount = rand.Intn(credits * rate / left)
				}
				redpacketnmap.Set(redpacketKey, left-1)
				redpacketmap.Set(redpacketKey, credits-amount)

				if amount == 0 {
					cp.Response("cb.rp.nothing")
				} else {
					lastBest, _ := redpacketmap.Get(redpacketBestKey)
					if amount > lastBest {
						redpacketmap.Set(redpacketBestKey, amount)
						redpacketrankmap.Set(redpacketBestKey, GetQuotableUserName(cp.TriggerUser()))
					}
					cp.Response(Locale("cb.rp.get.1", cp.TriggerUser().LanguageCode) + strconv.Itoa(amount) + Locale("cb.rp.get.2", cp.Locale()))
					addCredit(gid, cp.TriggerUser(), int64(amount), true, OPByRedPacket)
				}

				SendRedPacket(cp.Callback.Message, gid, rpKey)
			} else {
				cp.Response("cb.rp.duplicated")
			}
		} else {
			cp.Response("cb.rp.notExists")
		}
	}).ShouldValidGroup(true).Should("r", "int64").Lock("credit")

	callbackHandler.Add("unban", func(cp *CallbackParams) {
		gid := cp.GroupID()
		uid, _ := cp.GetUserId("u")
		secuid, _ := cp.GetUserId("s")

		joinVerificationId := fmt.Sprintf("join,%d,%d", gid, uid)
		vtToken := fmt.Sprintf("vt-%d,%d", gid, uid)

		if Unban(gid, uid, 0) == nil {
			cp.Response("cb.unban.success")
		} else {
			cp.Response("cb.unban.failure")
		}
		SmartEdit(cp.Callback.Message, cp.Callback.Message.Text+Locale("cb.unblock.byadmin", cp.Locale()))
		joinmap.Unset(joinVerificationId)
		if secuid > 0 && votemap.Exist(vtToken) {
			addCredit(gid, &tb.User{ID: uid}, 50, true, OPByAbuse)
			votemap.Unset(vtToken)
			addCredit(gid, &tb.User{ID: secuid}, -15, true, OPByAbuse)
		}
	}).ShouldValidGroupAdmin(true).Should("u", "user")

	callbackHandler.Add("kick", func(cp *CallbackParams) {
		gid := cp.GroupID()
		uid, _ := cp.GetUserId("u")

		joinVerificationId := fmt.Sprintf("join,%d,%d", gid, uid)
		vtToken := fmt.Sprintf("vt-%d,%d", gid, uid)

		if Kick(gid, uid) == nil {
			cp.Response("cb.kick.success")
		} else {
			cp.Response("cb.kick.failure")
		}
		joinmap.Unset(joinVerificationId)
		votemap.Unset(vtToken)
		SmartEdit(cp.Callback.Message, cp.Callback.Message.Text+Locale("cb.kicked.byadmin", cp.Locale()))
	}).ShouldValidGroupAdmin(true).Should("u", "user")

	callbackHandler.Add("check", func(cp *CallbackParams) {
		gid := cp.GroupID()
		uid, _ := cp.GetUserId("u")
		gc := cp.GroupConfig()

		joinVerificationId := fmt.Sprintf("join,%d,%d", gid, uid)

		if uid == cp.TriggerUserID() {
			usrStatus := UserIsInGroup(gc.MustFollow, uid)
			if usrStatus == UIGIn {
				if Unban(gid, uid, 0) == nil {
					Bot.Delete(cp.Callback.Message)
					cp.Response("cb.validate.success")
					joinmap.Unset(joinVerificationId)
				} else {
					cp.Response("cb.validate.success.cannotUnban")
				}
			} else {
				cp.Response("cb.validate.failure")
			}
		} else {
			cp.Response("cb.validate.others")
		}
	}).ShouldValidGroup(true).Should("u", "user")

	callbackHandler.Add("lt", func(cp *CallbackParams) {
		// 1: lottery, 2: start, 3: draw
		cmdtype, _ := cp.GetInt64("t")
		lotteryId, _ := cp.GetString("id")
		li := GetLottery(lotteryId)
		if li != nil {
			isMiaoGroupAdmin := IsGroupAdminMiaoKo(cp.Callback.Message.Chat, cp.TriggerUser())
			if cmdtype == 2 && isMiaoGroupAdmin {
				li.StartLottery()
				cp.Response("cb.lottery.start")
			} else if cmdtype == 3 && isMiaoGroupAdmin {
				li.CheckDraw(true)
			} else if cmdtype == 1 {
				ci := GetCredit(li.GroupID, cp.TriggerUserID())
				if ci != nil {
					if ci.Credit >= int64(li.Limit) {
						if li.Consume {
							addCredit(li.GroupID, cp.TriggerUser(), -int64(li.Limit), true, OPByLottery)
						}
						if err := li.Join(cp.TriggerUserID(), GetQuotableUserName(cp.TriggerUser())); err == nil {
							cp.Response("cb.lottery.enroll")
							if li.Participant > 0 {
								// check draw by particitant
								li.CheckDraw(false)
							}
							debouncer(func() {
								if li.Status == 0 {
									li.UpdateTelegramMsg()
								}
							})
						} else {
							if li.Consume {
								addCredit(li.GroupID, cp.TriggerUser(), int64(li.Limit), true, OPByLottery)
							}
							cp.Response(err.Error())
						}
					} else {
						cp.Response("cb.lottery.noEnoughCredit")
					}
				} else {
					cp.Response("cb.lottery.checkFailed")
				}
			} else {
				cp.Response("cb.notMiaoAdmin")
			}
		} else {
			cp.Response("cb.noEvent")
		}
	}).ShouldValidGroup(true).Should("t", "int64").Should("id", "string").Lock("credit")

}
