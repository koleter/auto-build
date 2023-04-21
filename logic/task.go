package logic

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gith "github.com/go-git/go-git/v5/plumbing/transport/http"
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

	writeSuccess(wr, "start building...")
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

	r, err := git.PlainOpen(p.LocalPath)
	if err != nil {
		log.Errorf("git open error:%s", err)
		return
	}
	w, err := r.Worktree()
	if err != nil {
		log.Errorf("git open worktree error:%s", err)
		return
	}
	err = w.Clean(&git.CleanOptions{
		Dir: true,
	})
	if err != nil {
		log.Errorf("git clean error:%s", err)
		return
	}
	err = w.Pull(&git.PullOptions{
		RemoteName: defaultRemoteName,
		Auth:       &gith.BasicAuth{Password: p.Token, Username: "auto-build"},
	})
	if err != nil {
		log.Errorf("git pull error:%s", err)
		return
	}
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(t.Branch),
	})
	if err != nil {
		log.Errorf("git checkout %s error:%s", t.Branch, err)
		return
	}

	ref, err := r.Reference(plumbing.NewBranchReferenceName(t.Branch), false)
	if err != nil {
		log.Errorf("git get ref %s error:%s", t.Branch, err)
		return
	}
	com, err := r.CommitObject(ref.Hash())
	if err != nil {
		log.Errorf("git get commit %s error:%s", t.Branch, err)
		return
	}
	model.UpdateTaskLogDescription(id, com.Message)

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
	c.Env = append(c.Env, "GOOS="+t.DestOs)
	c.Env = append(c.Env, "GOARCH="+t.DestArch)
	c.Env = append(c.Env, strings.Split(p.Env, ";")...)
	c.Env = append(c.Env, strings.Split(t.Env, ";")...)

	outfilename := fmt.Sprintf("%s.%d.out.log", t.DestFile, id)
	outfilepath := path.Join(config.C.LogPath, p.Name, t.Branch, outfilename)
	os.MkdirAll(filepath.Dir(outfilepath), os.ModePerm)
	outfile, err := os.OpenFile(outfilepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Errorf("openfile %s error", err)
		return
	}
	model.UpdateTaskLogOut(id, outfilepath)
	log.Debugf("task log id:%d out file:%s", id, outfilepath)

	errfilename := fmt.Sprintf("%s.%d.err.log", t.DestFile, id)
	errfilepath := path.Join(config.C.LogPath, p.Name, t.Branch, errfilename)
	os.MkdirAll(filepath.Dir(outfilepath), os.ModePerm)
	errfile, err := os.OpenFile(errfilepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Errorf("openfile %s error", err)
		return
	}
	model.UpdateTaskLogErr(id, errfilepath)
	log.Debugf("task log id:%d err file:%s", id, errfilepath)

	c.Stdout = outfile
	c.Stderr = errfile
	c.Start()
	model.UpdateTaskLog(id, Running)

	c.Wait()
	if c.Err != nil {
		model.UpdateTaskLog(id, Failed)
		return
	}

	ip, err := util.GetLocalIp()
	if err != nil {
		log.Errorf("get local ip error:%s", err)
		ip = "127.0.0.1"
	}
	url := fmt.Sprintf("http://%s:%d/output/%s/%s/%s", ip, config.C.Port, p.Name, t.Branch, t.DestFile)
	log.Debugf("task log id:%d file url:%s", id, url)
	model.UpdateTaskLogUrl(id, url)

	model.UpdateTaskLog(id, Success)
	log.Infof("task log id:%d build success", id)
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
