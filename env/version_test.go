package env

import "testing"

func TestInstall(t *testing.T) {
	err := Install("../test/version", "go1.20.6")
	if err != nil {
		t.Error(err)
	}
}
