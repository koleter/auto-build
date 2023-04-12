package model

import (
	_ "github.com/mattn/go-sqlite3"
	"igit.58corp.com/mengfanyu03/auto-build-go/log"
	"xorm.io/xorm"
)

var engine *xorm.Engine

func InitSqlLite(filepath string) (err error) {
	log.Infof("db file path:%s", filepath)
	engine, err = xorm.NewEngine("sqlite3", filepath)
	if err != nil {
		return err
	}

	return engine.Ping()
}

func AuthMergeTable() error {
	return engine.Sync(new(GoEnv), new(Project), new(Task), new(TaskLog))
}

func Close() {
	engine.Close()
}
