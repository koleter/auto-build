package logic

import (
	"net/http"
	"path"
	"text/template"

	"github.com/hash-rabbit/auto-build/config"
)

func Index(w http.ResponseWriter, r *http.Request) {
	t1, err := template.ParseFiles(path.Join(config.C.WebPath, "index.html"))
	if err != nil {
		panic(err)
	}
	t1.Execute(w, nil)
}
