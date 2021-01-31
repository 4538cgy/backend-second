package main

import (
	"github.com/4538cgy/backend-second/api"
	"github.com/4538cgy/backend-second/config"
	"github.com/4538cgy/backend-second/database"
	"github.com/4538cgy/backend-second/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.Get()
	log.Init(cfg)

	dbmgr, err := database.NewDBManager(cfg)
	if err != nil {
		log.Fatal("database manager create failed..", err.Error())
	}

	err = dbmgr.Connect()
	if err != nil {
		log.Fatal("connection failed... ", err.Error())
	} else {
		log.Infof("connection ok.. -> %s", dbmgr.DSN())
	}

	api.StartAPI(cfg, dbmgr)

	sig := make(chan os.Signal, 32)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
