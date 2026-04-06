package test

import (
	"testing"
	"os"
	"github.com/mohan-zeyu/golang-login-zju/zju"
)

func TestHistory (t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	if username == "" || password == "" {
		t.Fatal("Missing params")
	}
	c, err := zju.NewClassroom(username, password)
	if err != nil {
		t.Error(err)
	}
	c.History()
}
