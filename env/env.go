package env

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hash-rabbit/auto-build/config"
	"github.com/hash-rabbit/auto-build/util"
	"github.com/robfig/cron"
	"github.com/subchen/go-log"
)

var c *cron.Cron

func Init() {
	if _, err := os.Stat(getDlpath()); os.IsNotExist(err) {
		err := util.Clone(getDlpath(), "https://github.com/golang/dl.git")
		if err != nil {
			log.Panicf("clone https://github.com/golang/dl.git error:%s", err)
		}
	}

	refreshVersions()

	c = cron.New()
	c.AddFunc("@daily", refreshVersions)
	c.Start()
}

func refreshVersions() {
	dirs, err := os.ReadDir(getDlpath())
	if err != nil {
		log.Errorf("read dir:%s error:%s", getDlpath(), err)
		return
	}

	for _, dir := range dirs {
		version := dir.Name()
		if !strings.HasPrefix(version, "go1.") {
			continue
		}

		if _, err := os.Stat(GetGoPath(version)); err == nil {
			continue
		}

		log.Infof("installing go version:%s", version)
		err := Install(GetGoPath(version), version)
		if err != nil {
			log.Errorf("install version:%s error:%s", version, err)
			continue
		}
	}
}

func getDlpath() string {
	return filepath.Join(config.C.GoEnvPath, "dl")
}

func GetGoPath(version string) string {
	return filepath.Join(config.C.GoEnvPath, version)
}

func ListEnv() ([]string, error) {
	dirs, err := os.ReadDir(getDlpath())
	if err != nil {
		log.Errorf("read dir:%s error:%s", getDlpath(), err)
		return nil, err
	}

	envs := make([]string, 0)
	for _, dir := range dirs {
		if strings.HasPrefix(dir.Name(), "go1.") {
			envs = append(envs, dir.Name())
		}
	}

	return envs, nil
}
