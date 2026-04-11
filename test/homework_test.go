package test

import (
	"testing"
	"github.com/mohan-zeyu/golang-login-zju/zju"
	"os"
)

func TestHomework (t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	zju.Homework(username, password)
}
