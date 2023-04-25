package util

import (
	"os/exec"
	"testing"
)

func TestGitPull(t *testing.T) {
	// PullBranch("../auto-build", "build", "megE6yPQ1qnr-y1dyL3k", "dev")
	err := Pull("../auto-build", "build", "dev")
	t.Log(err)
}

func TestIsGit(t *testing.T) {
	is := CheckIsGit("../addddd")
	t.Log(is)
}

func TestPipe(t *testing.T) {
	c1 := exec.Command("git", "remote")
	c1.Dir = "../"
	c2 := exec.Command("grep", "build")
	c2.Stdin, _ = c1.StdoutPipe()
	c3 := exec.Command("wc", "-l")
	c3.Stdin, _ = c2.StdoutPipe()

	c1.Run()
	c2.Run()
	l, _ := c3.CombinedOutput()
	t.Log(string(l))
}

func TestRemote(t *testing.T) {
	checkRemoteExist("./", "build")
}

func TestCheckout(t *testing.T) {
	err := Checkout("../auto-build", "master")
	t.Log(err)
}

func TestLog(t *testing.T) {
	ls, _ := GitLog("../", 10)
	for _, v := range ls {
		t.Logf("%s %s\n", v.Sha1, v.Commit)
	}
}
