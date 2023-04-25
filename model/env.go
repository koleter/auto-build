package model

import (
	"fmt"
)

type GoVersion struct {
	Id        int64  `xorm:"pk" json:"id"` //TODO:因为前端精度丢失问题,暂将 id 转为 string
	Version   string `xorm:"varchar(10) not null" json:"version"`
	Os        string `xorm:"varchar(10)" json:"os"`
	Arch      string `xorm:"varchar(10)" json:"arch"`
	Url       string `xorm:"varchar(100)" json:"url"`
	Sha2      string `xorm:"varchar(64)" json:"sha2"`
	LocalPath string `xorm:"varchar(100)" json:"localpath"` // 本地的 go bin 上一级的绝对路径
}

func InsertGoVersion(ge *GoVersion) error {
	ge.Id = node.Generate().Int64()
	_, err := engine.InsertOne(ge)
	return err
}

func GetGoVersion(id int64) (*GoVersion, error) {
	v := &GoVersion{}
	has, err := engine.ID(id).Get(v)
	if !has {
		return nil, fmt.Errorf("couldn't find record")
	}
	return v, err
}

func GoVersionList(version string) ([]*GoVersion, error) {
	envs := make([]*GoVersion, 0)
	s := engine.NewSession()
	if len(version) > 0 {
		s.Where("version = ?", version)
	}
	err := s.Find(&envs)
	return envs, err
}
