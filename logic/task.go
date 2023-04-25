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
	status := Success
	defer model.UpdateTaskLog(id, status)

	t, err := model.GetTask(taskid)
	if err != nil {
		log.Error("get task error:%s", err)
		status = Failed
		return
	}
	p, err := model.GetProject(t.ProjectId)
	if err != nil {
		log.Errorf("get project error:%s", err)
		status = Failed
		return
	}
	g, err := model.GetGoVersion(t.GoVersion)
	if err != nil {
		log.Errorf("get version error:%s", err)
		status = Failed
		return
	}

	err = util.Pull(p.LocalPath, defaultRemoteName, t.Branch)
	if err != nil {
		log.Errorf("get pull error:%s", err)
		status = Failed
		return
	}

	err = util.Checkout(p.LocalPath, t.Branch)
	if err != nil {
		log.Errorf("get pull error:%s", err)
		status = Failed
		return
	}

	ls, err := util.GitLog(p.LocalPath, 1)
	if err != nil {
		log.Errorf("get log error:%s", err)
		status = Failed
		return
	}
	if len(ls) < 1 {
		log.Errorf("get log 0")
		status = Failed
		return
	}
	// r, err := git.PlainOpen(p.LocalPath)
	// if err != nil {
	// 	log.Errorf("git open error:%s", err)
	// 	return
	// }

	// w, err := r.Worktree()
	// if err != nil {
	// 	log.Errorf("git open worktree error:%s", err)
	// 	return
	// }

	// err = w.Clean(&git.CleanOptions{
	// 	Dir: true,
	// })
	// if err != nil {
	// 	log.Errorf("git clean error:%s", err)
	// 	return
	// }

	// err = w.Pull(&git.PullOptions{
	// 	RemoteName:    defaultRemoteName,
	// 	Auth:          &gith.BasicAuth{Password: p.Token, Username: "auto-build"},
	// 	ReferenceName: plumbing.NewBranchReferenceName(t.Branch),
	// 	SingleBranch:  true,
	// })
	// if err == git.NoErrAlreadyUpToDate {
	// 	log.Debugf("git pull err:%s", err)
	// } else if err != nil {
	// 	log.Errorf("git pull error:%s", err)
	// 	return
	// }

	// err = w.Checkout(&git.CheckoutOptions{
	// 	Branch: plumbing.NewBranchReferenceName(t.Branch),
	// })
	// if err != nil {
	// 	log.Errorf("git checkout %s error:%s", t.Branch, err)
	// 	return
	// }

	// ref, err := r.Reference(plumbing.NewBranchReferenceName(t.Branch), false)
	// if err != nil {
	// 	log.Errorf("git get ref %s error:%s", t.Branch, err)
	// 	return
	// }
	// com, err := r.CommitObject(ref.Hash())
	// if err != nil {
	// 	log.Errorf("git get commit %s error:%s", t.Branch, err)
	// 	return
	// }
	model.UpdateTaskLogDescription(id, ls[0].Commit)

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
		status = Failed
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
		status = Failed
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
		status = Failed
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
