package logic

import (
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5"
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
