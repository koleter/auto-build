package model

import (
	"fmt"
)

type Project struct {
	Id        int64  `xorm:"pk" json:"id"` //TODO:因为前端精度丢失问题,暂将 id 转为 string
	Name      string `xorm:"varchar(20) not null" json:"name"`
	LocalPath string `xorm:"varchar(50) not null"  json:"path"`
	Url       string `xorm:"varchar(50)"  json:"url"`
	Token     string `xorm:"varchar(50)"  json:"token"`
	GoMod     bool   `xorm:"bool" json:"go_mod"`
	WorkSpace string `xorm:"varchar(50)" json:"workspace"` //only go path(not mod) used
	Env       string `xorm:"varchar(255)" json:"env"`      // 环境变量key1=value1;key2=value2
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

func ListProject(name string) ([]*Project, error) {
	ps := make([]*Project, 0)
	s := engine.NewSession()

	if len(name) > 0 {
		s.Where("name = ?", name)
	}

	err := s.Find(&ps)
	return ps, err
}
