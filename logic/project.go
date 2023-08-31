package logic

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/hash-rabbit/auto-build/config"
	"github.com/hash-rabbit/auto-build/model"
	"github.com/hash-rabbit/auto-build/util"
	"github.com/subchen/go-log"
)

func AddPorject(wr http.ResponseWriter, r *http.Request) {
	p := &model.Project{}
	var err error
	if err := ParseParam(r, p); err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "param error", err.Error())
		return
	}
	log.Debugf("recv param:%+v", p)

	if err := checkProject(p); err != nil {
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
	log.Debugf("set workspace path:%s", p.WorkSpace)

	p.LocalPath, err = filepath.Abs(p.LocalPath)
	if err != nil {
		log.Errorf("filepath abs error:%s", err)
		writeError(wr, "path error", err.Error())
		return
	}

	if path_exist, err := PathExists(p.LocalPath); err != nil {
		log.Errorf("path %s check error:%s", p.LocalPath, err)
		writeError(wr, "path error", err.Error())
		return
	} else if path_exist {
		log.Errorf("path %s has exist", p.LocalPath)
		writeError(wr, "path error", "local path has exist")
		return
	}

	if pathExist, err := PathExists(getBarePath(p.Name)); err != nil {
		writeError(wr, "path error", err.Error())
		return
	} else if pathExist {
		os.RemoveAll(getBarePath(p.Name))
	}

	if err := os.MkdirAll(getBarePath(p.Name), os.ModePerm); err != nil {
		log.Errorf("mkdir %s error:%s", getBarePath(p.Name), err)
		writeError(wr, "path error", "make die error")
		return
	}

	if err := util.CloenWithBare(getBarePath(p.Name), p.Url, p.Token); err != nil {
		log.Errorf("clone bare error:%s", err)
		writeError(wr, "git error", "clone bare error")
		return
	}
	log.Debugf("clone to %s success", getBarePath(p.Name))

	if err := model.InsertProject(p); err != nil {
		log.Errorf("insert sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	writeSuccess(wr, "add project ok")
}

func getBarePath(projectName string) string {
	return filepath.Join(config.C.BarePath, projectName)
}

func checkProject(p *model.Project) error {
	if match, _ := regexp.MatchString("[0-9|a-z|A-Z|-|_]{1,30}", p.Name); !match {
		return errors.New("project name not allowed")
	}

	if _, err := model.GetProjectByName(p.Name); err == nil {
		return fmt.Errorf("name:%s 已存在", p.Name)
	}

	_, err := url.Parse(p.Url)
	if err != nil {
		return err
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
	ps, err := model.ListProject(r.FormValue("project_name"))
	if err != nil {
		log.Errorf("selet sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}
	writeJson(wr, ps)
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

	ts, err := model.ListTask(p.Id)
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

	os.RemoveAll(getBarePath(p.Name))

	writeSuccess(wr, "删除成功")
}

func ListBranch(wr http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		log.Errorf("parse param id error:%s", err)
		writeError(wr, "param error", err.Error())
		return
	}

	p, err := model.GetProject(id)
	if err != nil {
		log.Errorf("selet sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	err = util.Fetch(getBarePath(p.Name), "origin", p.Token)
	if err != nil {
		log.Errorf("git fetch error:%s", err)
		writeError(wr, "logic error", err.Error())
		return
	}

	branchs, err := util.BranchList(getBarePath(p.Name), "origin", p.Token)
	if err != nil {
		log.Errorf("get branch list error:%s", err)
		writeError(wr, "logic error", err.Error())
		return
	}

	sort.Slice(branchs, func(i, j int) bool {
		return branchs[i] < branchs[j]
	})

	writeJson(wr, branchs)
}
