package model

type Project struct {
	Id        int64  `xorm:"pk" json:"id"`
	Name      string `xorm:"varchar(20) not null" json:"name"`
	LocalPath string `xorm:"varchar(50) not null"  json:"path"`
	Url       string `xorm:"varchar(50)"  json:"url"`
	Token     string `xorm:"varchar(50)"  json:"token"`
	GoMod     bool   `xorm:"bool" json:"go_mod"`
	WorkSpace string `xorm:"varchar(50)" json:"go_path"` //only go path(not mod) used
	Env       string `xorm:"varchar(255)" json:"env"`    // 环境变量key1=value1;key2=value2
}

func (p *Project) Insert() error {
	_, err := engine.InsertOne(p)
	return err
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
