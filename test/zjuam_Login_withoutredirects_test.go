package test

import (
	"fmt"
	"time"
	"testing"
	"os"
	"context"
	"github.com/mohan-zeyu/golang-login-zju/zju"
)

func TestLogin (t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	fmt.Println("Start testing for Login")
	am := zju.NewZJUAM(username, password, zju.WithRedirectsDisabled())
	start := time.Now()
	str, err := am.Login(ctx)
	fmt.Println(time.Since(start))
	if err != nil {
		t.Errorf("%s", err)
	}
	_, err = am.Fetch(ctx, "https://courses.zju.edu.cn", nil)
	if err != nil {
		t.Errorf("%s", err)
	}
	fmt.Println(str)
}
