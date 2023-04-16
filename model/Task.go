package model

import (
	"fmt"
	"time"
)

type Task struct {
	Id        int64  `xorm:"pk autoincr"`
	ProjectId int64  `xorm:"index"`
	GoVersion int64  `xorm:"index"` // envid
	Branch    string `xorm:"varchar(10)"`
	MainFile  string `xorm:"varchar(20)"`  // 主文件
	DestFile  string `xorm:"varchar(20)"`  // 目标文件
	DestOs    string `xorm:"varchar(10)"`  // 目标系统
	DestArch  string `xorm:"varchar(10)"`  // 目标架构
	Env       string `xorm:"varchar(255)"` // 环境变量key1=value1;key2=value2
}

type TaskLog struct {
	Id          int64     `xorm:"pk autoincr"`
	TaskId      int64     `xorm:"index"`
	Description string    `xorm:"varchar(50)"`
	Status      int       `xorm:"index"`
	Url         string    `xorm:"varchar(50)"` //目标文件
	CreateAt    time.Time `xorm:"datetime created"`
	FinishAt    time.Time `xorm:"datetime updated"`
}

func InsertTask(t *Task) error {
	_, err := engine.InsertOne(t)
	return err
}

func InsertTaskLog(tl *TaskLog) error {
	_, err := engine.InsertOne(tl)
	return err
}

func UpdateTaskLog(id int64, status int) {
	tl := &TaskLog{
		Status: status,
	}
	engine.Where("id = ?", id).Cols("status").Update(tl)
}

func GetTask(id int64) (*Task, error) {
	t := &Task{}
	has, err := engine.Where("id = ?", id).Get(t)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("couldn't find task")
	}
	return t, nil
}

func ListTask(projectid int64) ([]*Task, error) {
	ts := make([]*Task, 0)
	err := engine.Where("project_id = ?", projectid).Find(&ts)
	return ts, err
}

func ListTaskLog(taskid int64) ([]*TaskLog, error) {
	tls := make([]*TaskLog, 0)
	err := engine.Where("task_id = ?", taskid).Find(&tls)
	return tls, err
}
