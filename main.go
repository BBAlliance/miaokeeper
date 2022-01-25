package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var version = false
var ping = false
var escape = false
var setadmin = int64(0)

func main() {
	if version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if ping {
		InitTelegram()
		t := time.Now().UnixMilli()
		Bot.GetCommands()
		// resp, _ := Bot.Raw("getMe", nil)
		t1 := time.Now().UnixMilli() - t
		Bot.GetCommands()
		// _, _ = Bot.Raw("getMe", nil)
		t2 := time.Now().UnixMilli() - t - t1
		Bot.GetCommands()
		// _, _ = Bot.Raw("getMe", nil)
		t3 := time.Now().UnixMilli() - t - t1 - t2
		fmt.Printf("Response Time: %dms, %dms, %dms (avg: %dms)\n", t1, t2, t3, (t1+t2+t3)/3)
		os.Exit(0)
	}

	if err := InitDatabase(); err != nil {
		DErrorE(err, "Database Error | Cannot initialize database.")
		os.Exit(1)
	}
	DInfo("Database Init | Database is initialzed.")

	InitTables()
	ReadConfigs()

	if setadmin > 0 {
		UpdateAdmin(setadmin, UMSet)
		os.Exit(0)
	}

	InitTelegram()

	<-MakeSysChan()
	DInfo("shutting down.")
}

func init() {
	flag.StringVar(&TOKEN, "token", "", "telegram bot token")
	flag.StringVar(&TELEGRAMURL, "upstream", "", "telegram upstream api url")
	flag.StringVar(&DBCONN, "database", "", "mysql or its compatible database connection URL")
	flag.BoolVar(&VerboseMode, "verbose", false, "display all logs")
	flag.BoolVar(&version, "version", false, "display current version and exit")
	flag.BoolVar(&ping, "ping", false, "test the round time trip between bot and telegram server")
	flag.BoolVar(&escape, "escape", false, "ignore all messages from polling")
	flag.Int64Var(&setadmin, "setadmin", 0, "set admin and delete all the other existing admins")

	flag.Parse()
}
