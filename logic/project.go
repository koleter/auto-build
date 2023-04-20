package logic

import (
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	gitc "github.com/go-git/go-git/v5/config"
	gith "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hash-rabbit/auto-build/config"
	"github.com/hash-rabbit/auto-build/log"
	"github.com/hash-rabbit/auto-build/model"
)

var defaultRemoteName = "build"

func AddPorject(wr http.ResponseWriter, r *http.Request) {
	param, err := checkParam(r)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "param error", err.Error())
		return
	}

	if len(param["name"].(string)) == 0 {
		writeError(wr, "param error", "project name length == 0")
		return
	}

	if ps, _ := model.ListProject(param["name"].(string)); len(ps) > 0 {
		writeError(wr, "param error", "project name has exist")
		return
	}

	_, err = url.Parse(param["url"].(string))
	if err != nil {
		log.Errorf("url %s parse error:%s", param["url"].(string), err)
		writeError(wr, "param error", "url parse error")
		return
	}

	var workspace string
	if param["gomod"].(bool) {
		workspace = config.C.DefaultGoPath
	} else if len(param["workspace"].(string)) > 0 {
		workspace, _ = filepath.Abs(param["workspace"].(string))
	} else {
		log.Errorf("workspace not set")
		writeError(wr, "git error", "must set workspace")
		return
	}

	pro_path, err := filepath.Abs(param["path"].(string))
	if err != nil {
		log.Errorf("path %s set error", param["path"].(string))
		writeError(wr, "path error", "path set error")
		return
	}

	repo, err := git.PlainOpen(pro_path)
	if err == git.ErrRepositoryNotExists {
		c := &git.CloneOptions{
			URL:        param["url"].(string),
			RemoteName: defaultRemoteName,
		}
		if len(param["token"].(string)) > 0 {
			c.Auth = &gith.BasicAuth{Password: param["token"].(string), Username: "auto-build"}
		}
		repo, err = git.PlainClone(pro_path, false, c)
	}
	if err != nil {
		log.Errorf("git open project error:%s", err)
		writeError(wr, "git error", err.Error())
		return
	}

	_, err = repo.Remote(defaultRemoteName)
	if err != nil {
		git.NewRemote(repo.Storer, &gitc.RemoteConfig{
			Name: defaultRemoteName,
			URLs: []string{param["url"].(string)},
		})
	}

	p := &model.Project{
		Name:      param["name"].(string),
		LocalPath: pro_path,
		Url:       param["url"].(string),
		Token:     param["token"].(string),
		GoMod:     param["gomod"].(bool),
		WorkSpace: workspace,
		Env:       param["env"].(string),
	}

	err = model.InsertProject(p)
	if err != nil {
		log.Errorf("insert sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	writeSuccess(wr, "add project ok")
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

	repo, err := git.PlainOpen(p.LocalPath)
	if err != nil {
		log.Errorf("open project error:%s", err)
		writeError(wr, "git error", err.Error())
		return
	}

	w, err := repo.Worktree()
	if err != nil {
		log.Errorf("open project worktree error:%s", err)
		writeError(wr, "git error", err.Error())
		return
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: defaultRemoteName,
		Auth:       &gith.BasicAuth{Password: p.Token, Username: "auto-build"},
	})
	if err != nil {
		log.Errorf("git pull error:%s", err)
		writeError(wr, "git error", err.Error())
		return
	}

	writeSuccess(wr, "git pull success")
}
