package test

import (
	"testing"
	"io"
	"fmt"
	"github.com/mohan-zeyu/golang-login-zju/zju"
	"time"
	"context"
	"os"
)

func TestCourses (t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	c := zju.NewCourses(username, password)
	ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
	defer cancel()
	res, err := c.Fetch(ctx, "https://courses.zju.edu.cn/api/radar/rollcalls", nil)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()
	a, _ := io.ReadAll(res.Body)
	str := string(a)
	fmt.Println(str)
}
