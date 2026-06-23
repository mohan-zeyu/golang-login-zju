package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mohan-zeyu/golang-login-zju/zju"
)

func TestJudge(t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fmt.Println("Start testing for Classroom")
	c := zju.NewZJUAM(username, password, zju.WithRedirectsDisabled())
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	c.Judge(ctx)
}
