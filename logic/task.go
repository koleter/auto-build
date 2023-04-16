package logic

import (
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"igit.58corp.com/mengfanyu03/auto-build-go/config"
	"igit.58corp.com/mengfanyu03/auto-build-go/log"
	"igit.58corp.com/mengfanyu03/auto-build-go/model"
)

func AddTask(wr http.ResponseWriter, r *http.Request) {
	param, err := checkParam(r)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}

	projectId := param["project_id"].(int64)
	if _, err := model.GetProject(projectId); err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	go_ver_id := param["go_version_id"].(int64)
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

	destos := param["dest_os"].(string)
	if len(destos) == 0 {
		destos = runtime.GOOS
	}

	destArch := param["dest_arch"].(string)
	if len(destArch) == 0 {
		destArch = runtime.GOARCH
	}

	t := &model.Task{
		ProjectId: projectId,
		GoVersion: go_ver_id,
		Branch:    param["branch"].(string),
		MainFile:  param["main_file"].(string),
		DestFile:  param["dest_file"].(string),
		DestOs:    destos,
		DestArch:  destArch,
		Env:       param["env"].(string),
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
		writeError(wr, "params error", err.Error())
		return
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
	taskid, err := strconv.Atoi(r.FormValue("task_id"))
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
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
		Status:      0,
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
	log.Debug("go bin:%s", gobin)

	srcfile := path.Join(p.LocalPath, t.MainFile)
	log.Debug("src file:%s", srcfile)

	destfile := path.Join(config.C.DestPath, p.Name, t.Branch, t.DestFile)
	log.Debug("dest file:%s", destfile)

	c := exec.Command(gobin, "build", srcfile, "-o", destfile)
	c.Dir = p.LocalPath
	c.Env = append(c.Env, "GOPATH="+p.WorkSpace)
	c.Env = append(c.Env, strings.Split(p.Env, ";")...)
	c.Env = append(c.Env, strings.Split(t.Env, ";")...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Run()
}

func ListTaskLog(wr http.ResponseWriter, r *http.Request) {
	taskid, err := strconv.Atoi(r.FormValue("task_id"))
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}
	ts, err := model.ListTaskLog(int64(taskid))
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}
	writeJson(wr, ts)
}
