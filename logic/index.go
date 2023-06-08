package logic

import (
	"net/http"
	"path"
	"text/template"

	"github.com/hash-rabbit/auto-build/config"
	"github.com/subchen/go-log"
)

func Index(w http.ResponseWriter, r *http.Request) {
	t1, err := template.ParseFiles(path.Join(config.C.WebPath, "index.html"))
	if err != nil {
		log.Error(err)
		return
	}
	t1.Execute(w, nil)
}
