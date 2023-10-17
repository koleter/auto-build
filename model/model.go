package model

import (
	"github.com/hash-rabbit/auto-build/config"
	"github.com/hash-rabbit/snowflake"
	_ "github.com/mattn/go-sqlite3"
	"github.com/subchen/go-log"
	"xorm.io/xorm"
)

var engine *xorm.Engine
var node *snowflake.Node

func InitModel() {
	err := InitNode()
	if err != nil {
		log.Panicf("init node error:%s", err)
	}

	err = InitSqlLite(config.C.SqlFile)
	if err != nil {
		log.Panicf("init sql lite error:%s", err)
	}

	err = AuthMergeTable()
	if err != nil {
		log.Panicf("auto merge table error:%s", err)
	}
}

func InitNode() error {
	var err error
	node, err = snowflake.NewNode(1)
	if err != nil {
		log.Errorf("create node failed")
		return err
	}
	return node.SetNodeAndStepBits(4, 4) //nodeid 4 位,step 4 位
}

func InitSqlLite(filepath string) (err error) {
	log.Infof("db file path:%s", filepath)
	engine, err = xorm.NewEngine("sqlite3", filepath)
	if err != nil {
		return err
	}
	engine.ShowSQL(false)

	return engine.Ping()
}

func AuthMergeTable() error {
	return engine.Sync(new(Project), new(Task), new(TaskLog))
}

func Close() {
	engine.Close()
}
