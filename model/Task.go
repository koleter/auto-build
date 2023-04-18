package model

import (
	"fmt"
	"time"
)

type Task struct {
	Id        int64  `xorm:"pk autoincr" json:"id"`
	ProjectId int64  `xorm:"index" json:"-"`
	GoVersion int64  `xorm:"index" json:"-"` // envid
	Branch    string `xorm:"varchar(10)" json:"branch"`
	MainFile  string `xorm:"varchar(20)" json:"main_file"` // 主文件
	DestFile  string `xorm:"varchar(20)" json:"dest_file"` // 目标文件
	DestOs    string `xorm:"varchar(10)" json:"dest_os"`   // 目标系统
	DestArch  string `xorm:"varchar(10)" json:"dest_arch"` // 目标架构
	Env       string `xorm:"varchar(255)" json:"env"`      // 环境变量key1=value1;key2=value2
}

type TaskLog struct {
	Id          int64     `xorm:"pk autoincr" json:"id"`
	TaskId      int64     `xorm:"index" json:"-"`
	Description string    `xorm:"varchar(50)" json:"description"`
	Status      int       `xorm:"index" json:"status"`
	Url         string    `xorm:"varchar(50)" json:"url"` //目标文件
	OutFilePath string    `xorm:"varchar(50)" json:"out_file_path"`
	ErrFilePath string    `xorm:"varchar(50)" json:"err_file_path"`
	CreateAt    time.Time `xorm:"datetime created" json:"create_at"`
	FinishAt    time.Time `xorm:"datetime updated" json:"finish_at"`
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

func UpdateTaskLogUrl(id int64, url string) {
	tl := &TaskLog{
		Url: url,
	}
	engine.Where("id = ?", id).Cols("url").Update(tl)
}

func UpdateTaskLogOut(id int64, filepath string) {
	tl := &TaskLog{
		OutFilePath: filepath,
	}
	engine.Where("id = ?", id).Cols("out_file_path").Update(tl)
}

func UpdateTaskLogErr(id int64, filepath string) {
	tl := &TaskLog{
		ErrFilePath: filepath,
	}
	engine.Where("id = ?", id).Cols("err_file_path").Update(tl)
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
	s := engine.NewSession()
	if projectid > 0 {
		s.Where("project_id = ?", projectid)
	}
	err := s.Find(&ts)
	return ts, err
}

func ListTaskLog(taskid int64) ([]*TaskLog, error) {
	tls := make([]*TaskLog, 0)
	s := engine.NewSession()
	if taskid > 0 {
		s.Where("task_id = ?", taskid)
	}
	err := s.Find(&tls)
	return tls, err
}
