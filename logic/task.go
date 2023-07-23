package logic

import (
	"bytes"
	"errors"
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

	"github.com/hash-rabbit/auto-build/config"
	goenv "github.com/hash-rabbit/auto-build/env"
	"github.com/hash-rabbit/auto-build/model"
	"github.com/hash-rabbit/auto-build/util"
	"github.com/subchen/go-log"
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
		log.Debugf("check param error:%s", err)
		projectid = 0
	}
	ts, err := model.ListTask(int64(projectid), 0)
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
		log.Errorf("get task error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	p, err := model.GetProject(tk.ProjectId)
	if err != nil {
		log.Errorf("get project error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	// g, err := model.GetGoVersion(tk.GoVersion)
	// if err != nil {
	// 	log.Errorf("get version error:%s", err)
	// 	writeError(wr, "sql error", err.Error())
	// 	return
	// }

	tl := &model.TaskLog{
		TaskId: ti,
		Status: model.Init,
	}

	err = model.InsertTaskLog(tl)
	if err != nil {
		log.Errorf("insert sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	t := &task{
		id: tl.Id,
		// g:     g,
		goversion: tk.GoVersion,
		p:         p,
		t:         tk,
		tl:        tl,
		files:     make([]*os.File, 0),
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

	return taskid, nil
}

type task struct {
	id        int64
	goversion string
	p         *model.Project
	t         *model.Task
	tl        *model.TaskLog

	files   []*os.File
	out_log *log.Logger

	err error
}

func (t *task) start() {
	defer t.clean()
	defer t.checkError()
	defer t.checkoutMaster()

	log.Infof("star build task:%d", t.id)

	t.createOutFile()
	if t.err != nil {
		log.Error("create out file error")
		return
	}
	t.out_log.Info("create out put file success")

	t.out_log.Infof("git checkout %s", t.t.Branch)
	t.err = util.Checkout(t.p.LocalPath, defaultRemoteName, t.t.Branch)
	if t.err != nil {
		t.out_log.Error(t.err)
		log.Error(t.err)
		return
	}
	t.out_log.Info("git checkout success")

	t.out_log.Infof("git pull %s %s", defaultRemoteName, t.t.Branch)
	t.err = util.Pull(t.p.LocalPath, defaultRemoteName, t.t.Branch) //TODO:这块很容易卡住,需要排查
	if t.err != nil {
		t.out_log.Error(t.err)
		return
	}
	t.out_log.Info("git pull success")

	t.out_log.Infof("git log %s", t.t.Branch)
	ls, err := util.GitLog(t.p.LocalPath, 1)
	if err != nil {
		t.out_log.Error(err)
		t.err = err
		return
	}
	if len(ls) < 1 {
		t.out_log.Errorf("couldn't find git log")
		t.err = errors.New("couldn't find git log")
		return
	}
	model.UpdateTaskLogDescription(t.id, ls[0].Commit)
	t.out_log.Info("git get commmit log success")

	gobin := path.Join(goenv.GetGoPath(t.goversion), "bin/go")
	t.out_log.Infof("go bin:%s", gobin)

	srcfile := path.Join(t.p.LocalPath, t.t.MainFile)
	t.out_log.Infof("src file:%s", srcfile)

	destfile := path.Join(config.C.DestPath, t.p.Name, t.t.Branch, t.t.DestFile)
	t.out_log.Infof("dest file:%s", destfile)

	ospath := os.Getenv("PATH")
	env := make([]string, 0)
	if t.p.GoMod {
		env = append(env, "GO111MODULE=on")
	} else {
		env = append(env, "GO111MODULE=off")
	}
	env = append(env, "PATH=/opt/gcc-4.8.1/bin:"+ospath)
	env = append(env, "GOBIN="+goenv.GetGoPath(t.goversion))
	env = append(env, "GOPATH="+t.p.WorkSpace)
	env = append(env, "GOPROXY=https://goproxy.cn,direct")
	env = append(env, "GOCACHE="+path.Join(t.p.WorkSpace, ".cache/"))
	env = append(env, "GOOS="+t.t.DestOs)
	env = append(env, "GOARCH="+t.t.DestArch)
	env = append(env, strings.Split(t.p.Env, ";")...)
	env = append(env, strings.Split(t.t.Env, ";")...)

	goenv := exec.Command(gobin, "env")
	goenv.Dir = t.p.LocalPath
	out, err := goenv.CombinedOutput()
	if err != nil {
		t.out_log.Error(t.err)
		t.err = err
		return
	}
	t.out_log.Infof("go env:\n%s", string(out))

	// TODO
	// go get -insecure
	// var stderr bytes.Buffer
	// goget := exec.Command(gobin, "get", "-insecure", "./...")
	// goget.Dir = t.p.LocalPath
	// goget.Env = env
	// goget.Stdout = t.out_log.Out
	// goget.Stderr = &stderr
	// t.out_log.Info(goget.String())
	// err = goget.Run()
	// if err != nil {
	// 	t.out_log.Error(err)
	// 	t.err = err
	// 	return
	// }
	// if stderr.Len() > 0 {
	// 	t.out_log.Error(stderr.String())
	// 	t.err = errors.New(stderr.String())
	// 	return
	// }

	// go build
	var err_out bytes.Buffer
	c := exec.Command(gobin, "build", "-o", destfile, srcfile)
	c.Dir = t.p.LocalPath
	c.Env = env
	c.Stdout = t.out_log.Out
	c.Stderr = &err_out

	c.Start()
	model.UpdateTaskLog(t.id, model.Running)
	t.out_log.Info("start building")

	c.Wait()
	if c.Err != nil {
		t.err = c.Err
		return
	}
	if err_out.Len() > 0 {
		t.out_log.Error(err_out.String())
		t.err = errors.New(err_out.String())
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

	t.out_log.Infof("task log id:%d file url:%s", t.id, url)
}

func (t *task) createOutFile() {
	outfilepath := path.Join(config.C.RecordPath, t.p.Name, t.t.Branch,
		fmt.Sprintf("%s.%d.out.log", t.t.DestFile, t.id))
	model.UpdateTaskLogOut(t.id, outfilepath)
	t.out_log, t.err = t.newLog(outfilepath)
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
		log.Infof("build taskid:%d failed", t.id)
		model.UpdateTaskLog(t.id, model.Failed)
	} else {
		log.Infof("build taskid:%d success", t.id)
		model.UpdateTaskLog(t.id, model.Success)
	}
}

func (t *task) clean() {
	for _, v := range t.files {
		v.Close()
	}
}

func (t *task) checkoutMaster() {
	err := util.Checkout(t.p.LocalPath, defaultRemoteName, t.p.MainBranch)
	if err != nil {
		log.Error(err)
	}
}

// TODO:尝试编译的时候加锁
// git pull 总是卡住,先不加锁
func tryLock() {

}

func ListTaskLog(wr http.ResponseWriter, r *http.Request) {
	projectid, err := strconv.ParseInt(r.FormValue("project_id"), 10, 64)
	if err != nil {
		log.Warnf("check param error:%s", err)
		projectid = 0
	}

	versionid, err := strconv.ParseInt(r.FormValue("version_id"), 10, 64)
	if err != nil {
		log.Warnf("check param error:%s", err)
		versionid = 0
	}

	taskid, err := strconv.ParseInt(r.FormValue("task_id"), 10, 64)
	if err != nil {
		log.Warnf("check param error:%s", err)
		taskid = 0
	}

	limit, err := strconv.Atoi(r.FormValue("page_size"))
	if err != nil {
		log.Warnf("check param error:%s", err)
		limit = 20
	}

	offset, err := strconv.Atoi(r.FormValue("page_num"))
	if err != nil {
		log.Warnf("check param error:%s", err)
		offset = 0
	}

	ts, err := model.ListTaskLog(versionid, projectid, taskid, limit, offset)
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	writeJson(wr, ts)
}

func GetTaskLogOutput(wr http.ResponseWriter, r *http.Request) {
	recordid, err := strconv.Atoi(r.FormValue("task_log_id"))
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "check param error", err.Error())
		return
	}

	tl, err := model.GetTaskLog(int64(recordid))
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	f, err := os.Open(tl.OutFilePath)
	if err != nil {
		log.Errorf("logic error:%s", err)
		writeError(wr, "logic error", err.Error())
		return
	}

	data, err := io.ReadAll(f)
	if err != nil {
		log.Errorf("logic error:%s", err)
		writeError(wr, "logic error", err.Error())
		return
	}

	writeJson(wr, string(data))
}

func SetTaskAutoBuild(wr http.ResponseWriter, r *http.Request) {
	t := &model.Task{}
	err := ParseParam(r, t)
	if err != nil {
		log.Debugf("check param error:%s", err)
	}

	err = model.UpdateTaskAutoBuild(t.Id, t.AutoBuild)
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}
	writeSuccess(wr, "更新成功")
}

func DelTask(wr http.ResponseWriter, r *http.Request) {
	t := new(model.Task)
	err := ParseParam(r, t)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}

	_, err = model.GetTask(t.Id)
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	err = model.DelTask(t.Id)
	if err != nil {
		log.Errorf("delete sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	writeSuccess(wr, "删除成功")
}
