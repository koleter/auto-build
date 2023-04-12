package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"igit.58corp.com/mengfanyu03/auto-build-go/config"
	"igit.58corp.com/mengfanyu03/auto-build-go/log"
	"igit.58corp.com/mengfanyu03/auto-build-go/logic"
	"igit.58corp.com/mengfanyu03/auto-build-go/model"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("usage: auto-build config.toml")
		os.Exit(0)
	}
	log.InitLogger()

	config.LoadConfig(os.Args[1])

	err := model.InitSqlLite(config.C.SqlFile)
	if err != nil {
		log.Panicf("init sql error:%s", err)
	}
	err = model.AuthMergeTable()
	if err != nil {
		log.Panicf("merge error:%s", err)
	}
	defer model.Close()

	srv := &http.Server{
		Handler:      route(config.C),
		Addr:         fmt.Sprintf(":%d", config.C.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Infof("start listen port:%d", config.C.Port)
	log.Fatal(srv.ListenAndServe())
}

func route(c *config.Config) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", nil)
	r.HandleFunc("/porject/add", logic.DoAddPorject).Methods(http.MethodPost)
	r.HandleFunc("/porject/list", logic.DoListPorject).Methods(http.MethodGet)
	r.HandleFunc("/goenv/add", logic.DoAddPorject).Methods(http.MethodPost)
	r.HandleFunc("/goenv/list", logic.DoListPorject).Methods(http.MethodGet)

	r.PathPrefix("/output/").Handler(http.StripPrefix("/output/", http.FileServer(http.Dir(c.DistPath))))

	return r
}
