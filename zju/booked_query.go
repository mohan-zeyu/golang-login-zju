package zju

import (
	"encoding/json"
	"fmt"
	"time"
	"context"
	"net/http"
)

func (c *Classroom) History () {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	opt := &RequestOptions {
		Headers: http.Header {
			"Access-Token": {c.accessToken},
		},
	}
	res, err := c.am.Fetch(ctx, "https://jxzygl.zju.edu.cn/service-zypt/api/jsjy/jsjyjs/pagingJsjysq?page_num=1&page_size=10&xnm=2025&xqm=", opt)
	if err != nil {
		return 
	}
	defer res.Body.Close()
	var body map[string]any
	err = convertStreamToJSON(res.Body, &body)
	if err != nil {
		return 
	}
	nb, err := json.MarshalIndent(body, "", " ")
	fmt.Println(string(nb))
}
