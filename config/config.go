package config

import (
	"github.com/BurntSushi/toml"
	"igit.58corp.com/mengfanyu03/auto-build-go/log"
)

type Config struct {
	Port          int    `toml:"port"`
	GoEnvPath     string `toml:"go_env_path"`     //存放 golang 环境的
	DefaultGoPath string `toml:"default_go_path"` //默认 go_path,主要用于 gomod
	DistPath      string `toml:"dist_path"`       //编译完成的文件存放位置
	SqlFile       string `toml:"sql_file"`
}

var C *Config

func LoadConfig(file string) {
	C = &Config{}
	_, err := toml.DecodeFile(file, C)
	if err != nil {
		log.Panicf("decode config error:%s", err)
	}
}
