package util

import (
	"os/exec"

	"github.com/subchen/go-log"
)

func RunCmd(cmd *exec.Cmd) error {
	log.Infof("paht:%s cmd:%s", cmd.Dir, cmd.String())
	out, err := cmd.CombinedOutput()
	log.Infof("result:%s", string(out))
	return err
}
