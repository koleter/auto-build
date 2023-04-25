package util

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestCmd(t *testing.T) {
	c := exec.Command("")
	fmt.Println(c.Dir)
	fmt.Println(c.Path)
}
