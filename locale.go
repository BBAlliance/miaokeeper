package main

import tb "gopkg.in/tucnak/telebot.v2"

var LocaleAlias = map[string]string{
	"zh-hans": "zh",
	"zh-hant": "zh",
}

var LocaleMap = map[string]map[string]string{
	"zh": {
		"system.unexpected": "❌ 无法完成任务，请检查服务器错误日志",

		"cmd.zc.notAllowed":  "当前群组不允许互相臭嘴哦 ~",
		"cmd.zc.indeed":      "确实",
		"cmd.zc.cantBan":     "我拿它没办法呢 ...",
		"cmd.zc.cooldown10":  "😠 你自己先漱漱口呢，不要连续臭别人哦！扣 10 分警告一下",
		"cmd.zc.cooldown":    "😳 用指令对线是不对的，请大家都冷静下呢～",
		"cmd.zc.exec":        "👮 %s, 您被热心的 %s 警告了 ⚠️，请注意管理好自己的行为！暂时扣除 25 分作为警告，如果您的分数低于 -50 分将被直接禁言。若您觉得这是恶意举报，请理性对待，并联系群管理员处理。",
		"cmd.zc.noAnonymous": "😠 匿名就不要乱啵啵啦！叭了个叭叭了个叭叭了个叭 ...",

		"cmd.ey.selfReport":  "👮 举报自己？那没办法...只好把你 🫒 半小时哦～",
		"cmd.ey.notSuccess":  "😭 呜呜呜，封不掉 ～",
		"cmd.ey.unexpected":  "😠 叭了个叭叭了个叭叭了个叭 ～",
		"cmd.ey.killChannel": "👮 好的！这就把这个频道封掉啦～ PS: %s 的主人，如果您觉得这是恶意举报，请赶快联系管理员解封哦 ～）",
		"cmd.ey.killBot":     "👮 好的！这就把这个机器人封禁半小时～ PS: %s 的主人，如果您觉得这是恶意举报，请赶快联系管理员解封哦 ～）",
		"cmd.ey.cooldown5":   "😠 消停一下消停一下，举报太多次啦，扣 5 分缓一缓",
		"cmd.ey.exec":        "👮 %s, 您被热心群友 %s 报告有发送恶意言论的嫌疑 ⚠️，请注意自己的发言哦！暂时禁言半小时并扣除 50 分作为警告，举报者 15 分奖励已到账。若您觉得这是恶意举报，可以呼吁小伙伴们公投为您解封（累计满 6 票可以解封并抵消扣分），或者直接联系群管理员处理。",
		"cmd.ey.duplicated":  "👮 他已经被检察官带走啦，不要鞭尸啦 ～",

		"cmd.privateChatFirst":        "❌ 请先私聊我然后再运行这个命令哦",
		"cmd.noPerm":                  "❌ 您没有使用这个命令的权限呢",
		"cmd.mustReply":               "❌ 请在群组内回复一个有效用户使用这个命令哦 ～",
		"cmd.noGroupPerm":             "❌ 您没有权限，亦或是您未再对应群组使用这个命令",
		"cmd.noMiaoPerm":              "❌ 您没有喵组权限，亦或是您未再对应群组使用这个命令",
		"cmd.mustReplyChannelOrInput": "❌ 请回复一则转发的频道消息或者手动加上频道 id ～",
		"cmd.mustInGroup":             "❌ 请在群组发送这条命令哦 ～",

		"cmd.misc.version": "👀 当前版本为: %s",
		"cmd.misc.ping.1":  "🔗 与 Telegram 伺服器的延迟约为:\n\n机器人 DC: `%dms`",
		"cmd.misc.ping.2":  "🔗 与 Telegram 伺服器的延迟约为:\n\n机器人 DC: `%dms`\n群组 DC: `%dms`",

		"grant.assign.success":  "✔️ TA 已经成为管理员啦 ～",
		"grant.assign.failure":  "❌ TA 已经是管理员啦 ～",
		"grant.dismiss.success": "✔️ 已将 TA 的管理员移除 ～",
		"grant.dismiss.failure": "❌ TA 本来就不是管理员呢 ～",

		"forward.ban.success":   "✔️ TA 已经被我封掉啦 ～",
		"forward.ban.failure":   "❌ TA 已经被封禁过啦 ～",
		"forward.unban.success": "✔️ TA 已经被我解封啦 ～",
		"forward.unban.failure": "❌ TA 还没有被封禁哦 ～",

		"credit.set.invalid":      "❌ 使用方法错误：/setcredit <UserId:Optional> <Credit>",
		"credit.add.invalid":      "❌ 使用方法错误：/addcredit <UserId:Optional> <Credit>",
		"credit.set.success":      "✔️ 设置成功，TA 的积分为: %d",
		"credit.importSuccess":    "✔️ 导入 %d 条成功，您可以输入 /creditrank 查看导入后积分详情",
		"credit.exportSuccess":    "✔️ 导出成功，请在私聊查看结果",
		"credit.importError":      "❌ 无法下载积分备份，请确定您上传的文件格式正确且小于 20MB，大文件请联系管理员手动导入",
		"credit.importParseError": "❌ 解析积分备份错误，请确定您上传的文件格式正确",
		"credit.check.success":    "👀 `%s`, TA 当前的积分为: %d",
		"credit.check.my":         "👀 `%s`, 您当前的积分为: %d",
		"credit.rank.info":        "#开榜 当前的积分墙为: \n\n",
		"credit.lottery.info":     "🎉 恭喜以下用户中奖：\n\n",

		"spoiler.invalid": "❌ 使用方法错误：/set_antispoiler <on|off>",
		"spoiler.success": "✔️ 已经设置好反·反剧透消息啦 `(Status=%v)` ～",

		"su.group.addSuccess":   "✔️ 已将该组加入积分统计 ～",
		"su.group.addDuplicate": "❌ 该组已经开启积分统计啦 ～",
		"su.group.delSuccess":   "✔️ 已将该组移除积分统计 ～",
		"su.group.delDuplicate": "❌ 该组尚未开启积分统计哦 ～",

		// not support yet
		"rp.complete":  "🧧 *积分红包*\n\n小伙伴们手速都太快啦，`%s`的大红包已被瓜分干净，没抢到的小伙伴们请期待下次的活动哦～",
		"rp.guessLeft": "猜猜看还剩多少？",
		"rp.text":      "🧧 *积分红包*\n\n``%s发红包啦！大家快抢哦～\n\n剩余积分: `%s`\n剩余数量: `%d`",
		"rp.lucky":     "\n\n恭喜手气王 `%s` 获得了 `%d` 分 🎉 ~",

		"rp.admin":              "管理员-",
		"rp.set.invalid":        "❌ 使用方法不正确呢，请输入 /redpacket `<总分数>` `<红包个数>` 来发红包哦～\n\n备注：红包总分需在 1 ~ 1000 之间，红包个数需在 1 ~ 20 之间，且红包大小不能低于参与人数哦～",
		"rp.set.noEnoughCredit": "❌ 您的积分不够发这个红包哦，请在努力赚积分吧～",

		"gp.ban.success":   "🎉 恭喜 `%s` 获得禁言大礼包，可喜可贺可喜可贺！",
		"gp.ban.failure":   "❌ 您没有办法禁言 TA 呢",
		"gp.unban.success": "🎉 恭喜 `%s` 重新获得了自由 ～",
		"gp.unban.failure": "❌ 您没有办法解禁 TA 呢",
		"gp.kick.success":  "🎉 恭喜 `%s` 被踢出去啦！",
		"gp.kick.failure":  "❌ 您没有踢掉 TA 呢",

		"channel.set.cancel":           "✔️ 已经取消加群频道验证啦 ～",
		"channel.set.success":          "✔️ 已经设置好加群频道验证啦 `(Join=%v, Msg=%v)` ～",
		"channel.bot.permit":           "👏 欢迎 %s 加入群组，已为机器人自动放行 ～",
		"channel.user.alreadyFollowed": "👏 欢迎 %s 加入群组，您已关注频道自动放行 ～",
		"channel.request":              "[🎉](tg://user?id=%d) 欢迎 `%s`，您还没有关注本群组关联的频道哦，您有 5 分钟时间验证自己 ～ 请点击下面按钮跳转到频道关注后再回来验证以解除发言限制 ～",
		"channel.cannotSendMsg":        "❌ 无法发送验证消息，请管理员检查群组权限 ～",
		"channel.cannotBanUser":        "❌ 无法完成验证流程，请管理员检查机器人封禁权限 ～",
		"channel.cannotCheckChannel":   "❌ 无法检测用户是否在目标频道内，请管理员检查机器人权限 ～",
		"channel.kicked":               "👀 [TA](tg://user?id=%d) 没有在规定时间内完成验证，已经被我带走啦 ～",

		"locale.set": "✔️ 设置成功，当前群组的默认语言为: %s ～",
		"locale.get": "👀 当前群组的默认语言为: %s ～",

		// not support yet
		"btn.rp.draw": "🤏 我要抢红包|rp/%d/1/%d",
		"btn.notFair": "😠 这不公平 (%d)|vt/%d/%d/%d",

		"btn.adminPanel":    "🚩 解封[管理]|unban/%d/%d/%d||🚮 清退[管理]|kick/%d/%d/%d",
		"btn.channel.step1": "👉 第一步：关注频道 👈|https://t.me/%s",
		"btn.channel.step2": "👉 第二步：点我验证 👈|check/%d/%d",

		"cb.unblock.byadmin": "\n\nTA 已被管理员解封 👊",
		"cb.kicked.byadmin":  "\n\nTA 已被管理员踢出群聊 🦶",
		"cb.unblock.byvote":  "\n\n于多名用户投票后决定，该用户不是恶意广告，用户已解封，积分已原路返回。",

		"cb.unban.success":                "✔️ 已解除封禁，请您手动处理后续事宜 ~",
		"cb.unban.failure":                "❌ 解封失败，TA 可能已经被解封或者已经退群啦 ~",
		"cb.kick.success":                 "✔️ 已将 TA 送出群留学去啦 ~",
		"cb.kick.failure":                 "❌ 踢出失败，可能 TA 已经退群啦 ~",
		"cb.validate.success":             "✔️ 验证成功，欢迎您的加入 ~",
		"cb.validate.success.cannotUnban": "❌ 验证成功，但是解禁失败，请联系管理员处理 ~",
		"cb.validate.failure":             "❌ 验证失败，请确认自己已经加入对应群组 ~",
		"cb.validate.others":              "😠 人家的验证不要乱点哦！！！",
		"cb.vote.success":                 "✔️ 投票成功，感谢您的参与 ~",
		"cb.vote.failure":                 "❌ 您已经参与过投票了，请不要多次投票哦 ~",
		"cb.vote.notExists":               "❌ 投票时间已过，请联系管理员处理 ~",
		"cb.rp.nothing":                   "🐢 您的运气也太差啦！什么都没有抽到哦...",
		"cb.rp.get.1":                     "🎉 恭喜获得 ",
		"cb.rp.get.2":                     " 积分，积分已经实时到账～",
		"cb.rp.duplicated":                "❌ 您已经参与过这次活动了，不能太贪心哦！",
		"cb.rp.notExists":                 "❌ 抽奖活动已经结束啦！请期待下一次活动～",
		"cb.lottery.start":                "🎉 活动已确认，请号召群友踊跃参与哦！",
		"cb.lottery.enroll":               "🎉 参与成功 ~ 请耐心等待开奖呀 ~",
		"cb.lottery.noEnoughCredit":       "❌ 你的积分不满足活动要求哦！",
		"cb.lottery.checkFailed":          "❌ 请加群后再参与活动哦！",
		"cb.notMiaoAdmin":                 "❌ 请不要乱玩喵组管理员指令！",
		"cb.notAdmin":                     "❌ 请不要乱玩管理员指令！",
		"cb.noEvent":                      "❌ 未找到这个活动，请联系管理员解决！",
		"cb.notParsed":                    "❌ 指令解析出错，请联系管理员解决 ~",
		"cb.disabled":                     "❌ 这个群组还没有被授权哦 ~",
	},
	"en": {
		"system.unexpected": "❌ cannot fulfill the task, please check logs",

		"cmd.zc.notAllowed":  "嘴臭 is not permitted in this group",
		"cmd.zc.indeed":      "INDEED",
		"cmd.zc.cantBan":     "Well, I have nothing to do with it ...",
		"cmd.zc.cooldown10":  "😠 DO NOT TALK LIKE SHIT, YOU WILL BE PUNISHED BY 10 POINTS",
		"cmd.zc.cooldown":    "😳 Calm down, calm down ...",
		"cmd.zc.exec":        "👮 %s, you are warned by %s ⚠️, please do not be too aggressive! You are punished by 25 credit points. If your credit is below -50, you would be restricted in this group. Please contact group admin if you think the judgement is a mistake.",
		"cmd.zc.noAnonymous": "😠 PLEASE WELL BEHAVE WHEN YOU ARE ANONYMOUS ...",

		"cmd.ey.selfReport":  "👮 Yeah, you know what you are doing. You are restricted for half an hour.",
		"cmd.ey.notSuccess":  "😭 Wwww, I cannot do that.",
		"cmd.ey.unexpected":  "😠 Ba Le Ge Ba, Ba Le Ge Ba Ba Ba Ba ～",
		"cmd.ey.killChannel": "👮 This channel has been banned, PS: if the owner of %s finds it a mistake, please contact the group admin asap.",
		"cmd.ey.killBot":     "👮 This bot has been restricted for half an hour, PS: if the owner of %s finds it a mistake, please contact the group admin asap.",
		"cmd.ey.cooldown5":   "😠 DO NOT TALK LIKE SHIT, YOU WILL BE PUNISHED BY 5 POINTS",
		"cmd.ey.exec":        "👮 %s, you are reported by %s to shot spam into the group ⚠️, please well behave! You are punished by 50 credit points and the reporter has gained 25 points. Please contact group admin if you think the judgement is a mistake, or you could ask for other members to vote to help.",
		"cmd.ey.duplicated":  "👮 The user has already been banned.",

		"cmd.privateChatFirst":        "❌ Please start me in the private chat before using this command.",
		"cmd.noPerm":                  "❌ You are not permitted to use this command.",
		"cmd.mustReply":               "❌ Please reply this command to a valid user is a valid group.",
		"cmd.noGroupPerm":             "❌ You are not authorized to use this command, or the group is not set up yet by admin.",
		"cmd.noMiaoPerm":              "❌ You are not authorized to use this miao-perm command, or the group is not set up yet by admin.",
		"cmd.mustReplyChannelOrInput": "❌ Please reply this command to a forwarded channel message, or pass in the channel id as a parameter.",
		"cmd.mustInGroup":             "❌ Please send this command in a group chat.",

		"cmd.misc.version": "👀 Current Version: %s",
		"cmd.misc.ping.1":  "🔗 Telegram Server Transmission Delay:\n\nBot DC: `%dms`",
		"cmd.misc.ping.2":  "🔗 Telegram Server Transmission Delay:\n\nBot DC: `%dms`\nGroup DC: `%dms`",

		"grant.assign.success":  "✔️ The user is promoted ～",
		"grant.assign.failure":  "❌ The user does not need to be promoted ～",
		"grant.dismiss.success": "✔️ The user is dismissed ～",
		"grant.dismiss.failure": "❌ The user does not need to be dismissed ～",

		"forward.ban.success":   "✔️ The user has been banned ～",
		"forward.ban.failure":   "❌ The user was banned ～",
		"forward.unban.success": "✔️ The user has been released ～",
		"forward.unban.failure": "❌ The user does not need to be released ～",

		"credit.set.invalid":      "❌ Invalid Params. Please refer to: /setcredit <UserId:Optional> <Credit>",
		"credit.add.invalid":      "❌ Invalid Params. Please refer to: /addcredit <UserId:Optional> <Credit>",
		"credit.set.success":      "✔️ Success, the credit point of the user is: %d.",
		"credit.importSuccess":    "✔️ Imported %d columns，you can check credit details with /creditrank.",
		"credit.exportSuccess":    "✔️ Exported, please check the result in the private chat.",
		"credit.importError":      "❌ Unable to fetch the file, please make sure the file is valid and less than 20MB.",
		"credit.importParseError": "❌ Unable to decode the file, please try again.",
		"credit.check.success":    "👀 `%s` has %d credit points",
		"credit.check.my":         "👀 `%s`, you have %d credit points",
		"credit.rank.info":        "#RANK The credit rank of the group: \n\n",
		"credit.lottery.info":     "🎉 Congrats to the following users:\n\n",

		"spoiler.invalid": "❌ Invalid Params. Please refer to: /set_antispoiler <on|off>",
		"spoiler.success": "✔️ Anti-spoiler settings has been updated `(Status=%v)` ～",

		"su.group.addSuccess":   "✔️ The group is enrolled successfully ～",
		"su.group.addDuplicate": "❌ The group has already been enrolled ～",
		"su.group.delSuccess":   "✔️ The group is quitted successfully ～",
		"su.group.delDuplicate": "❌ The group has not been enrolled yet ～",

		// not support yet
		// "rp.complete":  "🧧 *积分红包*\n\n小伙伴们手速都太快啦，`%s`的大红包已被瓜分干净，没抢到的小伙伴们请期待下次的活动哦～",
		// "rp.guessLeft": "猜猜看还剩多少？",
		// "rp.text":      "🧧 *积分红包*\n\n``%s发红包啦！大家快抢哦～\n\n剩余积分: `%s`\n剩余数量: `%d`",
		// "rp.lucky":     "\n\n恭喜手气王 `%s` 获得了 `%d` 分 🎉 ~",

		"rp.admin":              "Admin ",
		"rp.set.invalid":        "❌ Invalid Params. Please refer to: /redpacket `<Total Credit>` `<Num of Share>`\n\nPS: Total Credit should be with in 1 and 1000. Number of Share should be with in 1 and 20 and no less than the Total Credit.",
		"rp.set.noEnoughCredit": "❌ You do not have that much credit to send this redpacket.",

		"gp.ban.success":   "🎉 Congrats to `%s` to be restricted!",
		"gp.ban.failure":   "❌ You cannot restrict the user.",
		"gp.unban.success": "🎉 Congrats to `%s` to be released!",
		"gp.unban.failure": "❌ You cannot release the user.",
		"gp.kick.success":  "🎉 Congrats to `%s` to be kicked!",
		"gp.kick.failure":  "❌ You cannot kick the user.",

		"channel.set.cancel":           "✔️ Group MFC has been turned off ～",
		"channel.set.success":          "✔️ Group MFC has been turned on `(Join=%v, Msg=%v)` ～",
		"channel.bot.permit":           "👏 Welcome %s, bots are permitted to join by default ～",
		"channel.user.alreadyFollowed": "👏 Welcome %s, you already followed the linked channel, you are all set ～",
		"channel.request":              "[🎉](tg://user?id=%d) Welcome `%s`, you have not yet followed the linked channel of the group for multi-factor CAPTCHA purpose. Please join the channel within 5 minutes to prove you are not a robot ～",
		"channel.cannotSendMsg":        "❌ Cannot send the verification message, please check my permission ～",
		"channel.cannotBanUser":        "❌ Cannot complete the CAPTCHA, please check my permission ～",
		"channel.cannotCheckChannel":   "❌ Cannot read the user list of targetted channel, please make sure the bot has enough permission in the channel ～",
		"channel.kicked":               "👀 [The user](tg://user?id=%d) did not pass the MFC verification, so it is banned ～",

		"locale.set": "✔️ The default language of this group has been changed to: %s ～",
		"locale.get": "👀 The default language of this group is: %s ～",

		// not support yet
		// "btn.rp.draw": "🤏 我要抢红包|rp/%d/1/%d",
		// "btn.notFair": "😠 这不公平 (%d)|vt/%d/%d/%d",

		"btn.adminPanel":    "🚩 UNBAN [ADMIN]|unban/%d/%d/%d||🚮 KICK [ADMIN]|kick/%d/%d/%d",
		"btn.channel.step1": "👉 1ST: JOIN THE CHANNEL 👈|https://t.me/%s",
		"btn.channel.step2": "👉 2ND: RELEASE ME 👈|check/%d/%d",

		"cb.unblock.byadmin": "\n\nThe user is unbanned by admin 👊",
		"cb.kicked.byadmin":  "\n\nThe user has been kicked 🦶",
		"cb.unblock.byvote":  "\n\nThe user is voted to be innocent, the credit punishment is reverted.",

		"cb.unban.success":                "✔️ User is unbanned ~",
		"cb.unban.failure":                "❌ User cannot be unbanned ~",
		"cb.kick.success":                 "✔️ User is kicked ~",
		"cb.kick.failure":                 "❌ User cannot be kicked ~",
		"cb.validate.success":             "✔️ Verified. Welcome to the group ~",
		"cb.validate.success.cannotUnban": "❌ Verified, yet I cannot unban you. Please contact group admin for help ~",
		"cb.validate.failure":             "❌ Fail to verify. Please make sure you have followed the channel ~",
		"cb.validate.others":              "😠 DO NOT PLAY CAPTCHA OF OTHERS",
		"cb.vote.success":                 "✔️ Voted. Thanks for your participation ~",
		"cb.vote.failure":                 "❌ Duplicated vote ~",
		"cb.vote.notExists":               "❌ The vote has been closed ~",
		"cb.rp.nothing":                   "🐢 AHA you get nothing...",
		"cb.rp.get.1":                     "🎉 You get ",
		"cb.rp.get.2":                     " credit points. Congrats ～",
		"cb.rp.duplicated":                "❌ Duplicated draw, DONT BE VORACIOUS ~",
		"cb.rp.notExists":                 "❌ The event is over, please engage next time ~",
		"cb.lottery.start":                "🎉 The lottery is submitted.",
		"cb.lottery.enroll":               "🎉 Enrolled. Thanks for your participation ~",
		"cb.lottery.noEnoughCredit":       "❌ Your credit points are not enough to enroll in the lottery.",
		"cb.lottery.checkFailed":          "❌ Please make sure you are in the group.",
		"cb.notMiaoAdmin":                 "❌ Do not play with the button!",
		"cb.notAdmin":                     "❌ Do not play with the button!",
		"cb.noEvent":                      "❌ The event is not found.",
		"cb.notParsed":                    "❌ The event is invalid.",
		"cb.disabled":                     "❌ The group is not authorized.",
	},
}

const DEFAULT_LANG = "en"

func Locale(identifier string, locale string) string {
	// process alias
	if alias, ok := LocaleAlias[locale]; ok && alias != "" {
		locale = alias
	}

	// find keywords
	if locales, ok := LocaleMap[locale]; ok && locales != nil {
		if text, ok := locales[identifier]; ok && text != "" {
			return text
		}
	}

	// fallback
	if text, ok := LocaleMap[DEFAULT_LANG][identifier]; ok && text != "" {
		return text
	}

	return identifier
}

func GetUserLocale(c *tb.Chat, u *tb.User) string {
	if u != nil && u.LanguageCode != "" {
		return u.LanguageCode
	}

	if c != nil {
		gc := GetGroupConfig(c.ID)
		if gc.Locale != "" {
			return gc.Locale
		}
	}

	return DEFAULT_LANG
}

func GetSenderLocale(m *tb.Message) string {
	user := m.Sender
	if m.UserJoined != nil {
		user = m.UserJoined
	}

	return GetUserLocale(m.Chat, user)
}

func GetSenderLocaleCallback(c *tb.Callback) string {
	return GetUserLocale(c.Message.Chat, c.Sender)
}
