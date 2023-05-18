package model

import (
	"fmt"
	"time"
)

type Task struct {
	Id        int64  `xorm:"pk" json:"id"`
	ProjectId int64  `xorm:"index" json:"project_id"`
	GoVersion int64  `xorm:"index" json:"go_version_id"` // envid
	Branch    string `xorm:"varchar(10)" json:"branch"`
	MainFile  string `xorm:"varchar(20)" json:"main_file"` // 主文件
	DestFile  string `xorm:"varchar(20)" json:"dest_file"` // 目标文件
	DestOs    string `xorm:"varchar(10)" json:"dest_os"`   // 目标系统
	DestArch  string `xorm:"varchar(10)" json:"dest_arch"` // 目标架构
	Env       string `xorm:"varchar(255)" json:"env"`      // 环境变量key1=value1;key2=value2
}

type TaskLog struct {
	Id          int64     `xorm:"pk" json:"id"`
	TaskId      int64     `xorm:"index" json:"task_id"`
	Description string    `xorm:"varchar(50)" json:"description"`
	Status      int       `xorm:"index" json:"status"`           //0:init,1:building,2:success,3:failed TODO:100代表 success,0-100 代表进度,<0 代表失败
	Url         string    `xorm:"varchar(50)" json:"url"`        //目标文件
	LocalPath   string    `xorm:"varchar(50)" json:"local_path"` //生成文件本地路径
	Size        int64     `xorm:"default 0" json:"size"`         // TODO:增加编译后本地校验
	Sha2        string    `xorm:"varchar(50)" json:"sha2"`       // TODO:生成后生成 sha2
	OutFilePath string    `xorm:"varchar(50)" json:"out_file_path"`
	ErrFilePath string    `xorm:"varchar(50)" json:"err_file_path"`
	CreateAt    time.Time `xorm:"datetime created" json:"create_at"`
	FinishAt    time.Time `xorm:"datetime updated" json:"finish_at"`
}

func InsertTask(t *Task) error {
	t.Id = node.Generate().Int64()
	_, err := engine.InsertOne(t)
	return err
}

func InsertTaskLog(tl *TaskLog) error {
	tl.Id = node.Generate().Int64()
	_, err := engine.InsertOne(tl)
	return err
}

func UpdateTaskLog(id int64, status int) {
	tl := &TaskLog{
		Status: status,
	}
	engine.Where("id = ?", id).Cols("status").Update(tl)
}

func UpdateTaskLogDescription(id int64, desc string) {
	tl := &TaskLog{
		Description: desc,
	}
	engine.Where("id = ?", id).Cols("description").Update(tl)
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

type TaskLogInfo struct {
	TaskLog `xorm:"extends"`
	Name    string `json:"name"`
	Branch  string `json:"branch"`
	Version string `json:"version"`
}

func ListTaskLog(versionId, projectId, taskid int64, limit int, offset ...int) ([]*TaskLogInfo, error) {
	tls := make([]*TaskLogInfo, 0)
	s := engine.NewSession()
	s.Table("task_log").Join("INNER", "task", "task.id = task_log.task_id").
		Join("INNER", "project", "task.project_id = project.id").
		Join("INNER", "go_version", "task.go_version = go_version.id")

	if versionId > 0 {
		s.Where("go_version.id = ?", versionId)
	}

	if projectId > 0 {
		s.Where("project.id = ?", versionId)
	}

	if taskid > 0 {
		s.Where("task.id = ?", taskid)
	}

	err := s.Desc("create_at").Limit(limit, offset...).Find(&tls)
	return tls, err
}

func GetTaskLog(record_id int64) (*TaskLog, error) {
	t := &TaskLog{}
	has, err := engine.Where("id = ?", record_id).Get(t)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("couldn't find record id:%d", record_id)
	}

	return t, nil
}
