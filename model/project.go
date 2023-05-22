package model

import (
	"fmt"
	"time"
)

type Project struct {
	Id         int64     `xorm:"pk" json:"id"` //TODO:因为前端精度丢失问题,暂将 id 转为 string
	Name       string    `xorm:"varchar(20) not null index" json:"name"`
	LocalPath  string    `xorm:"varchar(50) not null"  json:"path"`
	Url        string    `xorm:"varchar(50)"  json:"url"`
	MainBranch string    `xorm:"varchar(20) default master" json:"main_branch"`
	Token      string    `xorm:"varchar(50)"  json:"token"`
	GoMod      bool      `xorm:"bool" json:"go_mod"`
	WorkSpace  string    `xorm:"varchar(50)" json:"workspace"` //only go path(not mod) used
	Env        string    `xorm:"varchar(255)" json:"env"`      // 环境变量key1=value1;key2=value2
	DeletedAt  time.Time `xorm:"deleted" json:"-"`
}

func InsertProject(p *Project) error {
	p.Id = node.Generate().Int64()
	_, err := engine.InsertOne(p)
	return err
}

func GetProject(id int64) (*Project, error) {
	p := &Project{
		Id: id,
	}
	has, err := engine.Get(p)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("couldn't find record")
	}
	return p, err
}

func GetProjectByName(name string) (*Project, error) {
	p := &Project{
		Name: name,
	}
	has, err := engine.Get(p)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("couldn't find record")
	}
	return p, err
}

func ListProject(name string) ([]*Project, error) {
	ps := make([]*Project, 0)
	s := engine.NewSession()

	if len(name) > 0 {
		s.Where("name = ?", name)
	}

	err := s.Find(&ps)
	return ps, err
}

func DelProject(id int64) error {
	p := &Project{}
	n, err := engine.ID(id).Delete(p)
	if n != 1 {
		return fmt.Errorf("delete project affect line number:%d", n)
	}
	return err
}
