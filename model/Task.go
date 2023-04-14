package model

import "time"

type Task struct {
	Id        int64  `xorm:"pk autoincr"`
	ProjectId int64  `xorm:"index"`
	Branch    string `xorm:"varchar(10)"`
	MainFile  string `xorm:"varchar(20)"`  // 主文件
	DistFile  string `xorm:"varchar(20)"`  // 目标文件
	GoVersion int64  `xorm:"varchar(10)"`  // envid
	Env       string `xorm:"varchar(255)"` // 环境变量key1=value1;key2=value2
}

type TaskLog struct {
	Id       int64     `xorm:"pk autoincr"`
	TaskId   int64     `xorm:"index"`
	Success  bool      `xorm:"bool"`
	Url      string    `xorm:"varchar(50)"`
	CreateAt time.Time `xorm:"datetime"`
	FinishAt time.Time `xorm:"datetime"`
}

func (t *Task) Insert() error {
	_, err := engine.InsertOne(t)
	return err
}

func (tl *TaskLog) Insert() error {
	_, err := engine.InsertOne(tl)
	return err
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
