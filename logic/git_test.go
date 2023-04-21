package logic

import (
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestLog(t *testing.T) {
	r, err := git.PlainOpen("../")
	if err != nil {
		fmt.Print(err)
		return
	}

	l, err := r.Log(&git.LogOptions{})
	if err != nil {
		fmt.Print(err)
		return
	}
	err = l.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		return nil
	})
	if err != nil {
		fmt.Print(err)
		return
	}
}

func TestBranch(t *testing.T) {
	r, err := git.PlainOpen("../")
	if err != nil {
		fmt.Print(err)
		return
	}

	bs, err := r.Branches()
	if err != nil {
		fmt.Print(err)
		return
	}

	var h plumbing.Hash

	bs.ForEach(func(r *plumbing.Reference) error {
		fmt.Printf("%s %s \n", r.Name(), r.Hash().String())
		if r.Name() == "refs/heads/master" {
			h = r.Hash()
		}
		return nil
	})

	r.CommitObject(h)

	l, _ := r.Log(&git.LogOptions{
		From: h,
	})

	// w, _ := r.Worktree()
	// w.Clean(&git.CleanOptions{})
	// w.Checkout(&git.CheckoutOptions{
	// 	Branch: plumbing.ReferenceName("master"),
	// })

	// l, err := r.Log(&git.LogOptions{})
	// if err != nil {
	// 	fmt.Print(err)
	// 	return
	// }
	err = l.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		return nil
	})
	if err != nil {
		fmt.Print(err)
		return
	}
}
