package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

var DBCONN = ""
var MYSQLDB *sql.DB
var UpdateLock sync.Mutex
var GroupConfigLock sync.Mutex
var LotteryConfigLock sync.Mutex

type UpdateMethod int

const (
	UMSet UpdateMethod = iota
	UMAdd
	UMDel
)

type CreditInfo struct {
	Username string `json:"username"`
	Name     string `json:"nickname"`
	ID       int64  `json:"id"`
	Credit   int64  `json:"credit"`
	GroupId  int64  `json:"groupId"`
}

var GroupConfigCache map[int64]*GroupConfig
var LotteryConfigCache map[string]*LotteryInstance

type GroupConfig struct {
	ID            int64
	Admins        []int64
	BannedForward []int64
	MergeTo       int64

	Locale           string
	MustFollow       string
	MustFollowOnJoin bool
	MustFollowOnMsg  bool

	AntiSpoiler bool
	DisableWarn bool
}

func InitDatabase() (err error) {
	MYSQLDB, err = sql.Open("mysql", DBCONN)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err = MYSQLDB.PingContext(ctx)
		cancel()
	}

	return
}

func InitTables() {
	q, err := MYSQLDB.Query(`CREATE TABLE IF NOT EXISTS MiaoKeeper_Config (
		k VARCHAR(128) NOT NULL PRIMARY KEY,
		v TEXT NOT NULL
	) DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		DErrorf("Config Table Creation Error | error=%v", err.Error())
		os.Exit(1)
	}
	if q != nil {
		q.Close()
	}

	q, err = MYSQLDB.Query(`CREATE TABLE IF NOT EXISTS MiaoKeeper_Lottery (
		id VARCHAR(128) NOT NULL PRIMARY KEY,
		config TEXT NOT NULL,
		createdat DATETIME DEFAULT CURRENT_TIMESTAMP
	) DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		DErrorf("Lottery Table Creation Error | error=%v", err.Error())
		os.Exit(1)
	}
	if q != nil {
		q.Close()
	}

	q, err = MYSQLDB.Query(`CREATE TABLE IF NOT EXISTS MiaoKeeper_Lottery_Participation (
		id VARCHAR(128) NOT NULL,
		participant BIGINT NOT NULL,
		username TEXT NOT NULL,
		createdat DATETIME DEFAULT CURRENT_TIMESTAMP,
		INDEX (id),
		UNIQUE KEY uniq_participant (id, participant)
	) DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		DErrorf("Lottery Participation Table Creation Error | error=%v", err.Error())
		os.Exit(1)
	}
	if q != nil {
		q.Close()
	}
}

func ReadConfig(key string) string {
	ret := ""
	err := MYSQLDB.QueryRow(`SELECT v FROM MiaoKeeper_Config WHERE k = ?;`, key).Scan(&ret)
	if err != nil {
		DLogf("Config Read Error | key=%s error=%v", key, err.Error())
	}
	return ret
}

func WriteConfig(key, value string) {
	q, err := MYSQLDB.Query(`INSERT INTO MiaoKeeper_Config
			(k, v)
		VALUES
			(?, ?)
		ON DUPLICATE KEY UPDATE
			v = VALUES(v)`, key, value)
	if err != nil {
		DLogf("Config Write Error | key=%s value=%s error=%v", key, value, err.Error())
	}
	if q != nil {
		q.Close()
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
			GroupConfigCache[groupId] = gc
			return gc
		}
	}
	return nil
}

func SetGroupConfig(groupId int64, gc *GroupConfig) *GroupConfig {
	if !IsGroup(groupId) {
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
	q, err := MYSQLDB.Query(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS MiaoKeeper_Credit_%d (
		userid BIGINT NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		username TEXT NOT NULL,
		credit BIGINT NOT NULL
	) DEFAULT CHARSET=utf8mb4`, Abs(groupId)))
	if err != nil {
		DErrorf("Table Creation Error | error=%v", err.Error())
	}
	if q != nil {
		q.Close()
	}

	if GetGroupConfig(groupId) == nil {
		NewGroupConfig(groupId)
	}
}

func NewGroupConfig(groupId int64) *GroupConfig {
	return SetGroupConfig(groupId, &GroupConfig{
		ID:            groupId,
		Admins:        make([]int64, 0),
		BannedForward: make([]int64, 0),
	})
}

func ReadConfigs() {
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

func (gc *GroupConfig) UpdateAdmin(userId int64, method UpdateMethod) bool {
	changed := false
	if method == UMSet {
		if len(gc.Admins) != 1 || gc.Admins[0] != userId {
			changed = true
			gc.Admins = []int64{userId}
		}
	} else if method == UMAdd {
		gc.Admins, changed = AddIntoInt64Arr(gc.Admins, userId)
	} else if method == UMDel {
		gc.Admins, changed = DelFromInt64Arr(gc.Admins, userId)
	}
	SetGroupConfig(gc.ID, gc)
	return changed
}

func (gc *GroupConfig) UpdateBannedForward(id int64, method UpdateMethod) bool {
	changed := false
	if method == UMSet {
		if len(gc.BannedForward) != 1 || gc.BannedForward[0] != id {
			changed = true
			gc.BannedForward = []int64{id}
		}
	} else if method == UMAdd {
		gc.BannedForward, changed = AddIntoInt64Arr(gc.BannedForward, id)
	} else if method == UMDel {
		gc.BannedForward, changed = DelFromInt64Arr(gc.BannedForward, id)
	}
	SetGroupConfig(gc.ID, gc)
	return changed
}

func (gc *GroupConfig) IsAdmin(userId int64) bool {
	return I64In(&gc.Admins, userId)
}

func (gc *GroupConfig) IsBannedForward(id int64) bool {
	return I64In(&gc.BannedForward, id)
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
	err := MYSQLDB.QueryRow(fmt.Sprintf(`SELECT userid, name, username, credit FROM MiaoKeeper_Credit_%d WHERE userid = ?;`, Abs(realGroup)), userId).Scan(
		&ret.ID, &ret.Name, &ret.Username, &ret.Credit,
	)
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
	row, _ := MYSQLDB.Query(fmt.Sprintf(`SELECT userid, name, username, credit FROM MiaoKeeper_Credit_%d ORDER BY credit DESC LIMIT ?;`, Abs(realGroup)), limit)
	for row.Next() {
		ret := &CreditInfo{}
		row.Scan(&ret.ID, &ret.Name, &ret.Username, &ret.Credit)
		if ret.ID > 0 {
			ret.GroupId = groupId
		}
		returns = append(returns, ret)
	}
	row.Close()
	return returns
}

// does not apply MergeTo
func DumpCredits(groupId int64) [][]string {
	ret := [][]string{}
	id, name, username, credit := int64(0), "", "", int64(0)
	row, _ := MYSQLDB.Query(fmt.Sprintf(`SELECT userid, name, username, credit FROM MiaoKeeper_Credit_%d WHERE credit > 0 ORDER BY credit;`, Abs(groupId)))
	for row.Next() {
		row.Scan(&id, &name, &username, &credit)
		if id > 0 && credit > 0 {
			ret = append(ret, []string{strconv.FormatInt(id, 10), name, username, strconv.FormatInt(credit, 10)})
		}
	}
	row.Close()

	DInfof("Credit Dump | group=%d columns=%d", groupId, len(ret))
	return ret
}

// does not apply MergeTo
func FlushCredits(groupId int64, records [][]string) {
	if len(records) == 0 {
		return
	}

	params := []interface{}{}
	sqlCmd := fmt.Sprintf(`INSERT INTO MiaoKeeper_Credit_%d (userid, name, username, credit) VALUES`, Abs(groupId))
	for _, r := range records {
		sqlCmd += ` (?, ?, ?, ?),`
		for _, rc := range r {
			params = append(params, rc)
		}
	}
	sqlCmd = sqlCmd[0 : len(sqlCmd)-1]
	sqlCmd += ` ON DUPLICATE KEY UPDATE
		name = VALUES(name),
		username = VALUES(username),
		credit = VALUES(credit) + credit`

	query, err := MYSQLDB.Query(sqlCmd, params...)
	if err != nil {
		DErrorE(err, "Database Credit Flush Error")
	}
	if query != nil {
		query.Close()
	}

	DInfof("Flush Credit | group=%d columns=%d", groupId, len(records))
}

func UpdateCredit(user *CreditInfo, method UpdateMethod, value int64) *CreditInfo {
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

	var query *sql.Rows
	var err error

	realGroup := GetAliasedGroup(user.GroupId)
	if method != UMDel {
		query, err = MYSQLDB.Query(fmt.Sprintf(`INSERT INTO MiaoKeeper_Credit_%d
				(userid, name, username, credit)
			VALUES
				(?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				username = VALUES(username),
				credit = VALUES(credit)
			`, Abs(realGroup)), user.ID, user.Name, user.Username, user.Credit)
	} else if realGroup == user.GroupId {
		// when the method is UMDel, do not delete aliased credit
		query, err = MYSQLDB.Query(fmt.Sprintf(`DELETE FROM MiaoKeeper_Credit_%d
			WHERE userid = ?;`, Abs(user.GroupId)), user.ID)
	}
	if err != nil {
		DErrorE(err, "Database Credit Update Error")
	}
	if query != nil {
		query.Close()
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

	Payload     string
	Limit       int
	Consume     bool
	Num         int
	Duration    int
	Participant int

	Winners          []int64
	WinnersName      []string
	ParticipantCache int        `json:"-"`
	JoinLock         sync.Mutex `json:"-"`
}

func (li *LotteryInstance) UpdateTelegramMsg() *tb.Message {
	btns := []string{}
	if li.Status == 0 {
		btns = append(btns, fmt.Sprintf("🤏 我要抽奖|lt/%d/1/%s", li.GroupID, li.ID))
	}
	if li.Status >= 0 && li.Status < 2 {
		btns = append(btns, fmt.Sprintf("📦 手动开奖[管理]|lt/%d/3/%s", li.GroupID, li.ID))
	}
	if li.Status == -1 {
		btns = append(btns, fmt.Sprintf("🎡 开启活动[管理]|lt/%d/2/%s", li.GroupID, li.ID))
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
			drawMsg += " *或* "
		}
		drawMsg += fmt.Sprintf("%d 小时后自动开奖", li.Duration)
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
	if len(li.Winners) > 0 && len(li.Winners) <= len(li.WinnersName) {
		status += "\n\n*🏆 获奖者:*"
		for i := range li.Winners {
			status += fmt.Sprintf("\n`%2d.` `%s` ([%d](%s))\n", i+1, GetQuotableStr(li.WinnersName[i]), li.Winners[i], fmt.Sprintf("tg://user?id=%d", li.Winners[i]))
		}
	}

	return fmt.Sprintf(
		"🤖️ *抽奖任务:* `%s`.\n\n*抽奖配置:*\n积分要求: `%d`\n积分消耗: `%v`\n奖品数量: `%d`\n开奖方式: `%s`\n\n*任务状态:* %s",
		GetQuotableStr(li.Payload), li.Limit, li.Consume, li.Num, drawMsg, status,
	)
}

func (li *LotteryInstance) Update() bool {
	cfg, _ := jsoniter.Marshal(li)
	q, err := MYSQLDB.Query(`INSERT INTO MiaoKeeper_Lottery
		(id, config)
	VALUES
		(?, ?)
	ON DUPLICATE KEY UPDATE
		config = VALUES(config)
	`, li.ID, string(cfg))

	if q != nil {
		q.Close()
	}
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

	q, err := MYSQLDB.Query(`INSERT INTO MiaoKeeper_Lottery_Participation
			(id, participant, username)
		VALUES
			(?, ?, ?)`, li.ID, userId, username)
	if q != nil {
		q.Close()
	}
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

		ret := 0
		err := MYSQLDB.QueryRow(`SELECT count(*) FROM MiaoKeeper_Lottery_Participation WHERE id = ?;`, li.ID).Scan(&ret)
		if err != nil {
			DLogf("Fetch Lottery Participants Number Error | id=%s error=%v", li.ID, err.Error())
			return -1
		}

		li.ParticipantCache = ret
		return ret
	}
	return -1
}

func (li *LotteryInstance) CheckDraw(force bool) bool {
	li.JoinLock.Lock()
	defer li.JoinLock.Unlock()

	if li.Status == 0 {
		if force {
			// manual draw
			li.Status = 2
		} else if li.Duration > 0 && li.CreatedAt+int64(li.Duration)*3600 < time.Now().Unix() {
			// timeout draw
			li.Status = 2
		} else if li.Participant >= 0 && li.Participants() >= li.Participant {
			// participant exceeding draw
			li.Status = 2
		}

		// draw
		if li.Status == 2 {
			li.Winners = []int64{}
			li.WinnersName = []string{}
			row, _ := MYSQLDB.Query(`SELECT participant, username FROM MiaoKeeper_Lottery_Participation WHERE id = ? ORDER BY RAND() LIMIT ?;`, li.ID, li.Num)
			for row.Next() {
				userid, username := int64(0), ""
				row.Scan(&userid, &username)
				if userid > 0 {
					li.Winners = append(li.Winners, userid)
					li.WinnersName = append(li.WinnersName, username)
				}
			}

			row.Close()
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

	ret := ""
	err := MYSQLDB.QueryRow(`SELECT config FROM MiaoKeeper_Lottery WHERE id = ?;`, lotteryId).Scan(&ret)
	if err != nil {
		DErrorf("Fetch Lottery Error | id=%s error=%v", lotteryId, err.Error())
		return nil
	}

	li := LotteryInstance{}
	err = jsoniter.Unmarshal([]byte(ret), &li)
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

		Payload:     payload,
		Limit:       limit,
		Consume:     consume,
		Num:         num,
		Duration:    duration,
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
