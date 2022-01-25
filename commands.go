package main

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

func CmdWarnUser(m *tb.Message) {
	gc := GetGroupConfig(m.Chat.ID)
	if gc != nil && m.ReplyTo != nil {
		if gc.DisableWarn {
			SmartSendDelete(m.ReplyTo, Locale("cmd.zc.notAllowed", GetSenderLocale(m)))
			return
		}
		if m.Sender.ID > 0 && m.Sender.Username != "Channel_Bot" {
			if m.ReplyTo.Sender.ID == m.Sender.ID {
				SmartSend(m, Locale("cmd.zc.indeed", GetSenderLocale(m)))
			} else if m.ReplyTo.Sender.ID < 0 || m.ReplyTo.Sender.IsBot {
				SmartSend(m, Locale("cmd.zc.cantBan", GetSenderLocale(m)))
			} else {
				token := fmt.Sprintf("%d,%d,%d", m.Chat.ID, m.Sender.ID, m.ReplyTo.Sender.ID)
				limSenderToken := fmt.Sprintf("lim%d,%d,%d", m.Chat.ID, m.Sender.ID)
				limReciverToken := fmt.Sprintf("lim%d,%d,%d", m.Chat.ID, m.ReplyTo.Sender.ID)
				if _, ok := zcomap.Get(token); ok {
					addCredit(m.Chat.ID, m.Sender, -10, true)
					SmartSend(m, Locale("cmd.zc.cooldown10", GetSenderLocale(m)))
				} else if senderLimit, _ := zcomap.Get(limSenderToken); senderLimit >= 2 {
					zcomap.Add(limReciverToken)
					SmartSend(m, Locale("cmd.zc.cooldown", GetSenderLocale(m)))
				} else {
					zcomap.Add(limSenderToken)
					zcomap.Add(limReciverToken)
					zcomap.Set(token, 1)
					ci := addCredit(m.Chat.ID, m.ReplyTo.Sender, -25, true)
					SmartSend(m.ReplyTo, fmt.Sprintf(Locale("cmd.zc.exec", GetSenderLocale(m)), GetUserName(m.ReplyTo.Sender), GetUserName(m.Sender)))
					LazyDelete(m)
					if ci.Credit < -50 {
						Ban(m.Chat.ID, m.ReplyTo.Sender.ID, 0)
					}
				}
			}
		} else {
			SmartSend(m, Locale("cmd.zc.noAnonymous", GetSenderLocale(m)))
		}
	}
}

func CmdBanUser(m *tb.Message) {
	if IsGroup(m.Chat.ID) && m.ReplyTo != nil {
		if m.Sender.ID > 0 && m.Sender.Username != "Channel_Bot" {
			if m.ReplyTo.Sender.ID == m.Sender.ID {
				if Ban(m.Chat.ID, m.Sender.ID, 1800) == nil {
					SmartSend(m, Locale("cmd.ey.selfReport", GetSenderLocale(m)))
					LazyDelete(m.ReplyTo)
				} else {
					SmartSend(m, Locale("cmd.ey.notSuccess", GetSenderLocale(m)))
				}
			} else if m.ReplyTo.Sender.IsBot && m.ReplyTo.SenderChat != nil {
				if m.ReplyTo.SenderChat != nil && m.ReplyTo.SenderChat.ID != m.Chat.ID {
					if BanChannel(m.Chat.ID, m.ReplyTo.SenderChat.ID) == nil {
						SmartSend(m, fmt.Sprintf(Locale("cmd.ey.killChannel", GetSenderLocale(m)), GetChatName(m.ReplyTo.SenderChat)))
						LazyDelete(m)
						LazyDelete(m.ReplyTo)
					} else {
						SmartSend(m, Locale("cmd.ey.notSuccess", GetSenderLocale(m)))
					}
				} else {
					SmartSend(m, Locale("cmd.ey.unexpected", GetSenderLocale(m)))
				}
			} else if m.ReplyTo.Sender.IsBot {
				if Ban(m.Chat.ID, m.ReplyTo.Sender.ID, 1800) == nil {
					SmartSend(m, fmt.Sprintf(Locale("cmd.ey.killBot", GetSenderLocale(m)), GetUserName(m.ReplyTo.Sender)))
					LazyDelete(m)
					LazyDelete(m.ReplyTo)
				} else {
					SmartSend(m, Locale("cmd.ey.notSuccess", GetSenderLocale(m)))
				}
			} else {
				userId := m.ReplyTo.Sender.ID
				vtToken := fmt.Sprintf("vt-%d,%d", m.Chat.ID, userId)
				token := fmt.Sprintf("ad-%d,%d", m.Chat.ID, m.Sender.ID)
				if zcomap.Add(token) > 3 {
					addCredit(m.Chat.ID, m.Sender, -5, true)
					SmartSend(m, Locale("cmd.ey.cooldown5", GetSenderLocale(m)))
				} else {
					if _, ok := votemap.Get(vtToken); !ok {
						if Ban(m.Chat.ID, userId, 1800) == nil {
							addCredit(m.Chat.ID, m.ReplyTo.Sender, -50, true)
							addCredit(m.Chat.ID, m.Sender, 15, true)
							votemap.Set(vtToken, 0)
							msgTxt := fmt.Sprintf(Locale("cmd.ey.exec", GetSenderLocale(m)), GetUserName(m.ReplyTo.Sender), GetUserName(m.Sender))
							SendBtns(m.ReplyTo, msgTxt, "", GenVMBtns(0, m.Chat.ID, userId, m.Sender.ID))
							LazyDelete(m)
							LazyDelete(m.ReplyTo)
						} else {
							SmartSend(m, Locale("cmd.ey.notSuccess", GetSenderLocale(m)))
						}
					} else {
						SmartSend(m, Locale("cmd.ey.duplicated", GetSenderLocale(m)))
					}
				}
			}
		} else {
			SmartSend(m, Locale("cmd.zc.noAnonymous", GetSenderLocale(m)))
		}
	}
}
