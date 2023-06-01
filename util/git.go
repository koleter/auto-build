package util

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Clone git clone url to path,token is user token,use: Clone(path,url,token)
// see https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
// if use username:password,should input Clone(path,url,username,password)
// for public repo use: Clone(path,url)
func Clone(path, url string, token ...string) error {
	op := &git.CloneOptions{
		URL: url,
	}

	switch len(token) {
	case 1:
		op.Auth = &http.BasicAuth{
			Username: "oauth2",
			Password: token[0],
		}
	case 2:
		op.Auth = &http.BasicAuth{
			Username: token[0],
			Password: token[1],
		}
	}

	_, err := git.PlainClone(path, false, op)
	return err
}

// url如果是私有工程需要带秘钥
// url 是已经配好用户名密码的 url
// insertOnly: true:如果远程名称已存在,则返回 false:如果远程名称存在则更新 url
func AddRemote(path, name, url string, insertOnly bool) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	_, err = r.Remote(name)
	if err != nil && err != git.ErrRemoteNotFound {
		return err
	}

	// 如果存在先删除
	if err == nil {
		if insertOnly {
			return fmt.Errorf("remote exist but insert only")
		}
		r.DeleteRemote(name)
	}

	op := &config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	}

	re, err := r.CreateRemote(op)
	if err != nil {
		return err
	}

	return re.Fetch(&git.FetchOptions{})
}

func RmRemote(path, name string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	return r.DeleteRemote(name)
}

// 请确保目前在 branch 分支上,否则会自动进行合并 branch 到当前分支
func Pull(path, remote, branch string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	op := &git.PullOptions{
		RemoteName: remote,
	}

	if len(branch) > 0 {
		op.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	err = w.Pull(op)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

func Fetch(path, remote string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	re, err := r.Remote(remote)
	if err != nil {
		return err
	}

	err = re.Fetch(&git.FetchOptions{})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

func BranchList(path, remote string) ([]string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	re, err := r.Remote(remote)
	if err != nil {
		return nil, err
	}

	refs, err := re.List(&git.ListOptions{})
	if err != nil {
		return nil, err
	}

	resu := make([]string, 0)
	for _, re := range refs {
		if re.Name().IsBranch() {
			resu = append(resu, re.Name().Short())
		}
	}

	return resu, nil
}

// 目前仅支持 branch 模式
func Checkout(path, remote, branch string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	exist := checkBranchExist(r, branch)
	if !exist {
		r.CreateBranch(&config.Branch{
			Name:   branch,
			Remote: remote,
			Merge:  plumbing.NewBranchReferenceName(branch),
		})
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	return w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: !exist,
	})
}

type LogItem struct {
	Sha1   string
	Commit string
}

func GitLog(path string, n int) ([]*LogItem, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	op := &git.LogOptions{}

	list, err := r.Log(op)
	if err != nil {
		return nil, err
	}
	defer list.Close()

	resu := make([]*LogItem, 0)
	for i := 0; i < n; i++ {
		commit, err := list.Next()
		if err == io.EOF {
			return resu, nil
		}
		if err != nil {
			return nil, err
		}
		resu = append(resu, &LogItem{
			Sha1:   commit.Hash.String(),
			Commit: commit.Message,
		})
	}
	return resu, nil
}

// giturl 为 git 地址(http[s]),不用带秘钥
// 公共仓库,token 不填
// 单个秘钥:token 为秘钥
// 用户名密码:token 为username,passwoed 的字符串数组
func GetUrl(giturl string, token ...string) string {
	if len(token) == 0 { //公有库
		return giturl
	}

	user := "oauth2"
	password := ""
	switch len(token) {
	case 1:
		password = token[0]
	case 2:
		user = token[0]
		password = token[1]
	default:
		return ""
	}

	user = url.QueryEscape(user)
	password = url.QueryEscape(password)

	if strings.HasPrefix(giturl, "https://") {
		return fmt.Sprintf("https://%s:%s@%s", user, password, strings.TrimPrefix(giturl, "https://"))
	} else if strings.HasPrefix(giturl, "http://") {
		return fmt.Sprintf("http://%s:%s@%s", user, password, strings.TrimPrefix(giturl, "http://"))
	}

	return "" //出错返回空,后面会检测出来
}

func checkBranchExist(r *git.Repository, branch string) bool {
	bs, _ := r.Branches()
	defer bs.Close()

	for {
		item, err := bs.Next()
		if err != nil {
			break
		}
		if branch == item.Name().Short() {
			return true
		}
	}
	return false
}
