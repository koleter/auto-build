package logic

import (
	"errors"
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

	"github.com/hash-rabbit/auto-build/config"
	"github.com/hash-rabbit/auto-build/model"
	"github.com/hash-rabbit/auto-build/util"
	"github.com/subchen/go-log"
)

// build status
const (
	Init = iota
	Running
	Success
	Failed
)

func AddTask(wr http.ResponseWriter, r *http.Request) {
	t := new(model.Task)

	err := ParseParam(r, t)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}
	log.Debugf("recv:%+v", t)

	_, err = model.GetProject(t.ProjectId)
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	if _, err := model.GetGoVersion(t.GoVersion); err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	if len(t.Branch) == 0 {
		log.Errorf("check param error")
		writeError(wr, "params error", "param error")
		return
	}

	if len(t.MainFile) == 0 || len(t.DestFile) == 0 {
		log.Errorf("check param error")
		writeError(wr, "params error", "param error")
		return
	}

	switch t.DestOs {
	case "":
		t.DestOs = runtime.GOOS
	case "linux":
	case "windows":
	case "darwin":
	default:
		log.Errorf("check GOOS error")
		writeError(wr, "params error", "GOOS error")
		return
	}

	switch t.DestArch {
	case "":
		t.DestArch = runtime.GOARCH
	case "386":
	case "amd64":
	case "arm64":
	default:
		log.Errorf("check GOARCH error")
		writeError(wr, "params error", "GOARCH error")
		return
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
	ti, err := getTaskId(r)
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	tk, err := model.GetTask(ti)
	if err != nil {
		log.Error("get task error:%s", err)
		return
	}

	p, err := model.GetProject(tk.ProjectId)
	if err != nil {
		log.Errorf("get project error:%s", err)
		return
	}

	g, err := model.GetGoVersion(tk.GoVersion)
	if err != nil {
		log.Errorf("get version error:%s", err)
		return
	}

	tl := &model.TaskLog{
		Id:          time.Now().UnixMilli(),
		TaskId:      ti,
		Status:      Init,
		Description: r.FormValue("description"),
	}

	err = model.InsertTaskLog(tl)
	if err != nil {
		log.Errorf("insert sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	t := &task{
		id:    tl.Id,
		g:     g,
		p:     p,
		t:     tk,
		tl:    tl,
		files: make([]*os.File, 0),
	}

	go t.start()

	writeSuccess(wr, "start building...")
}

func getTaskId(r *http.Request) (int64, error) {
	param, err := checkParam(r)
	if err != nil {
		log.Errorf("check param error:%s", err)
		return 0, err
	}

	var taskid int64
	if param["task_id"] != nil {
		taskid = int64(param["task_id"].(float64))
	} else {
		log.Error("taskid is nil")
		return 0, errors.New("task_id not allowed")
	}

	ts, err := model.ListTask(int64(taskid))
	if err != nil {
		log.Errorf("select sql error:%s", err)
		return 0, err
	}

	if len(ts) != 1 {
		log.Error("task select not 1")
		return 0, err
	}
	return taskid, nil
}

type task struct {
	id int64
	g  *model.GoVersion
	p  *model.Project
	t  *model.Task
	tl *model.TaskLog

	files   []*os.File
	err_log *log.Logger
	out_log *log.Logger
	err     error
}

func (t *task) start() {
	defer t.clean()
	defer t.checkError()

	t.createOutFile()
	if t.err != nil {
		return
	}
	t.out_log.Info("create out put file success")

	t.createErrFile()
	if t.err != nil {
		return
	}
	t.out_log.Info("create out put error file success")

	t.err = util.Pull(t.p.LocalPath, defaultRemoteName, t.t.Branch)
	if t.err != nil {
		return
	}
	t.out_log.Info("git pull success")

	t.err = util.Checkout(t.p.LocalPath, t.t.Branch)
	if t.err != nil {
		return
	}
	t.out_log.Info("git checkout success")

	ls, err := util.GitLog(t.p.LocalPath, 1)
	if err != nil {
		t.err = err
		return
	}
	if len(ls) < 1 {
		t.err = errors.New("couldn't find git log")
		return
	}
	model.UpdateTaskLogDescription(t.id, ls[0].Commit)
	t.out_log.Info("git get commmit log success")

	gobin := path.Join(t.g.LocalPath, "bin/go")
	log.Debugf("go bin:%s", gobin)

	srcfile := path.Join(t.p.LocalPath, t.t.MainFile)
	log.Debugf("src file:%s", srcfile)

	destfile := path.Join(config.C.DestPath, t.p.Name, t.t.Branch, t.t.DestFile)
	log.Debugf("dest file:%s", destfile)

	c := exec.Command(gobin, "build", "-o", destfile, srcfile)
	c.Dir = t.p.LocalPath
	if t.p.GoMod {
		c.Env = append(c.Env, "GO111MODULE=on")
	} else {
		c.Env = append(c.Env, "GO111MODULE=off")
	}
	c.Env = append(c.Env, "GOBIN="+t.g.LocalPath)
	c.Env = append(c.Env, "GOPATH="+t.p.WorkSpace)
	c.Env = append(c.Env, "GOCACHE="+path.Join(t.p.WorkSpace, ".cache/"))
	c.Env = append(c.Env, "GOOS="+t.t.DestOs)
	c.Env = append(c.Env, "GOARCH="+t.t.DestArch)
	c.Env = append(c.Env, strings.Split(t.p.Env, ";")...)
	c.Env = append(c.Env, strings.Split(t.t.Env, ";")...)
	c.Stdout = t.out_log.Out
	c.Stderr = t.err_log.Out

	c.Start()
	model.UpdateTaskLog(t.id, Running)
	t.out_log.Info("start building")

	c.Wait()
	if c.Err != nil {
		t.err = c.Err
		return
	}
	t.out_log.Info("building finished")

	// TODO:查看是否输出文件,校验本地输出文件 sha2 和文件大小
	ip, err := util.GetLocalIp()
	if err != nil {
		log.Errorf("get local ip error:%s", err)
		ip = "127.0.0.1"
	}
	url := fmt.Sprintf("http://%s:%d/output/%s/%s/%s", ip, config.C.Port, t.p.Name, t.t.Branch, t.t.DestFile)
	log.Debugf("task log id:%d file url:%s", t.id, url)
	model.UpdateTaskLogUrl(t.id, url)

	t.out_log.Info("build success")
	log.Infof("task log id:%d build success", t.id)
}

func (t *task) createOutFile() {
	outfilepath := path.Join(config.C.LogPath, t.p.Name, t.t.Branch,
		fmt.Sprintf("%s.%d.out.log", t.t.DestFile, t.id))
	model.UpdateTaskLogOut(t.id, outfilepath)
	t.out_log, t.err = t.newLog(outfilepath)
}

func (t *task) createErrFile() {
	errfilepath := path.Join(config.C.LogPath, t.p.Name, t.t.Branch,
		fmt.Sprintf("%s.%d.err.log", t.t.DestFile, t.id))
	model.UpdateTaskLogErr(t.id, errfilepath)
	t.err_log, t.err = t.newLog(errfilepath)
}

func (t *task) newLog(filename string) (*log.Logger, error) {
	outfile, err := newLogFile(filename)
	if err != nil {
		return nil, err
	}
	t.files = append(t.files, outfile)

	l := log.New()
	l.Out = outfile

	return l, nil
}

func newLogFile(filename string) (*os.File, error) {
	os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func (t *task) checkError() {
	if t.err != nil {
		t.err_log.Errorf("get error:%s", t.err)
		t.out_log.Errorf("get error:%s", t.err)
		model.UpdateTaskLog(t.id, Failed)
	}
}

func (t *task) clean() {
	for _, v := range t.files {
		v.Close()
	}
}

// TODO:尝试编译的时候加锁
func tryLock() {

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
