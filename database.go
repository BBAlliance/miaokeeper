package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BBAlliance/miaokeeper/memutils"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/telebot.v3"
)

var DBCONN = ""
var DBPREFIX = "MiaoKeeper"
var DB *gorm.DB
var UpdateLock sync.Mutex
var GroupConfigLock sync.Mutex
var LotteryConfigLock sync.Mutex

type UpdateMethod int

const (
	UMSet UpdateMethod = iota
	UMAdd
	UMDel
)

type OPReasons string

const (
	OPAll          OPReasons = ""
	OPFlush        OPReasons = "FLUSH"
	OPNormal       OPReasons = "NORMAL"
	OPByAdmin      OPReasons = "ADMIN"
	OPByAdminSet   OPReasons = "ADMINSET"
	OPByRedPacket  OPReasons = "REDPACKET"
	OPByLottery    OPReasons = "LOTTERY"
	OPByTransfer   OPReasons = "TRANSFER"
	OPByPolicy     OPReasons = "POLICY"
	OPByAbuse      OPReasons = "ABUSE"
	OPByAPIConsume OPReasons = "CONSUME"
	OPByAPIAdd     OPReasons = "ADD"
	OPByCleanUp    OPReasons = "CLEANUP"
)

var OPAllReasons = []OPReasons{OPAll, OPFlush, OPNormal, OPByAdmin, OPByAdminSet, OPByRedPacket, OPByLottery, OPByTransfer, OPByPolicy, OPByAbuse, OPByAPIConsume, OPByCleanUp}

func (op *OPReasons) Repr() string {
	if *op == OPAll {
		return "ALL"
	} else {
		return string(*op)
	}
}

func OPParse(s string) OPReasons {
	for _, op := range OPAllReasons {
		if string(op) == s {
			return op
		}
	}

	return OPAll
}

func DBTName(tableName string, extras ...int64) string {
	table := DBPREFIX + "_" + tableName
	if len(extras) > 0 {
		table += "_" + fmt.Sprintf("%d", Abs(extras[0]))
	}
	return table
}

func GetRandClause() string {
	if strings.HasPrefix(DBCONN, "postgres") {
		return "RANDOM()"
	}
	return "RAND()"
}

// GORM:%NAME%_Credit_%GROUP%
type CreditInfo struct {
	ID       int64  `json:"id" gorm:"column:userid;primaryKey;not null"`
	Username string `json:"username" gorm:"column:username;type:text;not null"`
	Name     string `json:"nickname" gorm:"column:name;type:text;not null"`
	Credit   int64  `json:"credit" gorm:"column:credit;not null"`
	GroupId  int64  `json:"groupId" gorm:"-"`
}

// GORM:%NAME%_Credit_Log_%GROUP%
type CreditLog struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement;not null"`
	UserID    int64     `gorm:"column:userid;not null;index"`
	Credit    int64     `gorm:"column:credit;not null"`
	Reason    OPReasons `gorm:"column:op;type:string;size:16;not null;index"`
	CreatedAt time.Time `gorm:"column:createdat;autoCreateTime"`
}

// GORM:%NAME%_Config
type DBGlobalConfig struct {
	K string `gorm:"column:k;type:string;primaryKey;size:128;not null"`
	V string `gorm:"column:v;type:text;not null"`
}

// GORM:%NAME%_Lottery
type DBLottery struct {
	ID        string    `gorm:"column:id;type:string;size:128;primaryKey;not null"`
	Config    string    `gorm:"column:config;type:text;not null"`
	CreatedAt time.Time `gorm:"column:createdat;autoCreateTime"`
}

// GORM:%NAME%_Lottery_Participation
type DBLotteryParticipation struct {
	LotteryID   string    `gorm:"column:lotteryid;type:string;size:128;uniqueIndex:uniq_participant;index:lottery_id;not null"`
	Participant int64     `gorm:"column:participant;uniqueIndex:uniq_participant;not null"`
	Username    string    `gorm:"column:username;type:text;not null"`
	CreatedAt   time.Time `gorm:"column:createdat;autoCreateTime"`
}

var GroupConfigCache map[int64]*GroupConfig
var LotteryConfigCache map[string]*LotteryInstance

func InitDatabase() (err error) {
	var conn gorm.Dialector
	if strings.HasPrefix(DBCONN, "postgres") {
		// postgresSQL
		conn = postgres.Open(DBCONN)
	} else {
		// mysql
		connStr := DBCONN
		if strings.Contains(connStr, "?") {
			connStr += "&parseTime=true"
		} else {
			connStr += "?parseTime=true"
		}
		conn = mysql.Open(connStr)
	}

	DB, err = gorm.Open(conn, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	return
}

func ReadConfig(key string) string {
	dbgc := DBGlobalConfig{}
	err := DB.Table(DBTName("Config")).First(&dbgc, "k = ?", key).Error
	if err != nil {
		DLogf("Config Read Error | key=%s error=%v", key, err.Error())
	}
	return dbgc.V
}

func WriteConfig(key, value string) {
	dbgc := DBGlobalConfig{key, value}
	err := DB.Table(DBTName("Config")).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "k"}},
		DoUpdates: clause.AssignmentColumns([]string{"v"}),
	}).Create(&dbgc).Error
	if err != nil {
		DLogf("Config Write Error | key=%s value=%s error=%v", key, value, err.Error())
	}
}

func GetAliasedGroup(groupId int64) int64 {
	if gc := GetGroupConfig(groupId); gc != nil {
		if gc.MergeTo < 0 {
			return gc.MergeTo
		}
	}
	return groupId
}

func GetGroupConfig(groupId int64) *GroupConfig {
	if !IsGroup(groupId) {
		return nil
	}

	GroupConfigLock.Lock()
	defer GroupConfigLock.Unlock()
	if gc, ok := GroupConfigCache[groupId]; ok {
		return gc
	}

	cfg := ReadConfig(fmt.Sprintf("gcfg_%d", Abs(groupId)))
	if cfg != "" {
		gc := &GroupConfig{}
		err := jsoniter.Unmarshal([]byte(cfg), gc)
		if err == nil {
			GroupConfigCache[groupId] = gc.Check()
			return gc
		}
	}
	return nil
}

func SetGroupConfig(groupId int64, gc *GroupConfig) *GroupConfig {
	if !IsGroup(groupId) {
		return nil
	}

	if groupId >= 0 {
		return nil
	}

	GroupConfigLock.Lock()
	defer GroupConfigLock.Unlock()
	GroupConfigCache[groupId] = gc

	cfg, _ := jsoniter.Marshal(gc)
	if len(cfg) > 0 {
		WriteConfig(fmt.Sprintf("gcfg_%d", Abs(groupId)), string(cfg))
	}
	return gc
}

func InitGroupTable(groupId int64) {
	PlainError("Unable to create table", DB.Table(DBTName("Credit", groupId)).AutoMigrate(&CreditInfo{}))
	PlainError("Unable to create table", DB.Table(DBTName("Credit_Log", groupId)).AutoMigrate(&CreditLog{}))

	if GetGroupConfig(groupId) == nil {
		NewGroupConfig(groupId)
	}
}

func ReadConfigs() {
	PlainError("Unable to create config table", DB.Table(DBTName("Config")).AutoMigrate(&DBGlobalConfig{}))
	PlainError("Unable to create lottery table", DB.Table(DBTName("Lottery")).AutoMigrate(&DBLottery{}))
	PlainError("Unable to create table", DB.Table(DBTName("Lottery_Participation")).AutoMigrate(&DBLotteryParticipation{}))

	ADMINS = ParseStrToInt64Arr(ReadConfig("ADMINS"))
	GROUPS = ParseStrToInt64Arr(ReadConfig("GROUPS"))

	for _, g := range GROUPS {
		InitGroupTable(g)
	}
}

func WriteConfigs() {
	WriteConfig("ADMINS", ParseInt64ArrToStr(ADMINS))
	WriteConfig("GROUPS", ParseInt64ArrToStr(GROUPS))
}

func UpdateAdmin(userId int64, method UpdateMethod) bool {
	UpdateLock.Lock()
	defer UpdateLock.Unlock()
	changed := false
	if method == UMSet {
		if len(ADMINS) != 1 || ADMINS[0] != userId {
			changed = true
			ADMINS = []int64{userId}
		}
	} else if method == UMAdd {
		ADMINS, changed = AddIntoInt64Arr(ADMINS, userId)
	} else if method == UMDel {
		ADMINS, changed = DelFromInt64Arr(ADMINS, userId)
	}
	WriteConfigs()
	return changed
}

func UpdateGroup(groupId int64, method UpdateMethod) bool {
	UpdateLock.Lock()
	defer UpdateLock.Unlock()
	changed := false
	if method == UMSet {
		if len(GROUPS) != 1 || GROUPS[0] != groupId {
			changed = true
			GROUPS = []int64{groupId}
		}
	} else if method == UMAdd {
		GROUPS, changed = AddIntoInt64Arr(GROUPS, groupId)
	} else if method == UMDel {
		GROUPS, changed = DelFromInt64Arr(GROUPS, groupId)
	}
	if changed && method == UMSet || method == UMAdd {
		InitGroupTable(groupId)
	}
	WriteConfigs()
	return changed
}

func GetCredit(groupId, userId int64) *CreditInfo {
	ret := &CreditInfo{}
	realGroup := GetAliasedGroup(groupId)
	err := DB.Table(DBTName("Credit", realGroup)).First(&ret, "userid = ?", userId).Error
	if err != nil {
		DLogf("Database Credit Read Error | gid=%d rgid=%d uid=%d error=%s", groupId, realGroup, userId, err.Error())
	}
	if ret.ID == userId {
		ret.GroupId = groupId
	}
	return ret
}

func GetCreditRank(groupId int64, limit int) []*CreditInfo {
	returns := []*CreditInfo{}
	realGroup := GetAliasedGroup(groupId)
	DB.Table(DBTName("Credit", realGroup)).Order("credit DESC").Limit(limit).Find(&returns)
	for _, ci := range returns {
		if ci != nil && ci.ID > 0 {
			ci.GroupId = groupId
		}
	}
	return returns
}

// does not apply MergeTo
func DumpCredits(groupId int64) [][]string {
	ret := [][]string{}
	batches := []CreditInfo{}
	DB.Table(DBTName("Credit", groupId)).FindInBatches(&batches, 100, func(tx *gorm.DB, batchNum int) error {
		for _, batch := range batches {
			if batch.ID > 0 && batch.Credit > 0 {
				ret = append(ret, []string{strconv.FormatInt(batch.ID, 10), batch.Name, batch.Username, strconv.FormatInt(batch.Credit, 10)})
			}
		}
		return nil
	})

	DInfof("Credit Dump | group=%d columns=%d", groupId, len(ret))
	return ret
}

// does not apply MergeTo
func FlushCredits(groupId int64, records [][]string) {
	if len(records) == 0 {
		return
	}

	batches := []CreditInfo{}
	logbatches := []CreditLog{}
	for _, r := range records {
		if len(r) >= 4 {
			batches = append(batches, CreditInfo{
				ID:       ParseInt64(r[0]),
				Name:     r[1],
				Username: r[2],
				Credit:   ParseInt64(r[3]),
			})
			logbatches = append(logbatches, CreditLog{
				UserID: ParseInt64(r[0]),
				Credit: ParseInt64(r[3]),
				Reason: OPFlush,
			})
		}
	}
	err := DB.Table(DBTName("Credit", groupId)).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "userid"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "username", "credit"}),
	}).CreateInBatches(&batches, 100).Error
	if err != nil {
		DErrorE(err, "Database Credit Flush Error")
	}

	// writing logs
	err = DB.Table(DBTName("Credit_Log", groupId)).CreateInBatches(&logbatches, 100).Error
	if err != nil {
		DErrorE(err, "Database Credit Log Flush Error")
	}

	DInfof("Flush Credit | group=%d columns=%d", groupId, len(records))
}

func QueryLogs(groupId int64, offset uint64, limit uint64, uid int64, before time.Time, vtype OPReasons) []CreditLog {
	var ret = []CreditLog{}

	tx := DB.Table(DBTName("Credit_Log", groupId)).Order("id DESC").Limit(int(limit)).Offset(int(offset))
	if uid > 0 {
		tx.Where("userid = ?", uid)
	}
	if vtype != "" {
		tx.Where("op = ?", string(vtype))
	}
	tx.Find(&ret)

	DInfof("Query Logs | group=%d offset=%d limit=%d userId=%d before=%d reason=%s columns=%d", groupId, offset, limit, uid, before.Unix(), vtype, len(ret))
	return ret
}

func UpdateCredit(user *CreditInfo, method UpdateMethod, value int64, reason OPReasons) *CreditInfo {
	ci := GetCredit(user.GroupId, user.ID)
	if user.Name == "" {
		user.Name = ci.Name
	}
	if user.Username == "" {
		user.Username = ci.Username
	}
	user.Credit = ci.Credit
	if method == UMSet {
		user.Credit = value
	} else if method == UMAdd {
		user.Credit += value
	} else if method == UMDel {
		user.Credit = 0
	}

	var err error
	realGroup := GetAliasedGroup(user.GroupId)
	if method != UMDel {
		err = DB.Table(DBTName("Credit", realGroup)).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "userid"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "username", "credit"}),
		}).Create(&user).Error
		DB.Table(DBTName("Credit_Log", realGroup)).Create(&CreditLog{
			UserID: user.ID,
			Credit: value,
			Reason: reason,
		})
	} else if realGroup == user.GroupId {
		// when the method is UMDel, do not delete aliased credit
		err = DB.Table(DBTName("Credit", realGroup)).Delete(&user).Error
		DB.Table(DBTName("Credit_Log", realGroup)).Create(&CreditLog{
			UserID: user.ID,
			Credit: -ci.Credit,
			Reason: reason,
		})
	}
	if err != nil {
		DErrorE(err, "Database Credit Update Error")
	}

	DLogf("Update Credit | gid=%d rgid=%d user=%d alter=%d credit=%d", user.GroupId, realGroup, user.ID, method, value)
	return user
}

// status: -1 not start, 0 start, 1 stopped, 2 finished
type LotteryInstance struct {
	ID        string
	Status    int
	GroupID   int64
	MsgID     int
	CreatedAt int64
	StartedAt int64

	Payload     string
	Limit       int
	Consume     bool
	Num         int
	Duration    time.Duration
	Participant int

	Winners          []DBLotteryParticipation
	ParticipantCache int        `json:"-"`
	JoinLock         sync.Mutex `json:"-"`
}

func (li *LotteryInstance) UpdateTelegramMsg() *tb.Message {
	btns := []string{}
	if li.Status == 0 {
		btns = append(btns, fmt.Sprintf("🤏 我要抽奖|lt?t=1&id=%s", li.ID))
	}
	if li.Status >= 0 && li.Status < 2 {
		btns = append(btns, fmt.Sprintf("📦 手动开奖[管理]|lt?t=3&id=%s", li.ID))
	}
	if li.Status == -1 {
		btns = append(btns, fmt.Sprintf("🎡 开启活动[管理]|lt?t=2&id=%s", li.ID))
	}
	if li.MsgID > 0 && li.Status == 2 {
		Bot.Delete(&tb.Message{ID: li.MsgID, Chat: &tb.Chat{ID: li.GroupID}})
		li.MsgID = 0
	}
	if li.MsgID <= 0 {
		msg, _ := SendBtnsMarkdown(&tb.Chat{ID: li.GroupID}, li.GenText(), "", btns)
		if msg != nil {
			li.MsgID = msg.ID
			li.Update()
		}
		return msg
	} else {
		msg, _ := EditBtnsMarkdown(&tb.Message{ID: li.MsgID, Chat: &tb.Chat{ID: li.GroupID}}, li.GenText(), "", btns)
		if msg == nil {
			li.MsgID = 0
			return li.UpdateTelegramMsg()
		}
		return msg
	}
}

func (li *LotteryInstance) GenText() string {
	drawMsg := ""
	if li.Participant > 0 {
		drawMsg = fmt.Sprintf("参与人数达 %d 人", li.Participant)
	}
	if li.Duration > 0 {
		if drawMsg != "" {
			drawMsg += " 或 "
		}
		durationStr := ""
		if li.Duration >= time.Hour {
			durationStr = fmt.Sprintf("%.1f 小时", li.Duration.Hours())
		} else {
			durationStr = fmt.Sprintf("%d 分钟", int(li.Duration.Minutes()))
		}
		drawMsg += fmt.Sprintf("%s后自动开奖", durationStr)
	}
	if drawMsg == "" {
		drawMsg = "手动开奖"
	}

	status := "`未知`"
	if li.Status == -1 {
		status = "`待验证`"
	} else if li.Status == 0 {
		status = "`进行中`"
	} else if li.Status == 1 {
		status = "`待手动开奖`"
	} else if li.Status == 2 {
		status = "`已开奖`"
	}
	if li.Status >= 0 {
		status += fmt.Sprintf("\n*参与人数:* %d", li.Participants())
	}
	if len(li.Winners) > 0 && len(li.Winners) <= len(li.Winners) {
		status += "\n\n*🏆 获奖者:*"
		for i := range li.Winners {
			status += fmt.Sprintf("\n`%2d.` `%s` ([%d](%s))", i+1, GetQuotableStr(li.Winners[i].Username), li.Winners[i].Participant, fmt.Sprintf("tg://user?id=%d", li.Winners[i].Participant))
		}
	}

	return fmt.Sprintf(
		"🤖️ *抽奖任务:* `%s`.\n\n*抽奖配置:*\n积分要求: `%d`\n积分消耗: `%v`\n奖品数量: `%d`\n开奖方式: `%s`\n\n*任务状态:* %s",
		GetQuotableStr(li.Payload), li.Limit, li.Consume, li.Num, drawMsg, status,
	)
}

func (li *LotteryInstance) Update() bool {
	cfg, _ := jsoniter.Marshal(li)
	err := DB.Table(DBTName("Lottery")).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"config"}),
	}).Create(&DBLottery{
		ID:     li.ID,
		Config: string(cfg),
	}).Error

	if err != nil {
		DErrorf("Update Lottery Error | id=%s value=%s error=%v", li.ID, string(cfg), err.Error())
		return false
	}

	return true
}

func (li *LotteryInstance) Join(userId int64, username string) error {
	li.JoinLock.Lock()
	defer li.JoinLock.Unlock()

	if li.Status != 0 {
		return errors.New("❌ 抽奖活动不在有效时间范围内，请检查后再试 ~")
	}

	err := DB.Table(DBTName("Lottery_Participation")).Create(&DBLotteryParticipation{
		LotteryID:   li.ID,
		Participant: userId,
		Username:    username,
	}).Error

	if err != nil {
		DLogf("Join Lottery Error | id=%s user=%d error=%v", li.ID, userId, err.Error())
		return errors.New("❌ 您已经参加过这个活动了，请不要重复参加哦 ~")
	}

	if li.ParticipantCache > 0 {
		li.ParticipantCache += 1
	}

	return nil
}

func (li *LotteryInstance) Participants() int {
	if li.Status >= 0 {
		if li.ParticipantCache > 0 {
			return li.ParticipantCache
		}

		ret := int64(0)
		err := DB.Table(DBTName("Lottery_Participation")).Where("lotteryid = ?", li.ID).Count(&ret).Error
		if err != nil {
			DLogf("Fetch Lottery Participants Number Error | id=%s error=%v", li.ID, err.Error())
			return -1
		}

		li.ParticipantCache = int(ret)
		return li.ParticipantCache
	}
	return -1
}

func (li *LotteryInstance) StartLottery() {
	li.JoinLock.Lock()
	defer li.JoinLock.Unlock()

	if li.Status == -1 {
		li.Status = 0
		li.StartedAt = time.Now().Unix()
		li.Update()
		li.UpdateTelegramMsg()

		if li.Duration > 0 {
			lazyScheduler.After(li.Duration+time.Second, memutils.LSC("checkDraw", &CheckDrawArgs{
				LotteryId: li.ID,
			}))
		}
	}
}

func (li *LotteryInstance) CheckDraw(force bool) bool {
	li.JoinLock.Lock()
	defer li.JoinLock.Unlock()

	if li.Status == 0 {
		if force {
			// manual draw
			li.Status = 2
		} else if li.Duration > 0 && li.StartedAt > 0 && li.StartedAt+int64(li.Duration/time.Second) < time.Now().Unix() {
			// timeout draw
			li.Status = 2
		} else if li.Participant >= 0 && li.Participants() >= li.Participant {
			// participant exceeding draw
			li.Status = 2
		}

		// draw
		if li.Status == 2 {
			winners := []DBLotteryParticipation{}
			DB.Table(DBTName("Lottery_Participation")).Clauses(clause.OrderBy{
				Expression: clause.Expr{SQL: GetRandClause()},
			}).Where("lotteryid = ?", li.ID).Limit(li.Num).Find(&winners)

			li.Winners = winners
			li.Update()
			li.UpdateTelegramMsg()
			return true
		}
	}

	return false
}

func GetLottery(lotteryId string) *LotteryInstance {
	LotteryConfigLock.Lock()
	defer LotteryConfigLock.Unlock()
	if li, ok := LotteryConfigCache[lotteryId]; ok && li != nil {
		return li
	}

	ret := DBLottery{}
	err := DB.Table(DBTName("Lottery")).First(&ret, "id = ?", lotteryId).Error
	if err != nil {
		DErrorf("Fetch Lottery Error | id=%s error=%v", lotteryId, err.Error())
		return nil
	}

	li := LotteryInstance{}
	err = jsoniter.Unmarshal([]byte(ret.Config), &li)
	if err != nil {
		DErrorf("Unmarshal Lottery Error | id=%s error=%v", lotteryId, err.Error())
		return nil
	}

	LotteryConfigCache[li.ID] = &li
	return &li
}

func CreateLottery(groupId int64, payload string, limit int, consume bool, num int, duration int, participant int) *LotteryInstance {
	li := LotteryInstance{
		ID:        fmt.Sprintf("%d:%d:%d", Abs(groupId), time.Now().Unix(), rand.Intn(9999)),
		Status:    -1,
		GroupID:   groupId,
		CreatedAt: time.Now().Unix(),
		StartedAt: 0,

		Payload:     payload,
		Limit:       limit,
		Consume:     consume,
		Num:         num,
		Duration:    time.Minute * time.Duration(duration),
		Participant: participant,
	}

	if li.Update() {
		LotteryConfigLock.Lock()
		LotteryConfigCache[li.ID] = &li
		LotteryConfigLock.Unlock()
		return &li
	}
	return nil
}

func init() {
	GroupConfigCache = make(map[int64]*GroupConfig)
	LotteryConfigCache = make(map[string]*LotteryInstance)
}
