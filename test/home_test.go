package test

import (
	"testing"
	"encoding/json"
	"fmt"
	"github.com/mohan-zeyu/golang-login-zju/zju"
	"time"
	"context"
	"os"
	"net/http"
)

func TestHome (t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	c := zju.NewCourses(username, password)
	ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Second)
	defer cancel()
	res, err := c.Fetch(ctx, "https://courses.zju.edu.cn/api/my-courses", 
	&zju.RequestOptions {
		Method: "POST",
		Headers: http.Header {
			"Content-Type": {"application/json"},
		},
	})

	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()
	var r map[string]any
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return
	}
	pretty, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		return
	}
	fmt.Println(string(pretty))
}
