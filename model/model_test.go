package model

import "testing"

func TestModel(t *testing.T) {
	err := InitSqlLite("./test.db")
	if err != nil {
		t.Error(err)
		return
	}
	err = AuthMergeTable()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestProject(t *testing.T) {
	TestModel(t)
	p, err := GetProject(1)
	t.Logf("project:%+v,err:%s", p, err)
}
