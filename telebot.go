package main

import (
	"math/rand"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/bep/debounce"
	tb "gopkg.in/tucnak/telebot.v2"
)

type UIGStatus int

const (
	UIGIn UIGStatus = iota
	UIGOut
	UIGErr
)

var Bot *tb.Bot
var TOKEN = ""
var TELEGRAMURL = ""

var GROUPS = []int64{}
var ADMINS = []int64{}

var lastID = int64(-1)
var lastText = ""
var puncReg *regexp.Regexp

var zcomap *ObliviousMapInt
var creditomap *ObliviousMapInt
var votemap *ObliviousMapInt

var joinmap *ObliviousMapInt

var redpacketrankmap *ObliviousMapStr
var redpacketmap *ObliviousMapInt
var redpacketnmap *ObliviousMapInt

var debouncer func(func())

var callbacklock sync.Mutex
var userredpacketlock sync.Mutex

func InitTelegram() {
	var err error
	Bot, err = tb.NewBot(tb.Settings{
		Token: TOKEN,
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
			AllowedUpdates: []string{
				"message",
				"edited_message",
				// "channel_post",
				// "edited_channel_post",
				// "inline_query",
				// "chosen_inline_result",
				"callback_query",
				// "shipping_query",
				// "pre_checkout_query",
				// "poll",
				// "poll_answer",
				"my_chat_member",
				"chat_member",
				// "chat_join_request",
			},
		},
		URL: TELEGRAMURL,
	})

	if !ping && !escape {

		if !Bot.Me.CanJoinGroups {
			DErrorf("TeleBot Error | bot cannot be added to group, please check with @botfather")
			os.Exit(1)
		}

		if !Bot.Me.CanReadMessages {
			DErrorf("TeleBot Error | bot cannot be run under privacy mode, please check with @botfather")
			os.Exit(1)
		}

		if err != nil {
			DErrorE(err, "TeleBot Error | cannot initialize telegram bot")
			os.Exit(1)
		}

		err = SetCommands()
		if err != nil {
			DErrorE(err, "TeleBot Error | cannot update commands for telegram bot")
		}

		// ---------------- Super Admin ----------------

		Bot.Handle("/su_export_credit", CmdSuExportCredit)
		Bot.Handle("/su_add_group", CmdSuAddGroup)
		Bot.Handle("/su_del_group", CmdSuDelGroup)
		Bot.Handle("/su_add_admin", CmdSuAddAdmin)
		Bot.Handle("/su_del_admin", CmdSuDelAdmin)

		// ---------------- Group Admin ----------------

		Bot.Handle("/add_admin", CmdAddAdmin)
		Bot.Handle("/del_admin", CmdDelAdmin)
		Bot.Handle("/ban_forward", CmdBanForward)
		Bot.Handle("/unban_forward", CmdUnbanForward)
		Bot.Handle("/set_credit", CmdSetCredit)
		Bot.Handle("/add_credit", CmdAddCredit)
		Bot.Handle("/check_credit", CmdCheckCredit)
		Bot.Handle("/set_antispoiler", CmdSetAntiSpoiler)
		Bot.Handle("/set_channel", CmdSetChannel)
		Bot.Handle("/set_locale", CmdSetLocale)
		Bot.Handle("/send_redpacket", CmdSendRedpacket)
		Bot.Handle("/create_lottery", CmdCreateLottery)

		Bot.Handle("/creditrank", CmdCreditRank)
		Bot.Handle("/redpacket", CmdRedpacket)
		Bot.Handle("/lottery", CmdLottery)

		// ---------------- Normal User ----------------

		Bot.Handle("/ban_user", CmdBanUserCommand)
		Bot.Handle("/unban_user", CmdUnbanUserCommand)
		Bot.Handle("/kick_user", CmdKickUserCommand)

		Bot.Handle("/mycredit", CmdMyCredit)
		Bot.Handle("/version", CmdVersion)
		Bot.Handle("/ping", CmdPing)

		Bot.Handle("口臭", CmdWarnUser)
		Bot.Handle("口 臭", CmdWarnUser)
		Bot.Handle("口  臭", CmdWarnUser)
		Bot.Handle("口臭!", CmdWarnUser)
		Bot.Handle("口臭！", CmdWarnUser)
		Bot.Handle("嘴臭", CmdWarnUser)
		Bot.Handle("嘴 臭", CmdWarnUser)
		Bot.Handle("嘴  臭", CmdWarnUser)
		Bot.Handle("嘴臭!", CmdWarnUser)
		Bot.Handle("嘴臭！", CmdWarnUser)

		Bot.Handle("恶意广告", CmdBanUser)
		Bot.Handle("惡意廣告", CmdBanUser)
		Bot.Handle("恶意发言", CmdBanUser)
		Bot.Handle("惡意發言", CmdBanUser)
		Bot.Handle("恶意举报", CmdBanUser)
		Bot.Handle("惡意舉報", CmdBanUser)
		Bot.Handle("惡意檢舉", CmdBanUser)
		Bot.Handle("恶意嘴臭", CmdBanUser)
		Bot.Handle("恶意口臭", CmdBanUser)

		Bot.Handle(tb.OnUserLeft, CmdOnUserLeft)
		Bot.Handle(tb.OnChatMember, CmdOnChatMember)
		Bot.Handle(tb.OnUserJoined, CmdOnUserJoined)
		Bot.Handle(tb.OnPinned, CmdOnPinned)

		Bot.Handle(tb.OnCallback, CmdOnCallback)
		Bot.Handle(tb.OnPhoto, CmdOnMisc)
		Bot.Handle(tb.OnAnimation, CmdOnMisc)
		Bot.Handle(tb.OnVideo, CmdOnMisc)
		Bot.Handle(tb.OnEdited, CmdOnMisc)
		Bot.Handle(tb.OnDocument, CmdOnDocument)

		Bot.Handle(tb.OnText, CmdOnText)
		Bot.Handle(tb.OnSticker, CmdOnSticker)

	}

	go Bot.Start()

	if !ping {
		// go StartCountDown()
		DInfo("MiaoKeeper is up.")
	}

	if escape {
		DInfo("Escape mode is on.")
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	puncReg = regexp.MustCompile(`^[!"#$%&'()*+,-./:;<=>?@[\]^_{|}~` + "`" + `][a-zA-Z0-9]+`)
	zcomap = NewOMapInt(60*60*1000, true)
	creditomap = NewOMapInt(60*60*1000, false)
	votemap = NewOMapInt(30*60*1000, false)

	joinmap = NewOMapInt(5*60*1000+30*1000, false)

	redpacketrankmap = NewOMapStr(24*60*60*1000, false)
	redpacketmap = NewOMapInt(24*60*60*1000, false)
	redpacketnmap = NewOMapInt(24*60*60*1000, false)
	debouncer = debounce.New(time.Second)
}
