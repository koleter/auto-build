package logic

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/hash-rabbit/auto-build/config"
	"github.com/hash-rabbit/auto-build/model"
	"github.com/subchen/go-log"
)

var GoPkgUrlTemp string

func init() {
	GoPkgUrlTemp = `go([0-9]*\.[0-9]*[\.[0-9]*]?).` + runtime.GOOS + `-` + runtime.GOARCH + `.tar.gz`
}

func AddEnv(wr http.ResponseWriter, r *http.Request) {
	param, err := checkParam(r)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}
	log.Infof("add env param %+v", param)

	go addEnv(param["url"].(string), param["sha2"].(string))

	writeSuccess(wr, "start download go env...")
}

func addEnv(url, sha2 string) {
	version, err := parseUrl(url)
	if err != nil {
		log.Errorf("check url error:%s", err)
		return
	}
	log.Debugf("version: %s", version)

	envs, err := model.GoVersionList(version)
	if err != nil {
		log.Errorf("get version list error:%s", err)
		return
	}
	if len(envs) > 0 {
		log.Error("get version list has exist")
		return
	}

	file, err := downloadFile(url, config.C.GoEnvPath, false)
	if err != nil {
		log.Errorf("download file error:%s", err)
		return
	}
	log.Debugf("downlaod file %s success", file)

	if getFileSha2(file) != strings.ToLower(sha2) {
		log.Errorf("check file sha2 error,sha2:%s,local file:%s", sha2, getFileSha2(file))
		return
	}
	log.Debugf("check file %s sha2 success", file)

	err = tarFile(file, config.C.GoEnvPath, true)
	if err != nil {
		log.Errorf("tar file error:%s", err)
		return
	}

	err = os.Rename(path.Join(config.C.GoEnvPath, "go"), path.Join(config.C.GoEnvPath, "go"+version))
	if err != nil {
		log.Errorf("rename file error:%s", err)
		return
	}

	gv := &model.GoVersion{
		Version:   version,
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Url:       url,
		Sha2:      sha2,
		LocalPath: path.Join(config.C.GoEnvPath, "go"+version),
	}

	err = model.InsertGoVersion(gv)
	if err != nil {
		log.Errorf("insert sql error:%s", err)
		return
	}
	log.Debugf("insert into table success %+v", gv)
}

func parseUrl(url string) (string, error) {
	log.Debug(GoPkgUrlTemp)
	re, err := regexp.Compile(GoPkgUrlTemp)
	if err != nil {
		log.Errorf("regexp compile error:%s", err)
		return "", err
	}

	res := re.FindStringSubmatch(parseUrlFilename(url))
	if len(res) < 2 {
		return "", fmt.Errorf("couldn't match")
	}
	return res[1], nil
}

func downloadFile(url, destPath string, insertOnly bool) (string, error) {
	dest := parseUrlFilename(url)
	if fileExist(dest) {
		if insertOnly {
			return "", fmt.Errorf("file %s exist", dest)
		} else {
			os.Remove(dest)
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, err := os.Create(path.Join(destPath, dest))
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return path.Join(destPath, dest), nil
}

func fileExist(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func parseUrlFilename(url string) string {
	urlparam := strings.Split(strings.Split(url, "?")[0], "/")
	return urlparam[len(urlparam)-1]
}

func getFileSha2(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("open file error:%s", err)
		return ""
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Errorf("io copy error:%s", err)
		return ""
	}
	sum := hash.Sum(nil)

	return fmt.Sprintf("%x", sum)
}

func tarFile(file, dest string, deleteAfterTar bool) error {
	fr, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if h.FileInfo().IsDir() {
			os.MkdirAll(path.Join(dest, h.Name), os.ModePerm)
			continue
		}

		fw, err := os.OpenFile(path.Join(dest, h.Name), os.O_CREATE|os.O_WRONLY, os.FileMode(h.Mode))
		if err != nil {
			return err
		}
		defer fw.Close()

		_, err = io.Copy(fw, tr)
		if err != nil {
			return err
		}
	}

	log.Infof("unzip %s ok", file)

	if deleteAfterTar {
		os.Remove(file)
	}

	return nil
}

func ListEnv(wr http.ResponseWriter, r *http.Request) {
	envs, err := model.GoVersionList(r.FormValue("version"))
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}
	writeJson(wr, envs)
}

func DelEnv(wr http.ResponseWriter, r *http.Request) {
	e := &model.GoVersion{}
	err := ParseParam(r, e)
	if err != nil {
		log.Errorf("check param error:%s", err)
		writeError(wr, "params error", err.Error())
		return
	}

	goenv, err := model.GetGoVersion(e.Id)
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	ts, err := model.ListTask(0, goenv.Id)
	if err != nil {
		log.Errorf("select sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	log.Infof("%+v", ts)

	if len(ts) > 0 {
		writeError(wr, "logic error", "请先删除使用该环境的任务")
		return
	}

	err = model.DelGoVersion(goenv.Id)
	if err != nil {
		log.Errorf("delete sql error:%s", err)
		writeError(wr, "sql error", err.Error())
		return
	}

	err = os.RemoveAll(goenv.LocalPath)
	if err != nil {
		log.Errorf("os remove path %s error:%s", goenv.LocalPath, err)
		writeError(wr, "logic error", err.Error())
		return
	}

	writeSuccess(wr, "删除成功")
}
