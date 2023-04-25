package model

import (
	"github.com/hash-rabbit/auto-build/log"
	"github.com/hash-rabbit/snowflake"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

var engine *xorm.Engine
var node *snowflake.Node

func InitModel() {
	node, err := snowflake.NewNode(1)
}

func InitSqlLite(filepath string) (err error) {
	log.Infof("db file path:%s", filepath)
	engine, err = xorm.NewEngine("sqlite3", filepath)
	if err != nil {
		return err
	}

	return engine.Ping()
}

func AuthMergeTable() error {
	return engine.Sync(new(GoVersion), new(Project), new(Task), new(TaskLog))
}

func Close() {
	engine.Close()
}
