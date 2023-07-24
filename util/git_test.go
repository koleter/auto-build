package util

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func TestClone(t *testing.T) {
	expected := [][]string{
		// {"./wos-client", "https://igit.58corp.com/mengfanyu03/wos-client.git", "mengfanyu03", "Meng9826873201+"},
		// {"./wos-web", "https://igit.58corp.com/storage/wos-web.git", "TEFymne5jfWmTw5xmG7y"},
		// {"./scf-go", "https://igit.58corp.com/arch-scf/scf-go.git"},
		{"../test/dl", "https://github.com/golang/dl.git"},
	}

	for _, v := range expected {
		if len(v) >= 3 {
			t.Log(Clone(v[0], v[1], v[2:]...))
		} else {
			t.Log(Clone(v[0], v[1]))
		}
	}
}

func TestGitPull(t *testing.T) {
	expected := [][]string{
		{"./wos-client", "https://igit.58corp.com/mengfanyu03/wos-client.git", "dev-mfy"},
		{"./wos-web", "https://igit.58corp.com/storage/wos-web.git", "dev_ljh"},
		{"./scf-go", "https://igit.58corp.com/arch-scf/scf-go.git", "dev"},
	}

	for _, v := range expected {
		t.Log(Pull(v[0], "test", v[2]))
	}
}

func TestLog(t *testing.T) {
	expected := [][]string{
		{"./wos-client", "https://igit.58corp.com/mengfanyu03/wos-client.git", "dev-mfy"},
		{"./wos-web", "https://igit.58corp.com/storage/wos-web.git", "dev_ljh"},
		{"./scf-go", "https://igit.58corp.com/arch-scf/scf-go.git", "dev"},
	}
	for _, v := range expected {
		ls, _ := GitLog(v[0], 10)
		for _, l := range ls {
			t.Logf("project:%s %s %s\n", v[0], l.Sha1, l.Commit)
		}
	}

}

func TestGit(t *testing.T) {
	path := "./wtable-web"
	// url := "https://igit.58corp.com/storage/wtable-web.git"
	token := "vxfzUdXE11Rb25JnVYpW"
	// remote := "test"
	// branch := "feat-cmdb-cluster"

	// r, err := git.PlainClone(path, false, &git.CloneOptions{
	// 	URL:  url,
	// 	Auth: &http.BasicAuth{Username: "oauth2", Password: token},
	// })
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }

	r, err := git.PlainOpen(path)
	if err != nil {
		t.Error(err)
		return
	}

	re, err := r.Remote("origin")
	if err != nil {
		t.Error(err)
		return
	}

	re.Fetch(&git.FetchOptions{})

	list, err := re.List(&git.ListOptions{
		Auth: &http.BasicAuth{Username: "oauth2", Password: token},
	})
	if err != nil {
		t.Error(err)
		return
	}

	for _, l := range list {

		if l.Name().IsBranch() {
			t.Log(l)
			t.Log(l.Hash())
		}
		//  else {
		// 	// t.Log(l.Type())
		// }
	}

	tgs, err := r.Tags()
	if err != nil {
		t.Error(err)
		return
	}

	tgs.ForEach(func(r *plumbing.Reference) error {
		t.Log(r)
		return nil
	})

	// err = r.CreateBranch(&config.Branch{
	// 	Name:   "build_test",
	// 	Remote: "origin",
	// 	Merge:  plumbing.NewBranchReferenceName("master"),
	// })
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }

	// wt, _ := r.Worktree()
	// wt.Checkout(&git.CheckoutOptions{
	// 	Hash: ,
	// })
	// r.ResolveRevision()
	// r.BlobObjects()
	// r.BlobObject()
	// r.Fetch()

	// t.Log(r.CreateBranch(&config.Branch{
	// 	Name:   branch,
	// 	Remote: remote,
	// 	Merge:  plumbing.NewBranchReferenceName(branch),
	// }))

	// t.Log(wt.Checkout(&git.CheckoutOptions{
	// 	Branch: plumbing.NewBranchReferenceName(branch),
	// 	Create: true,
	// }))

	// t.Log(wt.Pull(&git.PullOptions{
	// 	RemoteName:    remote,
	// 	ReferenceName: plumbing.NewBranchReferenceName(branch),
	// }))

	// bs, _ := r.Branches()

	// bs.ForEach(func(r *plumbing.Reference) error {
	// 	t.Log(r.Name().Short())
	// 	return nil
	// })

	// t.Log(r.DeleteBranch(branch))

}
