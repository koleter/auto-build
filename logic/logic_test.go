package logic

import (
	"net/url"
	"testing"
)

// func TestRep(t *testing.T) {
// 	version, err := parseUrl("go1.20.3.linux-amd64.tar.gz")
// 	t.Logf("ver:%s err:%s", version, err)
// }

// func TestDownlaod(t *testing.T) {
// 	file, err := downloadFile("https://go.dev/dl/go1.20.3.linux-amd64.tar.gz", "../goenv/", false)
// 	t.Logf("file:%s,err:%v", file, err)
// }

// func TestTar(t *testing.T) {
// 	err := tarFile("../goenv/go1.20.3.linux-amd64.tar.gz", "../goenv", false)
// 	t.Logf("err:%v", err)
// }

func TestUrl(t *testing.T) {
	exs := []string{
		"git@igit.58corp.com:mengfanyu03/auto-build-go.git",
		"https://igit.58corp.com/mengfanyu03/auto-build-go.git",
		"https://igit.58corp.com/mengfanyu03/auto-build-go.git?param=lll",
	}

	for _, v := range exs {
		u, err := url.Parse(v)
		if err != nil {
			continue
		}
		t.Log(u.EscapedPath())
	}

}
