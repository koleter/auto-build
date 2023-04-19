package logic

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hash-rabbit/auto-build/config"
	"github.com/hash-rabbit/auto-build/log"
	"github.com/hash-rabbit/auto-build/model"
	"github.com/hash-rabbit/auto-build/util"
)

// build status
const (
	Init = iota
	Running
	Success
	Failed
)

func AddTask(wr http.ResponseWriter, r *http.Request) {
	param, err := checkParam(r)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}

	projectId := int64(param["project_id"].(float64))
	if _, err := model.GetProject(projectId); err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	go_ver_id := int64(param["go_version_id"].(float64))
	if _, err := model.GetGoVersion(go_ver_id); err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	if len(param["branch"].(string)) == 0 || len(param["main_file"].(string)) == 0 || len(param["dest_file"].(string)) == 0 {
		log.Errorf("check param error")
		writeError(wr, "params error", "param error")
		return
	}

	destos := runtime.GOOS
	if param["dest_os"] != nil && len(param["dest_os"].(string)) > 0 {
		destos = param["dest_os"].(string)
	}
	destArch := runtime.GOARCH
	if param["dest_arch"] != nil && len(param["dest_arch"].(string)) > 0 {
		destArch = param["dest_arch"].(string)
	}

	env := ""
	if param["env"] != nil {
		env = param["env"].(string)
	}

	t := &model.Task{
		ProjectId: projectId,
		GoVersion: go_ver_id,
		Branch:    param["branch"].(string),
		MainFile:  param["main_file"].(string),
		DestFile:  param["dest_file"].(string),
		DestOs:    destos,
		DestArch:  destArch,
		Env:       env,
	}
	err = model.InsertTask(t)
	if err != nil {
		log.Errorf("insert sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}
	writeSuccess(wr, "create task ok")
}

func ListTask(wr http.ResponseWriter, r *http.Request) {
	projectid, err := strconv.Atoi(r.FormValue("project_id"))
	if err != nil {
		log.Errorf("check param error:%s", err)
		projectid = 0
	}
	ts, err := model.ListTask(int64(projectid))
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}
	writeJson(wr, ts)
}

func StartTask(wr http.ResponseWriter, r *http.Request) {
	param, err := checkParam(r)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}

	var taskid int64
	if param["task_id"] != nil {
		taskid = int64(param["task_id"].(float64))
	}

	ts, err := model.ListTask(int64(taskid))
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	if len(ts) != 1 {
		log.Error("task select not 1")
		writeError(wr, "params error", "task id not ok")
		return
	}

	tl := &model.TaskLog{
		Id:          time.Now().UnixMilli(),
		TaskId:      int64(taskid),
		Status:      Init,
		Description: r.FormValue("description"),
	}

	err = model.InsertTaskLog(tl)
	if err != nil {
		log.Errorf("insert sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	go startTask(tl.TaskId, tl.Id)

	writeSuccess(wr, "start success")
}

func startTask(taskid, id int64) {
	t, err := model.GetTask(taskid)
	if err != nil {
		log.Error("get task error:%s", err)
		return
	}
	p, err := model.GetProject(t.ProjectId)
	if err != nil {
		log.Errorf("get project error:%s", err)
		return
	}
	g, err := model.GetGoVersion(t.GoVersion)
	if err != nil {
		log.Errorf("get version error:%s", err)
		return
	}

	gobin := path.Join(g.LocalPath, "bin/go")
	log.Debugf("go bin:%s", gobin)

	srcfile := path.Join(p.LocalPath, t.MainFile)
	log.Debugf("src file:%s", srcfile)

	destfile := path.Join(config.C.DestPath, p.Name, t.Branch, t.DestFile)
	log.Debugf("dest file:%s", destfile)

	c := exec.Command(gobin, "build", "-o", destfile, srcfile)
	c.Dir = p.LocalPath
	c.Env = append(c.Env, "GOBIN="+g.LocalPath)
	c.Env = append(c.Env, "GOPATH="+p.WorkSpace)
	c.Env = append(c.Env, "GOCACHE="+path.Join(p.WorkSpace, ".cache/"))
	c.Env = append(c.Env, strings.Split(p.Env, ";")...)
	c.Env = append(c.Env, strings.Split(t.Env, ";")...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c.Stdout = stdout
	c.Stderr = stderr
	c.Start()
	model.UpdateTaskLog(id, Running)
	c.Wait()
	outfilepath := path.Join(config.C.LogPath, p.Name, t.Branch, t.DestFile+".out.log")
	os.MkdirAll(filepath.Dir(outfilepath), os.ModePerm)
	outfile, err := os.OpenFile(outfilepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Errorf("openfile %s error", err)
		return
	}
	log.Debugf("task log id:%d out file:%s", id, outfilepath)
	io.Copy(outfile, stdout)
	model.UpdateTaskLogOut(id, outfilepath)
	if stderr.Len() > 0 {
		errfilepath := path.Join(config.C.LogPath, p.Name, t.Branch, t.DestFile+".err.log")
		os.MkdirAll(filepath.Dir(outfilepath), os.ModePerm)
		errfile, err := os.OpenFile(errfilepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			log.Errorf("openfile %s error", err)
			return
		}
		log.Debugf("task log id:%d err file:%s", id, errfilepath)
		io.Copy(errfile, stderr)
		model.UpdateTaskLogErr(id, errfilepath)
		model.UpdateTaskLog(id, Failed)
		return
	}
	ip, err := util.GetLocalIp()
	if err != nil {
		log.Errorf("get local ip error:%s", err)
		ip = "127.0.0.1"
	}
	url := fmt.Sprintf("http://%s:%d/output/%s/%s/%s", ip, config.C.Port, p.Name, t.Branch, t.DestFile)
	log.Debug("task log id:%d file url:%s", id, url)
	model.UpdateTaskLogUrl(id, url)
	model.UpdateTaskLog(id, Success)
	log.Infof("task log id:%d build success")
}

func ListTaskLog(wr http.ResponseWriter, r *http.Request) {
	taskid, err := strconv.Atoi(r.FormValue("task_id"))
	if err != nil {
		log.Errorf("check param error:%s", err)
		taskid = 0
	}
	ts, err := model.ListTaskLog(int64(taskid))
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}
	writeJson(wr, ts)
}
