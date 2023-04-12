package model

type GoEnv struct {
	Version   string `xorm:"varchar(10) not null pk"`
	Os        string `xorm:"varchar(10)"`
	Arch      string `xorm:"varchar(10)"`
	Url       string `xorm:"varchar(100)"`
	Sha2      string `xorm:"varchar(64)"`
	LocalPath string `xorm:"varchar(100)"` // 本地的 go bin 上一级的绝对路径
}

func (ge *GoEnv) Insert() error {
	_, err := engine.InsertOne(ge)
	return err
}

func GoenvList(version string) ([]*GoEnv, error) {
	envs := make([]*GoEnv, 0)
	s := engine.NewSession()
	if len(version) > 0 {
		s.Where("version = ?", version)
	}
	err := s.Find(&envs)
	return envs, err
}
