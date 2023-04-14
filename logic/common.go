package logic

import (
	"encoding/json"
	"io"
	"net/http"

	"igit.58corp.com/mengfanyu03/auto-build-go/log"
)

func checkParam(r *http.Request) (map[string]interface{}, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("read body error:%s", err)
		return nil, err
	}
	param := make(map[string]interface{})
	err = json.Unmarshal(data, &param)
	if err != nil {
		log.Errorf("json unmarshal error:%s", err)
		return nil, err
	}
	return param, nil
}

type ResponseInfo struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func write500(w http.ResponseWriter, err error) {
	//w.WriteHeader(http.StatusInternalServerError)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func writeResponseInfo(w http.ResponseWriter, code, msg string, data interface{}) {
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Access-Control-Allow-Headers", "x-requested-with,content-type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	res := &ResponseInfo{code, msg, data}
	bytes, err := json.Marshal(res)
	if err != nil {
		write500(w, err)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		write500(w, err)
		return
	}
}

func writeSuccess(w http.ResponseWriter, msg string) {
	writeResponseInfo(w, "success", msg, nil)
}

func writeJson(w http.ResponseWriter, obj interface{}) {
	writeResponseInfo(w, "success", "success", obj)
}

func writeError(w http.ResponseWriter, code, msg string) {
	writeResponseInfo(w, code, msg, nil)
}
