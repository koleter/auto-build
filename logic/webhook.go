package logic

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hash-rabbit/auto-build/model"
	"github.com/subchen/go-log"
)

var (
	KIND_PUSH = "push"
	KIND_TAG  = "tag_push"
)

type Event struct {
	ObjectKind string `json:"object_kind"`
	Ref        string `json:"ref"`
}

func DoWebHook(wr http.ResponseWriter, r *http.Request) {
	projectName := mux.Vars(r)["project"]
	if len(projectName) == 0 {
		log.Errorf("check project name error")
		writeError(wr, "params error", "check project name error")
		return
	}

	e := new(Event)
	err := ParseParam(r, e)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}
	log.Debugf("recv webhook:%+v", e)

	if e.ObjectKind != KIND_PUSH {
		log.Errorf("event kind not push:%s", e.ObjectKind)
		writeError(wr, "params error", "event kind not push")
		return
	}

	branch := getBranch(e.Ref)
	if len(branch) == 0 {
		log.Errorf("parse branch form refs error:%s", e.Ref)
		writeError(wr, "logic error", "couldn't parse branch")
		return
	}

	p, err := model.GetProjectByName(projectName)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}

	ts, err := model.ListTask(p.Id, 0)
	if err != nil {
		log.Errorf("get project error:%s", err)
		writeError(wr, "logic error", err.Error())
		return
	}

	for _, t := range ts {
		if t.Branch == branch && t.AutoBuild {
			go autobuild(t.Id)
		}
	}

	writeSuccess(wr, "success")
}

func getBranch(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return strings.TrimPrefix(ref, "refs/heads/")
	}
	return ""
}

func autobuild(taskid int64) {
	tk, err := model.GetTask(taskid)
	if err != nil {
		log.Errorf("get task error:%s", err)
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
		TaskId: taskid,
		Status: Init,
	}

	err = model.InsertTaskLog(tl)
	if err != nil {
		log.Errorf("insert sql error:%s", err)
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
}
