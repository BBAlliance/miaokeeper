package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

func SendRedPacket(to interface{}, chatId int64, packetId int64) (*tb.Message, error) {
	redpacketKey := fmt.Sprintf("%d-%d", chatId, packetId)
	credits, _ := redpacketmap.Get(redpacketKey)
	left, _ := redpacketnmap.Get(redpacketKey)
	sender, _ := redpacketrankmap.Get(redpacketKey + ":sender")

	msg := fmt.Sprintf(Locale("rp.complete", "zh"), sender)
	btns := []string{}

	if credits > 0 && left > 0 {
		creditLeft := strconv.Itoa(credits)
		if left == 1 {
			creditLeft = Locale("rp.guessLeft", "zh")
		}
		msg = fmt.Sprintf(Locale("rp.text", "zh"), sender, creditLeft, left)
		btns = []string{fmt.Sprintf(Locale("btn.rp.draw", "zh"), chatId, packetId)}
	}

	redpacketBestKey := fmt.Sprintf("%d-%d:best", chatId, packetId)
	if lastBest, _ := redpacketmap.Get(redpacketBestKey); lastBest > 0 {
		bestDrawer, _ := redpacketrankmap.Get(redpacketBestKey)
		msg += fmt.Sprintf(Locale("rp.lucky", "zh"), bestDrawer, lastBest)
	}

	if Type(to) == "*telebot.Message" {
		mess, _ := to.(*tb.Message)
		return EditBtnsMarkdown(mess, msg, "", btns)
	} else {
		return SendBtnsMarkdown(to, msg, "", btns)
	}
}

func CheckChannelForward(m *tb.Message) bool {
	if m.IsForwarded() {
		if gc := GetGroupConfig(m.Chat.ID); gc != nil && len(gc.BannedForward) > 0 {
			shouldDelete := (m.OriginalChat != nil && gc.IsBannedForward(m.OriginalChat.ID)) ||
				(m.OriginalSender != nil && gc.IsBannedForward(m.OriginalSender.ID))
			if shouldDelete {
				Bot.Delete(m)
				return false
			}
		}
	}
	return true
}

func CheckSpoiler(m *tb.Message) bool {
	if gc := GetGroupConfig(m.Chat.ID); gc != nil && gc.AntiSpoiler {
		for _, e := range m.Entities {
			if e.Type == "spoiler" {
				return true
			}
		}
	}
	return false
}

func CheckChannelFollow(m *tb.Message, user *tb.User, isJoin bool) bool {
	showExceptDialog := isJoin
	if gc := GetGroupConfig(m.Chat.ID); gc != nil && gc.MustFollow != "" {
		if isJoin && !gc.MustFollowOnJoin {
			return true
		}
		if !isJoin && !gc.MustFollowOnMsg {
			return true
		}
		usrName := GetQuotableUserName(user)

		// ignore bot
		if user.IsBot {
			if showExceptDialog {
				SmartSendDelete(m.Chat, fmt.Sprintf(Locale("channel.bot.permit", GetSenderLocale(m)), usrName))
			}
			return true
		}

		// ignore channel
		if !ValidMessageUser(m) {
			return true
		}

		usrStatus := UserIsInGroup(gc.MustFollow, user.ID)
		if usrStatus == UIGIn {
			if showExceptDialog {
				SmartSendDelete(m.Chat, fmt.Sprintf(Locale("channel.user.alreadyFollowed", GetSenderLocale(m)), usrName))
			}
		} else if usrStatus == UIGOut {
			chatId, userId := m.Chat.ID, user.ID
			joinVerificationId := fmt.Sprintf("join,%d,%d", chatId, userId)
			if joinmap.Add(joinVerificationId) > 1 {
				// already in verification process
				Bot.Delete(m)
				return false
			}
			msg, err := SendBtnsMarkdown(m.Chat, fmt.Sprintf(Locale("channel.request", GetSenderLocale(m)), userId, usrName), "", []string{
				fmt.Sprintf(Locale("btn.channel.step1", GetSenderLocale(m)), strings.TrimLeft(gc.MustFollow, "@")),
				fmt.Sprintf(Locale("btn.channel.step2", GetSenderLocale(m)), chatId, userId),
				fmt.Sprintf(Locale("btn.adminPanel", GetSenderLocale(m)), chatId, userId, 0, chatId, userId, 0),
			})
			if msg == nil || err != nil {
				if showExceptDialog {
					SmartSendDelete(m.Chat, Locale("channel.cannotSendMsg", GetSenderLocale(m)))
				}
				joinmap.Unset(joinVerificationId)
			} else {
				if Ban(chatId, userId, 0) != nil {
					LazyDelete(msg)
					if showExceptDialog {
						SmartSendDelete(m.Chat, Locale("channel.cannotBanUser", GetSenderLocale(m)))
					}
					joinmap.Unset(joinVerificationId)
				} else {
					time.AfterFunc(time.Minute*5, func() {
						Bot.Delete(msg)
						if joinmap.Exist(joinVerificationId) {
							cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
							if err != nil || cm.Role == tb.Restricted || cm.Role == tb.Kicked || cm.Role == tb.Left {
								KickOnce(chatId, userId)
								SmartSend(m.Chat, fmt.Sprintf(Locale("channel.kicked", GetSenderLocale(m)), userId), &tb.SendOptions{
									ParseMode:             "Markdown",
									DisableWebPagePreview: true,
									AllowWithoutReply:     true,
								})
							}
						}
					})
					Bot.Delete(m)
					return false
				}
			}
		} else {
			if showExceptDialog {
				SmartSendDelete(m.Chat, Locale("channel.cannotCheckChannel", GetSenderLocale(m)))
			}
		}
	}
	return true
}

func Rsp(c *tb.Callback, msg string) {
	Bot.Respond(c, &tb.CallbackResponse{
		Text:      Locale(msg, c.Sender.LanguageCode),
		ShowAlert: true,
	})
}

func GenVMBtns(votes int, chatId, userId, secondUserId int64) []string {
	return []string{
		fmt.Sprintf(Locale("btn.notFair", "zh"), votes, chatId, userId, secondUserId),
		fmt.Sprintf(Locale("btn.adminPanel", "zh"), chatId, userId, secondUserId, chatId, userId, secondUserId),
	}
}

func addCreditToMsgSender(chatId int64, m *tb.Message, credit int64, force bool) *CreditInfo {
	if ValidMessageUser(m) {
		return addCredit(chatId, m.Sender, credit, force)
	}
	return nil
}

func addCredit(chatId int64, user *tb.User, credit int64, force bool) *CreditInfo {
	if chatId < 0 && user != nil && user.ID > 0 && credit != 0 {
		token := fmt.Sprintf("ac-%d-%d", chatId, user.ID)
		if creditomap.Add(token) < 20 || force { // can only get credit 20 times / hour
			return UpdateCredit(BuildCreditInfo(chatId, user, false), UMAdd, credit)
		}
	}
	return nil
}

func ValidReplyUser(m *tb.Message) bool {
	return m.ReplyTo != nil && ValidMessageUser(m) && ValidMessageUser(m.ReplyTo) && m.ReplyTo.Sender.ID != m.Sender.ID
}

func ValidMessageUser(m *tb.Message) bool {
	return m != nil && m.SenderChat == nil && ValidUser(m.Sender)
}

func ValidUser(u *tb.User) bool {
	return u != nil && u.ID > 0 && !u.IsBot && u.ID != 777000 && u.Username != "Channel_Bot" && u.Username != "GroupAnonymousBot" && u.Username != "Telegram"
}

func BuildCreditInfo(groupId int64, user *tb.User, autoFetch bool) *CreditInfo {
	ci := &CreditInfo{
		user.Username, GetUserName(user), user.ID, 0, groupId,
	}
	if autoFetch {
		ci.Credit = GetCredit(groupId, user.ID).Credit
	}
	return ci
}

func SmartEdit(to *tb.Message, what interface{}, options ...interface{}) (*tb.Message, error) {
	if len(options) == 0 {
		options = append([]interface{}{&tb.SendOptions{
			// ParseMode:             "Markdown",
			DisableWebPagePreview: true,
			AllowWithoutReply:     true,
		}}, options...)
	}
	m, err := Bot.Edit(to, what, options...)
	if err != nil {
		DErrorE(err, "Telegram Edit Error")
	}
	return m, err
}

func SmartSendDelete(to interface{}, what interface{}, options ...interface{}) (*tb.Message, error) {
	msg, err := SmartSend(to, what, options...)
	if err == nil && msg != nil {
		LazyDelete(msg)
	}
	return msg, err
}

func MakeBtns(prefix string, btns []string) [][]tb.InlineButton {
	btnsc := make([][]tb.InlineButton, 0)
	for _, row := range btns {
		btnscr := make([]tb.InlineButton, 0)
		for _, btn := range strings.Split(row, "||") {
			z := strings.SplitN(btn, "|", 2)
			if len(z) < 2 {
				continue
			}
			unique := ""
			link := ""
			if _, err := url.Parse(z[1]); err == nil && strings.HasPrefix(z[1], "https://") {
				link = z[1]
			} else {
				unique = prefix + z[1]
			}
			btnscr = append(btnscr, tb.InlineButton{
				Unique: unique,
				Text:   z[0],
				Data:   "",
				URL:    link,
			})
		}
		btnsc = append(btnsc, btnscr)
	}
	return btnsc
}

func SendBtns(to interface{}, what interface{}, prefix string, btns []string) (*tb.Message, error) {
	return SmartSendInner(to, what, &tb.SendOptions{
		// ParseMode:             "Markdown",
		DisableWebPagePreview: true,
		AllowWithoutReply:     true,
	}, &tb.ReplyMarkup{
		OneTimeKeyboard:     true,
		ResizeReplyKeyboard: true,
		ForceReply:          false,
		InlineKeyboard:      MakeBtns(prefix, btns),
	})
}

func SendBtnsMarkdown(to interface{}, what interface{}, prefix string, btns []string) (*tb.Message, error) {
	return SmartSendInner(to, what, &tb.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
		AllowWithoutReply:     true,
	}, &tb.ReplyMarkup{
		OneTimeKeyboard:     true,
		ResizeReplyKeyboard: true,
		ForceReply:          false,
		InlineKeyboard:      MakeBtns(prefix, btns),
	})
}

func EditBtns(to *tb.Message, what interface{}, prefix string, btns []string) (*tb.Message, error) {
	return SmartEdit(to, what, &tb.ReplyMarkup{
		OneTimeKeyboard:     true,
		ResizeReplyKeyboard: true,
		ForceReply:          false,
		InlineKeyboard:      MakeBtns(prefix, btns),
	})
}

func EditBtnsMarkdown(to *tb.Message, what interface{}, prefix string, btns []string) (*tb.Message, error) {
	return SmartEdit(to, what, &tb.SendOptions{
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
		AllowWithoutReply:     true,
	}, &tb.ReplyMarkup{
		OneTimeKeyboard:     true,
		ResizeReplyKeyboard: true,
		ForceReply:          false,
		InlineKeyboard:      MakeBtns(prefix, btns),
	})
}

func SmartSend(to interface{}, what interface{}, options ...interface{}) (*tb.Message, error) {
	if len(options) == 0 {
		return SmartSendInner(to, what, &tb.SendOptions{
			// ParseMode:             "Markdown",
			DisableWebPagePreview: true,
			AllowWithoutReply:     true,
		})
	}
	return SmartSendInner(to, what, options...)
}

func extractMessage(data []byte) (*tb.Message, error) {
	var resp struct {
		Result *tb.Message
	}
	if err := jsoniter.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func RevealSpoiler(msg *tb.Message) (*tb.Message, error) {
	var params = make(map[string]interface{})
	msgID, chatID := msg.MessageSig()
	params["chat_id"] = strconv.FormatInt(chatID, 10)
	params["reply_to_message_id"] = msgID
	params["disable_web_page_preview"] = true
	params["allow_sending_without_reply"] = true
	params["protect_content"] = true
	params["text"] = msg.Text

	for i, e := range msg.Entities {
		if e.Type == "spoiler" {
			msg.Entities[i].Type = tb.EntityStrikethrough
		}
	}
	params["entities"] = msg.Entities

	data, err := Bot.Raw("sendMessage", params)
	if err != nil {
		return nil, err
	}

	return extractMessage(data)
}

func SmartSendInner(to interface{}, what interface{}, options ...interface{}) (*tb.Message, error) {
	toType := Type(to)
	var m *tb.Message = nil
	var err error = nil
	if toType == "*telebot.Message" {
		mess, _ := to.(*tb.Message)
		m, err = Bot.Reply(mess, what, options...)
	} else if toType == "*telebot.Chat" {
		recp, _ := to.(*tb.Chat)
		if recp != nil {
			m, err = Bot.Send(recp, what, options...)
		} else {
			err = errors.New("chat is empty")
		}
	} else if toType == "*telebot.User" {
		recp, _ := to.(*tb.User)
		if recp != nil {
			m, err = Bot.Send(recp, what, options...)
		} else {
			err = errors.New("user is empty")
		}
	} else if toType == "int64" {
		recp, _ := to.(int64)
		m, err = Bot.Send(&tb.Chat{ID: recp}, what, options...)
	} else {
		err = errors.New("unknown type of message: " + toType)
	}
	if err != nil {
		DErrorE(err, "TeleBot Message Error")
	}
	return m, err
}

func GetUserName(u *tb.User) string {
	s := ""
	if u.FirstName != "" || u.LastName != "" {
		s = strings.TrimSpace(u.FirstName + " " + u.LastName)
	} else if u.Username != "" {
		s = "@" + u.Username
	}

	return s
}

func GetQuotableStr(s string) string {
	return strings.ReplaceAll(s, "`", "'")
}

func GetQuotableUserName(u *tb.User) string {
	return GetQuotableStr(GetUserName(u))
}

func GetChatName(u *tb.Chat) string {
	s := ""
	if u.FirstName != "" || u.LastName != "" {
		s = strings.TrimSpace(u.FirstName + " " + u.LastName)
	} else if u.Username != "" {
		s = "@" + u.Username
	}

	return s
}

func UserIsInGroup(chatRepr string, userId int64) UIGStatus {
	cm, err := ChatMemberOf(chatRepr, Bot.Me.ID)
	if err != nil {
		return UIGErr
	} else if cm.Role != tb.Administrator && cm.Role != tb.Creator {
		return UIGErr
	}

	if userId == Bot.Me.ID {
		return UIGIn
	}

	cm, err = ChatMemberOf(chatRepr, userId)
	// if is admin, pass
	if cm.Anonymous || cm.Role == tb.Administrator || cm.Role == tb.Creator {
		return UIGIn
	}

	if err != nil || cm == nil {
		return UIGOut
	}
	if cm.Role == tb.Left || cm.Role == tb.Kicked {
		return UIGOut
	}
	return UIGIn
}

func ChatMemberOf(chatRepr string, userId int64) (*tb.ChatMember, error) {
	params := map[string]string{
		"chat_id": chatRepr,
		"user_id": strconv.FormatInt(userId, 10),
	}

	data, err := Bot.Raw("getChatMember", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result *tb.ChatMember
	}
	if err := jsoniter.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func Kick(chatId, userId int64) error {
	cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
	if err == nil {
		return Bot.Ban(&tb.Chat{ID: chatId}, cm)
	}
	return err
}

func KickOnce(chatId, userId int64) error {
	cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
	if err == nil {
		err = Bot.Ban(&tb.Chat{ID: chatId}, cm)
		if err == nil {
			return Bot.Unban(&tb.Chat{ID: chatId}, &tb.User{ID: userId}, true)
		}
	}
	return err
}

func Ban(chatId, userId int64, duration int64) error {
	cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
	if err == nil {
		cm.CanSendMessages = false
		cm.CanSendMedia = false
		cm.CanSendOther = false
		cm.CanAddPreviews = false
		cm.CanSendPolls = false
		cm.CanInviteUsers = false
		cm.CanPinMessages = false
		cm.CanChangeInfo = false

		cm.RestrictedUntil = time.Now().Unix() + duration
		return RestrictChatMember(&tb.Chat{ID: chatId}, cm)
	}
	return err
}

func Unban(chatId, userId int64, duration int64) error {
	cm, err := Bot.ChatMemberOf(&tb.Chat{ID: chatId}, &tb.User{ID: userId})
	if err == nil {
		cm.CanSendMessages = true
		cm.CanSendMedia = true
		cm.CanSendOther = true
		cm.CanAddPreviews = true
		cm.CanSendPolls = true
		cm.CanInviteUsers = true
		cm.CanPinMessages = true
		cm.CanChangeInfo = true
		return RestrictChatMember(&tb.Chat{ID: chatId}, cm)
	}
	return err
}

func RestrictChatMember(chat *tb.Chat, member *tb.ChatMember) error {
	rights, until := member.Rights, member.RestrictedUntil

	params := map[string]interface{}{
		"chat_id":     chat.Recipient(),
		"user_id":     member.User.Recipient(),
		"permissions": &map[string]bool{},
		"until_date":  strconv.FormatInt(until, 10),
	}

	data, _ := jsoniter.Marshal(rights)
	_ = jsoniter.Unmarshal(data, params["permissions"])
	_, err := Bot.Raw("restrictChatMember", params)
	return err
}

func BanChannel(chatId, channelId int64) error {
	params := map[string]interface{}{
		"chat_id":        strconv.FormatInt(chatId, 10),
		"sender_chat_id": strconv.FormatInt(channelId, 10),
	}

	_, err := Bot.Raw("banChatSenderChat", params)
	return err
}

func SetCommands() error {
	allCommands := [][]string{
		{"mycredit", "获取自己的积分"},
		{"redpacket", "用自己的积分发红包，发 N (10~1000) 分给 K (1~20) 个人"},
		{"creditrank", "获取积分排行榜前 N 名"},
		{"lottery", "在积分排行榜前 N 名内抽出 K 名幸运儿"},
	}
	cmds := []tb.Command{}
	for _, cmd := range allCommands {
		cmds = append(cmds, tb.Command{
			Text:        cmd[0],
			Description: cmd[1],
		})
	}
	return Bot.SetCommands(cmds)
}

func IsGroup(gid int64) bool {
	return I64In(&GROUPS, gid)
}

func IsAdmin(uid int64) bool {
	return I64In(&ADMINS, uid)
}

func IsGroupAdmin(c *tb.Chat, u *tb.User) bool {
	isGAS := IsGroupAdminMiaoKo(c, u)
	if isGAS {
		return true
	}
	return IsGroupAdminTelegram(c, u)
}

func IsGroupAdminMiaoKo(c *tb.Chat, u *tb.User) bool {
	gc := GetGroupConfig(c.ID)
	return gc != nil && gc.IsAdmin(u.ID)
}

func IsGroupAdminTelegram(c *tb.Chat, u *tb.User) bool {
	cm, _ := Bot.ChatMemberOf(c, u)
	if cm != nil && (cm.Role == tb.Administrator || cm.Role == tb.Creator) {
		return true
	}
	return false
}

func LazyDelete(m *tb.Message) {
	time.AfterFunc(time.Second*10, func() {
		Bot.Delete(m)
	})
}

// func StartCountDown() {
// 	chat := int64(-1001270914368) // miao group
// 	// chat := int64(-1001681365705) // test group
// 	target := int64(1640408400)
// 	if target-time.Now().UnixMilli()/1000 < 0 {
// 		return
// 	}
// 	c := &tb.Chat{ID: chat}
// 	msg, _ := SmartSend(c, "🎄 EST 时区圣诞节倒计时已激活 ～")
// 	Bot.Pin(msg)

// 	for {
// 		time.Sleep(time.Second - time.Millisecond*10)
// 		ct := target - time.Now().UnixMilli()/1000
// 		if ct >= 3600 {
// 			if ct%3600 == 0 {
// 				go SmartEdit(msg, fmt.Sprintf("🎄 还有 %d 小时 EST 时区圣诞倒计时开始", ct/3600))
// 			}
// 		} else if ct >= 600 {
// 			if ct%600 == 0 {
// 				go SmartEdit(msg, fmt.Sprintf("🎄 还有 %d 分钟 EST 时区圣诞倒计时开始", ct/60))
// 			}
// 		} else if ct >= 60 {
// 			if ct%60 == 0 {
// 				Bot.Delete(msg)
// 				msg, _ = SmartSend(chat, fmt.Sprintf("🎄 还有 %d 分钟 EST 时区圣诞倒计时开始", ct/60))
// 				Bot.Pin(msg)
// 			}
// 		} else if ct > 0 && ct <= 10 {
// 			go SmartEdit(msg, fmt.Sprintf("🎄 倒计时开始！距离 EST 圣诞节还有 %d 秒 EST ～", ct))
// 		} else if ct <= 0 {
// 			go SmartEdit(msg, "🎄 各位喵群的小伙伴们！！！圣诞节快乐～～～")
// 			return
// 		}
// 	}
// }
