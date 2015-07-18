package main

import (
	"fmt"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	var pub = os.Getenv("AWSPUB")
	var secret = os.Getenv("AWSSecret")
	ukeys := getUsers()
	lenNum := 0
	for _, v := range ukeys {
		lenNum += len(v)
	}
	fmt.Println(len(ukeys), lenNum)
	return
	err := doRun(pub, secret, 10)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
