package util

import (
	"io"
	"strings"

	"github.com/go-git/go-git/v5"
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

func CloneSingleBranch(path, url, branch, token string) error {
	op := &git.CloneOptions{
		URL:           url,
		Auth:          getAuth(token),
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Depth:         1,
		Tags:          git.NoTags,
	}

	_, err := git.PlainClone(path, false, op)
	return err
}

func CloenWithBare(path, url, token string) error {
	op := &git.CloneOptions{
		URL:          url,
		Auth:         getAuth(token),
		SingleBranch: false,
	}
	_, err := git.PlainClone(path, true, op)
	return err
}

func getAuth(tokenStr string) *http.BasicAuth {
	if len(tokenStr) == 0 {
		return nil
	}

	token := strings.Split(tokenStr, ":")
	switch len(token) {
	case 1:
		return &http.BasicAuth{
			Username: "oauth2",
			Password: token[0],
		}
	case 2:
		return &http.BasicAuth{
			Username: token[0],
			Password: token[1],
		}
	default:
		return nil
	}
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

func Fetch(path, remote, token string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	op := &git.FetchOptions{
		RemoteName: remote,
		Auth:       getAuth(token),
		Force:      true,
	}

	err = r.Fetch(op)
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
	resu := make([]string, 0)

	re, err := r.Remote(remote)
	if err != nil {
		return nil, err
	}

	refs, err := re.List(&git.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, re := range refs {
		if re.Name().IsBranch() {
			resu = append(resu, re.Name().Short())
		}
	}

	return resu, nil
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
