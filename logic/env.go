package logic

import (
	"net/http"

	"github.com/hash-rabbit/auto-build/env"
	"github.com/subchen/go-log"
)

func ListEnv(wr http.ResponseWriter, r *http.Request) {
	envs, err := env.ListEnv()
	if err != nil {
		log.Errorf("list env error:%s", err)
		writeError(wr, "", "")
		return
	}

	writeJson(wr, envs)
}
