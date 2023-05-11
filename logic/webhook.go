package logic

import (
	"net/http"

	"github.com/hash-rabbit/auto-build/model"
	"github.com/hash-rabbit/auto-build/util"
	"github.com/subchen/go-log"
)

var (
	KIND_PUSH = "push"
	KIND_TAG  = "tag_push"
)

type Event struct {
	ObjectKind string           `json:"object_kind"`
	Project    *ProjectPushInfo `json:"project"`
}

type ProjectPushInfo struct {
	Name string `json:"name"`
}

func DoWebHook(wr http.ResponseWriter, r *http.Request) {
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

	p, err := model.GetProjectByName(e.Project.Name)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}

	err = util.Pull(p.LocalPath, defaultRemoteName, "")
	if err != nil {
		log.Errorf("git pull error:%s", err)
		writeError(wr, "run error", err.Error())
		return
	}

	writeSuccess(wr, "success")
}
