package model

type GoVersion struct {
	Id        int64  `xorm:"pk" json:"id"`
	Version   string `xorm:"varchar(10) not null" json:"version"`
	Os        string `xorm:"varchar(10)" json:"os"`
	Arch      string `xorm:"varchar(10)" json:"arch"`
	Url       string `xorm:"varchar(100)" json:"url"`
	Sha2      string `xorm:"varchar(64)" json:"sha2"`
	LocalPath string `xorm:"varchar(100)" json:"localpath"` // 本地的 go bin 上一级的绝对路径
}

func (ge *GoVersion) Insert() error {
	_, err := engine.InsertOne(ge)
	return err
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
