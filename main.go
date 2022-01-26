package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var VersionArg = false
var PingArg = false
var CleanArg = false
var setadmin = int64(0)

func main() {
	if VersionArg {
		fmt.Println(version)
		os.Exit(0)
	}

	if PingArg {
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
	if APIBind > 0 {
		InitRESTServer(APIBind, APIToken)
	}

	<-MakeSysChan()
	DInfo("shutting down.")
}

func init() {
	flag.StringVar(&TOKEN, "token", "", "telegram bot token")
	flag.StringVar(&TELEGRAMURL, "upstream", "", "telegram upstream api url")
	flag.StringVar(&DBCONN, "database", "", "mysql or its compatible database connection URL")
	flag.BoolVar(&VerboseMode, "verbose", false, "display all logs")
	flag.BoolVar(&VersionArg, "version", false, "display current version and exit")
	flag.BoolVar(&PingArg, "ping", false, "test the round time trip between bot and telegram server")
	flag.BoolVar(&CleanArg, "clean", false, "ignore all messages from polling")
	flag.Int64Var(&setadmin, "setadmin", 0, "set admin and delete all the other existing admins")

	flag.IntVar(&APIBind, "bind", 0, "specify a point number to bind and start an api server, for example 9876")
	flag.StringVar(&APIToken, "api-token", "", "specify a token that needs to be passed in when calling api methods")

	flag.Parse()
}
