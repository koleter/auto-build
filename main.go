package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

	err = checkDir(config.C)
	if err != nil {
		log.Panicf("create dir error:%s", err)
	}

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
	r.HandleFunc("/api/porject/add", logic.AddPorject).Methods(http.MethodPost)
	r.HandleFunc("/api/porject/list", logic.ListPorject).Methods(http.MethodGet)
	r.HandleFunc("/api/goenv/add", logic.AddEnv).Methods(http.MethodPost)
	r.HandleFunc("/api/goenv/list", logic.ListEnv).Methods(http.MethodGet)
	r.HandleFunc("/api/task/add", logic.AddTask).Methods(http.MethodPost)
	r.HandleFunc("/api/task/list", logic.ListTask).Methods(http.MethodGet)
	r.HandleFunc("/api/task/start", logic.StartTask).Methods(http.MethodPost)
	r.HandleFunc("/api/task/log/list", logic.ListTaskLog).Methods(http.MethodGet)
	r.PathPrefix("/output/").Handler(http.StripPrefix("/output/", http.FileServer(http.Dir(c.DestPath))))

	return r
}

func checkDir(c *config.Config) error {
	c.LogPath, _ = filepath.Abs(c.LogPath)
	err := os.MkdirAll(c.LogPath, os.ModePerm)
	if err != nil {
		return err
	}

	c.DefaultGoPath, _ = filepath.Abs(c.DefaultGoPath)
	err = os.MkdirAll(c.DefaultGoPath, os.ModePerm)
	if err != nil {
		return err
	}

	c.DestPath, _ = filepath.Abs(c.DestPath)
	err = os.MkdirAll(c.DestPath, os.ModePerm)
	if err != nil {
		return err
	}

	c.GoEnvPath, _ = filepath.Abs(c.GoEnvPath)
	err = os.MkdirAll(c.GoEnvPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
