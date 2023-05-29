package logic

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hash-rabbit/auto-build/config"
	"github.com/hash-rabbit/auto-build/model"
	"github.com/hash-rabbit/auto-build/util"
	"github.com/subchen/go-log"
)

var defaultRemoteName = "build"

func AddPorject(wr http.ResponseWriter, r *http.Request) {
	p := &model.Project{}
	err := ParseParam(r, p)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "param error", err.Error())
		return
	}
	log.Debugf("recv param:%+v", p)

	err = checkProject(p)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "param error", err.Error())
		return
	}
	log.Debugf("project:%s check success", p.Name)

	if p.GoMod {
		p.WorkSpace = config.C.DefaultGoPath
	} else if len(p.WorkSpace) > 0 {
		p.WorkSpace, _ = filepath.Abs(p.WorkSpace)
	} else {
		log.Errorf("workspace not set")
		writeError(wr, "git error", "must set workspace")
		return
	}

	path_exist, err := PathExists(p.LocalPath)
	if err != nil {
		log.Errorf("path %s check error:%s", p.LocalPath, err)
		writeError(wr, "path error", err.Error())
		return
	}

	if !path_exist {
		log.Debugf("clone %s to %s token:%s", p.Url, p.LocalPath, p.Token)
		err = util.Clone(p.LocalPath, p.Url, strings.Split(p.Token, " ")...)
		if err != nil {
			log.Errorf("path %s check error", p.LocalPath)
			writeError(wr, "path error", "path check error")
			return
		}
	}

	var url string
	if len(p.Token) > 0 {
		url = util.GetUrl(p.Url, strings.Split(p.Token, " ")...)
	} else {
		url = p.Url
	}

	log.Debugf("add remote url:%s", url)
	err = util.AddRemote(p.LocalPath, defaultRemoteName, url, false)
	if err != nil {
		log.Errorf("git add remote error:%s", err)
		writeError(wr, "path error", "git remote error")
		return
	}

	log.Debug("git pull")
	err = util.Pull(p.LocalPath, defaultRemoteName, "")
	if err != nil {
		log.Errorf("git pull error:%s", err)
		writeError(wr, "path error", "git pull error")
		return
	}

	err = model.InsertProject(p)
	if err != nil {
		log.Errorf("insert sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	writeSuccess(wr, "add project ok")
}

func checkProject(p *model.Project) error {
	if len(p.Name) == 0 || len(p.Name) > 30 {
		return errors.New("name 长度在  1-30 个字符以内")
	}

	if _, err := model.GetProjectByName(p.Name); err == nil {
		return fmt.Errorf("name:%s 已存在", p.Name)
	}

	u, err := url.Parse(p.Url)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(u.EscapedPath(), ".git") {
		return fmt.Errorf("url should include .git")
	}

	if !filepath.IsAbs(p.LocalPath) {
		return fmt.Errorf("local path should in abs")
	}

	return nil
}

// PathExists 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if !fi.IsDir() {
			return true, fmt.Errorf("path:%s is file", path)
		}
		return true, nil
	}

	return os.IsExist(err), nil
}

func ListPorject(wr http.ResponseWriter, r *http.Request) {
	ps, err := model.ListProject("")
	if err != nil {
		log.Errorf("selet sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}
	writeJson(wr, ps)
}

func PullPorject(wr http.ResponseWriter, r *http.Request) {
	param, err := checkParam(r)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "param error", err.Error())
		return
	}

	var projectId int64
	if param["project_id"] != nil {
		projectId = int64(param["project_id"].(float64))
	}

	p, err := model.GetProject(projectId)
	if err != nil {
		log.Errorf("select param error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	err = util.Pull(p.LocalPath, defaultRemoteName, "")
	if err != nil {
		log.Errorf("git pull error:%s", err)
		writeError(wr, "git error", err.Error())
		return
	}

	writeSuccess(wr, "git pull success")
}

func DelPorject(wr http.ResponseWriter, r *http.Request) {
	p := &model.Project{}
	err := ParseParam(r, p)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "param error", err.Error())
		return
	}

	pro, err := model.GetProject(p.Id)
	if err != nil {
		log.Errorf("selet sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	ts, err := model.ListTask(p.Id, 0)
	if err != nil {
		log.Errorf("selet sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	if len(ts) > 0 {
		writeError(wr, "logic error", "请先删除该工程的任务")
		return
	}

	err = model.DelProject(pro.Id)
	if err != nil {
		log.Errorf("delete sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	err = util.RmRemote(pro.LocalPath, defaultRemoteName)
	if err != nil {
		log.Errorf("delete git remote error:%s", err)
		writeError(wr, "logic error", err.Error())
		return
	}

	writeSuccess(wr, "删除成功")
}
