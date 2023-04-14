package logic

import "testing"

func TestRep(t *testing.T) {
	version, err := parseUrl("go1.20.3.linux-amd64.tar.gz")
	t.Logf("ver:%s err:%s", version, err)
}

func TestDownlaod(t *testing.T) {
	file, err := downloadFile("https://go.dev/dl/go1.20.3.linux-amd64.tar.gz", "../goenv/", false)
	t.Logf("file:%s,err:%v", file, err)
}

func TestTar(t *testing.T) {
	err := tarFile("../goenv/go1.20.3.linux-amd64.tar.gz", "../goenv", false)
	t.Logf("err:%v", err)
}
