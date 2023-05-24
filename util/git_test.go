package util

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestClone(t *testing.T) {
	expected := [][]string{
		{"./wos-client", "https://igit.58corp.com/mengfanyu03/wos-client.git", "mengfanyu03", "Meng9826873201+"},
		{"./wos-web", "https://igit.58corp.com/storage/wos-web.git", "TEFymne5jfWmTw5xmG7y"},
		{"./scf-go", "https://igit.58corp.com/arch-scf/scf-go.git"},
	}

	for _, v := range expected {
		if len(v) >= 3 {
			t.Log(Clone(v[0], v[1], v[2:]...))
		} else {
			t.Log(Clone(v[0], v[1]))
		}
	}
}

func TestAddRemote(t *testing.T) {
	expected := [][]string{
		{"./wos-client", "https://igit.58corp.com/mengfanyu03/wos-client.git", "mengfanyu03", "Meng9826873201+"},
		{"./wos-web", "https://igit.58corp.com/storage/wos-web.git", "TEFymne5jfWmTw5xmG7y"},
		{"./scf-go", "https://igit.58corp.com/arch-scf/scf-go.git"},
	}

	for _, v := range expected {
		if len(v) >= 3 {
			t.Log(AddRemote(v[0], "test", GetUrl(v[1], v[2:]...), false))
		} else {
			t.Log(AddRemote(v[0], "test", GetUrl(v[1]), false))
		}
	}

	for _, v := range expected {
		r, _ := git.PlainOpen(v[0])
		re, _ := r.Remote("test")
		t.Logf("path:%s remote 'test' urls: %+v", v[0], re.Config().URLs)
	}
}

func TestRmRemote(t *testing.T) {
	expected := [][]string{
		{"./wos-client", "https://igit.58corp.com/mengfanyu03/wos-client.git", "mengfanyu03", "Meng9826873201+"},
		{"./wos-web", "https://igit.58corp.com/storage/wos-web.git", "TEFymne5jfWmTw5xmG7y"},
		{"./scf-go", "https://igit.58corp.com/arch-scf/scf-go.git"},
	}

	for _, v := range expected {
		t.Log(RmRemote(v[0], "test"))
	}

	for _, v := range expected {
		r, _ := git.PlainOpen(v[0])
		_, err := r.Remote("test")
		t.Logf("path:%s remote 'test' err: %s", v[0], err)
	}

}

func TestCheckout(t *testing.T) {
	expected := [][]string{
		{"./wos-client", "https://igit.58corp.com/mengfanyu03/wos-client.git", "dev-mfy"},
		{"./wos-web", "https://igit.58corp.com/storage/wos-web.git", "dev_ljh"},
		{"./scf-go", "https://igit.58corp.com/arch-scf/scf-go.git", "dev"},
	}

	for _, v := range expected {
		t.Log(Checkout(v[0], "test", v[2]))
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
	path := "./wos-web"
	// url := "https://igit.58corp.com/storage/wos-web.git"
	remote := "test"
	branch := "feat-cmdb-cluster"

	r, err := git.PlainOpen(path)
	if err != nil {
		t.Error(err)
		return
	}

	wt, _ := r.Worktree()

	// t.Log(r.CreateBranch(&config.Branch{
	// 	Name:   branch,
	// 	Remote: remote,
	// 	Merge:  plumbing.NewBranchReferenceName(branch),
	// }))

	// t.Log(wt.Checkout(&git.CheckoutOptions{
	// 	Branch: plumbing.NewBranchReferenceName(branch),
	// 	Create: true,
	// }))

	t.Log(wt.Pull(&git.PullOptions{
		RemoteName:    remote,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	}))

	// bs, _ := r.Branches()

	// bs.ForEach(func(r *plumbing.Reference) error {
	// 	t.Log(r.Name().Short())
	// 	return nil
	// })

	// t.Log(r.DeleteBranch(branch))

}
