package zju

import (
	//"fmt"
	//"encoding/json"
	"context"
	"net/http"
)
type CourseInfo struct {

}

func (c *CoursesLogin) Course(ctx context.Context) ([]map[string]any, error) {
	res, err := c.Fetch(ctx, "https://courses.zju.edu.cn/api/my-courses", 
		&RequestOptions {
			Method: "POST",
			Headers: http.Header {
				"Content-Type": {"application/json"},
			},
		})
	if err != nil {
		return nil, &ClsrmErr {err: "failed to fetch course information"}
	}
	defer res.Body.Close()
	var r map[string][]map[string]any
	err = convertStreamToJSON(res.Body, &r)
	if err != nil {
		return nil, &ClsrmErr {err: "Invalid response body"}
	}
	courses := r["courses"]
	return courses, nil
	//for _, course := range courses {
	//	//这里可以加一个看看现在的时间的功能，然后就可以自动选取当前课程
	//	if course["semester"] == nil {
	//		continue
	//	}
	//	semester, ok := course["semester"].(map[string]any)
	//	if !ok {
	//		continue
	//	}
	//	id, ok := semester["id"].(float64)
	//	if !ok {
	//		continue
	//	}
	//	if id == 79 {
	//		fmt.Println(course["name"])
	//	}
	//}
}
