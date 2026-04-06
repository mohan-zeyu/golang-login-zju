package test

import (
	"fmt"
	"testing"
	"os"
	"github.com/mohan-zeyu/golang-login-zju/zju"
	"context"
	"time"
)

func TestLoginWithRedirects (t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	fmt.Println("Start testing for Login")
	am := zju.NewZJUAM(username, password, zju.WithRedirectsEnabled())
	str, err := am.Login(ctx)
	if err == nil {
		t.Errorf("%s", err)
	}
	fmt.Println(str)
}
