package logic

import (
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

	if len(p.Name) == 0 {
		writeError(wr, "param error", "project name length == 0")
		return
	}

	if ps, _ := model.ListProject(p.Name); len(ps) > 0 {
		writeError(wr, "param error", "project name has exist")
		return
	}

	_, err = url.Parse(p.Url)
	if err != nil {
		log.Errorf("url %s parse error:%s", p.Url, err)
		writeError(wr, "param error", "url parse error")
		return
	}

	if p.GoMod {
		p.WorkSpace = config.C.DefaultGoPath
	} else if len(p.WorkSpace) > 0 {
		p.WorkSpace, _ = filepath.Abs(p.WorkSpace)
	} else {
		log.Errorf("workspace not set")
		writeError(wr, "git error", "must set workspace")
		return
	}

	p.LocalPath, err = filepath.Abs(p.LocalPath)
	if err != nil {
		log.Errorf("path %s set error", p.LocalPath)
		writeError(wr, "path error", "path set error")
		return
	}

	path_exist, err := PathExists(p.LocalPath)
	if err != nil {
		log.Errorf("path %s check error", p.LocalPath)
		writeError(wr, "path error", "path check error")
		return
	}

	if path_exist {
		err := util.InitGit(p.LocalPath)
		if err != nil {
			log.Errorf("git init error:%s", err)
			writeError(wr, "path error", "git init error")
			return
		}
	} else {
		log.Errorf("path %s check error", p.LocalPath)
		writeError(wr, "path error", "path exist error")
		return
	}

	err = util.AddRemote(p.LocalPath, defaultRemoteName, GetUrl(p.Url, p.Token), false)
	if err != nil {
		log.Errorf("git add remote error:%s", err)
		writeError(wr, "path error", "git remote error")
		return
	}

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

// PathExists 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Errorf("mkdir failed![%v]\n", err)
		} else {
			return true, nil
		}
	}
	return false, err
}

// giturl 为 git 地址(http[s]),不用带秘钥
// 如果是 token 模式,则不输入 username
// 如果是 user:password 模式,则 token 就是 password
func GetUrl(giturl string, token string, username ...string) string {
	if len(token) == 0 { //公有库
		return giturl
	}

	user := "oauth2"
	if len(username) > 0 {
		user = url.QueryEscape(username[0])
	}

	if strings.HasPrefix(giturl, "https://") {
		return fmt.Sprintf("https://%s:%s@%s", user, token, strings.TrimPrefix(giturl, "https://"))
	} else if strings.HasPrefix(giturl, "http://") {
		return fmt.Sprintf("http://%s:%s@%s", user, token, strings.TrimPrefix(giturl, "http://"))
	}
	return ""
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

	ts, err := model.ListTask(pro.Id, 0)
	if err != nil {
		log.Errorf("selet sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	if len(ts) > 0 {
		writeError(wr, "logic error", "project have build tasks")
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
