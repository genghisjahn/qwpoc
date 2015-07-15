package main

import "testing"

func TestRun(t *testing.T) {
	err := makeRun("test", "test", 10)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
