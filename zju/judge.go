package zju

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/mohan-zeyu/golang-login-zju/util"
	"github.com/tidwall/gjson"
)
const (
	authorizeURL = "https://alt.zju.edu.cn/ua/login?platform=WEB&target=%2FstudentEvaluationBackend%2Flist"
	getListURL = "https://alt.zju.edu.cn/dapi/v2/tes/evaluation_plan_service/page_my_todo_plan_course_list"
	getCourseURL = "https://alt.zju.edu.cn/dapi/v2/tes/evaluation_plan_service/find_plan_courses_by_user"
	createFormURL = "https://alt.zju.edu.cn/dapi/v2/autoform/document_service/insert_document"
	saveJudgeURL = "https://alt.zju.edu.cn/dapi/v2/tes/evaluation_plan_service/save_plan_courses_by_user"
)

var judgeInfo map[string]any = map[string]any {
			"oadpflA": 5,
			"cHfvSga": 5,
			"iuFCIOj": 5,
			"FtFBhMR": 5,
			"UuWdHvl": 5,
			"AskLXei": []map[string]string{
				{
					"value": "SqszCjdd",
					"label": "知识：系统地说明课程的核心知识或概念，以及它们之间的逻辑关系",
				},
				{
					"value": "StFnMwbR",
					"label": "方法：深层次掌握课程中的重要原理、方法或技能",
				},
				{
					"value": "SimUtEdu",
					"label": "应用：灵活地解释、分析或解决一些现实中的问题",
				},
				{
					"value": "StGGVERJ",
					"label": "思维：较大程度上获得思维拓展或逻辑思维提升",
				},
				{
					"value": "SzEGpaIQ",
					"label": "价值观：更深入地认识世界，并能够以积极的心态和视角看待",
				},
			},
			"nzLUDxN": []map[string]string{
				{
					"value": "SAhKrdxb",
					"label": "无以上需要",
				},
			},
			"MfKamDv": "SKNzmbSY",
			"FanwXXp": "SezdRflh",
		}

func (z *ZJUAM) Judge(ctx context.Context) {
	authorizeOpt := &RequestOptions {
		Method: "GET",
		Headers: http.Header{
			"referer": []string{"https://alt.zju.edu.cn/studentEvaluationBackend/list"},
		},
	}
	res, _ := z.Fetch(ctx, authorizeURL, authorizeOpt)
	body, _ := io.ReadAll(res.Body)
	loc := res.Header.Get("Location")
	res.Body.Close()

	u, _ := url.Parse(loc)
	q := u.Query()
	for q.Get("token") == "" {
		res, _ = z.Fetch(ctx, loc, nil)
		loc = res.Header.Get("Location")
		res.Body.Close()
		u, _ = url.Parse(loc)
		q = u.Query()
	}
	jwtToken := q.Get("token")

	listBody, _ := json.Marshal(map[string]int {
		"pageNum": 0,
		"pageSize": 20,
	})
	getListOpt := &RequestOptions{
		Method: "POST",
		Headers: http.Header{
			"Authorization": []string{"Bearer " + jwtToken},
			"content-type": []string{"application/json"},
		},
		Body: bytes.NewBuffer(listBody),
	}
	res, _ = z.Fetch(ctx, getListURL, getListOpt)
	body, _ = io.ReadAll(res.Body)
	res.Body.Close()

	list := gjson.GetBytes(body, "data.data").Array()
	for _, re := range(list) {
		id := re.Get("id").String()

		courseBody, _ := json.Marshal(map[string]string{
			"planCourseId": id,
		})
		courseOpt := &RequestOptions {
			Method: "POST",
			Headers: http.Header {
				"Authorization": []string{"Bearer " + jwtToken},
				"content-type": []string{"application/json"},
			},
			Body: bytes.NewBuffer(courseBody),
		}
		res, _ = z.Fetch(ctx, getCourseURL, courseOpt)
		body, _ = io.ReadAll(res.Body)
		res.Body.Close()

		groupId := gjson.GetBytes(body, "data.groupId").String()
		TeaList := gjson.GetBytes(body, "data.teacherList").Array()
		getFormOpt := *getListOpt
		formBody, _ := json.Marshal(map[string]any{
			"groupId": groupId,
			"value": judgeInfo,
		})
		getFormOpt.Body = bytes.NewBuffer(formBody)

		res, _ = z.Fetch(ctx, createFormURL, &getFormOpt)
		body, _ = io.ReadAll(res.Body)
		res.Body.Close()

		formId := gjson.GetBytes(body, "data").String()
		for _, re := range(TeaList) {
			sid := re.Get("userSid").String()
			saveOpt := getFormOpt
			saveBody, _ := json.Marshal(map[string]any {
				"planCourseId": id,
				"teaching": true,
				"formId": formId,
				"teaSid": sid,
			})
			saveOpt.Body = bytes.NewBuffer(saveBody)
			res, _ = z.Fetch(ctx, saveJudgeURL, &saveOpt)
			body, _ = io.ReadAll(res.Body)
			
			util.PrintBody(body)
			util.PrintJSON(res.Header)
			res.Body.Close()

			code := gjson.GetBytes(body, "code").Int()
			if code == 200 {
				fmt.Printf("Teacher[%s] Succeed!\n", sid)
			}
		}
	}
}
