package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	var pub = os.Getenv("AWSPUB")
	var secret = os.Getenv("AWSSecret")
	err := doRun(pub, secret, 10)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
