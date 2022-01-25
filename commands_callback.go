package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func CmdOnCallback(c *tb.Callback) {
	m := c.Message
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil {
		callbacklock.Lock()
		defer callbacklock.Unlock()

		cmds := strings.Split(strings.TrimSpace(c.Data), "/")
		cmd, gid, uid, secuid := "", int64(0), int64(0), int64(0)
		if len(cmds) > 0 {
			cmd = cmds[0]
		}
		if len(cmds) > 1 {
			gid, _ = strconv.ParseInt(cmds[1], 10, 64)
		}
		if len(cmds) > 2 {
			uid, _ = strconv.ParseInt(cmds[2], 10, 64)
		}
		if len(cmds) > 3 {
			secuid, _ = strconv.ParseInt(cmds[3], 10, 64)
		}
		triggerUid := c.Sender.ID
		vtToken := fmt.Sprintf("vt-%d,%d", gid, uid)
		joinVerificationId := fmt.Sprintf("join,%d,%d", gid, uid)
		isGroupAdmin := IsGroupAdmin(m.Chat, c.Sender)
		isMiaoGroupAdmin := IsGroupAdminMiaoKo(m.Chat, c.Sender)
		if strings.Contains("vt unban kick check rp lt", cmd) && IsGroup(gid) && uid > 0 {
			if cmd == "unban" && isGroupAdmin {
				if Unban(gid, uid, 0) == nil {
					Rsp(c, "cb.unban.success")
				} else {
					Rsp(c, "cb.unban.failure")
				}
				SmartEdit(m, m.Text+Locale("cb.unblock.byadmin", GetSenderLocaleCallback(c)))
				joinmap.Unset(joinVerificationId)
				if secuid > 0 && votemap.Exist(vtToken) {
					addCredit(gid, &tb.User{ID: uid}, 50, true)
					votemap.Unset(vtToken)
					addCredit(gid, &tb.User{ID: secuid}, -15, true)
				}
			} else if cmd == "kick" && isGroupAdmin {
				if Kick(gid, uid) == nil {
					Rsp(c, "cb.kick.success")
				} else {
					Rsp(c, "cb.kick.failure")
				}
				joinmap.Unset(joinVerificationId)
				votemap.Unset(vtToken)
				SmartEdit(m, m.Text+Locale("cb.kicked.byadmin", GetSenderLocaleCallback(c)))
			} else if cmd == "check" {
				if uid == c.Sender.ID {
					usrStatus := UserIsInGroup(gc.MustFollow, uid)
					if usrStatus == UIGIn {
						if Unban(gid, uid, 0) == nil {
							Bot.Delete(m)
							Rsp(c, "cb.validate.success")
							joinmap.Unset(joinVerificationId)
						} else {
							Rsp(c, "cb.validate.success.cannotUnban")
						}
					} else {
						Rsp(c, "cb.validate.failure")
					}
				} else {
					Rsp(c, "cb.validate.others")
				}
			} else if cmd == "vt" {
				userVtToken := fmt.Sprintf("vu-%d,%d,%d", gid, uid, c.Sender.ID)
				if _, ok := votemap.Get(vtToken); ok {
					if votemap.Add(userVtToken) == 1 {
						votes := votemap.Add(vtToken)
						if votes >= 6 {
							Unban(gid, uid, 0)
							votemap.Unset(vtToken)
							SmartEdit(m, m.Text+Locale("cb.unblock.byvote", GetSenderLocaleCallback(c)))
							addCredit(gid, &tb.User{ID: uid}, 50, true)
							if secuid > 0 {
								addCredit(gid, &tb.User{ID: secuid}, -15, true)
							}
						} else {
							EditBtns(m, m.Text, "", GenVMBtns(votes, gid, uid, secuid))
						}
						Rsp(c, "cb.vote.success")
					} else {
						Rsp(c, "cb.vote.failure")
					}
				} else {
					Rsp(c, "cb.vote.notExists")
				}
			} else if cmd == "rp" {
				redpacketKey := fmt.Sprintf("%d-%d", gid, secuid)

				credits, _ := redpacketmap.Get(redpacketKey)
				left, _ := redpacketnmap.Get(redpacketKey)
				if credits > 0 && left > 0 {
					redpacketBestKey := fmt.Sprintf("%d-%d:best", gid, secuid)
					redpacketUserKey := fmt.Sprintf("%d-%d:%d", gid, secuid, triggerUid)
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
							Rsp(c, "cb.rp.nothing")
						} else {
							lastBest, _ := redpacketmap.Get(redpacketBestKey)
							if amount > lastBest {
								redpacketmap.Set(redpacketBestKey, amount)
								redpacketrankmap.Set(redpacketBestKey, GetQuotableUserName(c.Sender))
							}
							Rsp(c, Locale("cb.rp.get.1", c.Sender.LanguageCode)+strconv.Itoa(amount)+Locale("cb.rp.get.2", GetSenderLocaleCallback(c)))
							addCredit(gid, c.Sender, int64(amount), true)
						}

						SendRedPacket(m, gid, secuid)
					} else {
						Rsp(c, "cb.rp.duplicated")
					}
				} else {
					Rsp(c, "cb.rp.notExists")
				}
			} else if cmd == "lt" {
				cmdtype := uid // 做了转换 1: lottery, 2: start, 3: draw
				lotteryId := cmds[3]
				li := GetLottery(lotteryId)
				if li != nil {
					if cmdtype == 2 && isMiaoGroupAdmin {
						li.Status = 0
						li.Update()
						li.UpdateTelegramMsg()
						Rsp(c, "cb.lottery.start")
					} else if cmdtype == 3 && isMiaoGroupAdmin {
						li.CheckDraw(true)
					} else if cmdtype == 1 {
						ci := GetCredit(li.GroupID, triggerUid)
						if ci != nil {
							if ci.Credit >= int64(li.Limit) {
								if li.Consume {
									addCredit(li.GroupID, c.Sender, -int64(li.Limit), true)
								}
								if err := li.Join(triggerUid, GetQuotableUserName(c.Sender)); err == nil {
									Rsp(c, "cb.lottery.enroll")
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
										addCredit(li.GroupID, c.Sender, int64(li.Limit), true)
									}
									Rsp(c, err.Error())
								}
							} else {
								Rsp(c, "cb.lottery.noEnoughCredit")
							}
						} else {
							Rsp(c, "cb.lottery.checkFailed")
						}
					} else {
						Rsp(c, "cb.notMiaoAdmin")
					}
				} else {
					Rsp(c, "cb.noEvent")
				}
			} else {
				Rsp(c, "cb.notAdmin")
			}
		} else {
			Rsp(c, "cb.notParsed")
		}
	} else {
		Rsp(c, "cb.disabled")
	}
}
