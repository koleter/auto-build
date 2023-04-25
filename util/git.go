package util

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hash-rabbit/auto-build/log"
)

func InitGit(path string) error {
	if CheckIsGit(path) {
		return nil
	}

	cmd := exec.Command("git", "init")

	return cmd.Run()
}

// url如果是私有工程需要带秘钥
// insertOnly: true:如果远程名称已存在,则返回 false:如果远程名称存在则更新 url
func AddRemote(path, name, url string, insertOnly bool) error {
	if !CheckIsGit(path) {
		return fmt.Errorf("path:%s not a git", path)
	}

	exist, err := checkRemoteExist(path, name)
	if err != nil {
		return err
	}

	if exist && insertOnly {
		return fmt.Errorf("remote name:%s has exist", name)
	}

	option := "add"
	if exist {
		option = "set"
	}
	cmd := exec.Command("git", "remote", option, name, url)
	cmd.Dir = path

	return cmd.Run()
}

func checkRemoteExist(path, name string) (bool, error) {
	cmd := exec.Command("git", "remote")
	cmd.Dir = path
	remotes, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	for _, str := range strings.Split(strings.TrimSpace(string(remotes)), "\n") {
		if name == str {
			return true, nil
		}
	}

	return false, nil
}

// 默认远程和本地分支名称一样
func Pull(path, remote, branch string) error {
	if !CheckIsGit(path) {
		return fmt.Errorf("path:%s not a git", path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "pull", remote, branch)
	cmd.Dir = path

	return cmd.Run()
}

func Checkout(path, name string) error {
	if !CheckIsGit(path) {
		return fmt.Errorf("path:%s not a git", path)
	}

	cmd := exec.Command("git", "checkout", name)
	cmd.Dir = path

	return cmd.Run()
}

func CheckIsGit(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	isgit, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(isgit)) == "true"
}

type LogItem struct {
	Sha1   string
	Commit string
}

func GitLog(path string, n int) ([]*LogItem, error) {
	if !CheckIsGit(path) {
		return nil, fmt.Errorf("path:%s not a git", path)
	}

	resu := make([]*LogItem, 0)

	cmd := exec.Command("git", "log", "--oneline", "-"+strconv.Itoa(n))
	cmd.Dir = path
	logs, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	for _, str := range strings.Split(strings.TrimSpace(string(logs)), "\n") {
		start, end := strings.Index(str, "("), strings.Index(str, ")")
		if start >= 0 && end >= 0 {
			str = str[:start] + str[end+1:]
		}

		params := strings.Split(str, " ")
		if len(params) < 2 {
			log.Errorf("log:%s parse error", str)
			continue
		}

		resu = append(resu, &LogItem{
			Sha1:   params[0],
			Commit: strings.Join(params[1:], ""),
		})
	}
	return resu, nil
}
