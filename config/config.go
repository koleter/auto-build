package config

import (
	"github.com/BurntSushi/toml"
	"github.com/subchen/go-log"
)

type Config struct {
	Port          int    `toml:"port"`
	LogPath       string `toml:"log_path"`
	LogLevel      string `toml:"log_level"`
	RecordPath    string `toml:"record_path"`     //存放编译日志
	GoEnvPath     string `toml:"go_env_path"`     //存放 golang 环境的
	DefaultGoPath string `toml:"default_go_path"` //默认 go_path,主要用于 gomod
	DestPath      string `toml:"dest_path"`       //编译完成的文件存放位置
	SqlFile       string `toml:"sql_file"`
	WebPath       string `toml:"web_path"` //前端路径
}

var C *Config

func LoadConfig(file string) {
	C = &Config{}
	_, err := toml.DecodeFile(file, C)
	if err != nil {
		log.Panicf("decode config error:%s", err)
	}
}
