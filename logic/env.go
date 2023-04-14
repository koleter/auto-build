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

	"igit.58corp.com/mengfanyu03/auto-build-go/config"
	"igit.58corp.com/mengfanyu03/auto-build-go/log"
	"igit.58corp.com/mengfanyu03/auto-build-go/model"
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

	writeSuccess(wr, "start add go env")
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

	err = gv.Insert()
	if err != nil {
		log.Errorf("insert sql error:%s", err)
		return
	}
	log.Debugf("insert into table success %+v", gv)
}

func parseUrl(url string) (string, error) {
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

func downloadFile(url, distPath string, insertOnly bool) (string, error) {
	dist := parseUrlFilename(url)
	if fileExist(dist) {
		if insertOnly {
			return "", fmt.Errorf("file %s exist", dist)
		} else {
			os.Remove(dist)
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, err := os.Create(path.Join(distPath, dist))
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return path.Join(distPath, dist), nil
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

func tarFile(file, dist string, deleteAfterTar bool) error {
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
			os.MkdirAll(path.Join(dist, h.Name), os.ModePerm)
			continue
		}

		fw, err := os.OpenFile(path.Join(dist, h.Name), os.O_CREATE|os.O_WRONLY, os.FileMode(h.Mode))
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
