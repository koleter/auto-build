package config

import (
	"github.com/BurntSushi/toml"
	"igit.58corp.com/mengfanyu03/auto-build-go/log"
)

type Config struct {
	Port          int    `toml:"port"`
	GoEnvPath     string `toml:"go_env_path"`
	DefaultGoPath string `toml:"default_go_path"`
	DistPath      string `toml:"dist_path"`
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
