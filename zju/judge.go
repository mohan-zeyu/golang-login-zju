package zju

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tidwall/gjson"
)
const (
	authorizeURL = "https://alt.zju.edu.cn/ua/login?platform=WEB&target=%2FstudentEvaluationBackend%2Flist"
	getListURL = "https://alt.zju.edu.cn/dapi/v2/tes/evaluation_plan_service/page_my_todo_plan_course_list"
	getCourseURL = "https://alt.zju.edu.cn/dapi/v2/tes/evaluation_plan_service/find_plan_courses_by_user"
	createFormURL = "https://alt.zju.edu.cn/dapi/v2/autoform/document_service/insert_document"
	saveJudgeURL = "https://alt.zju.edu.cn/dapi/v2/tes/evaluation_plan_service/save_plan_courses_by_user"
)

var judgeInfo = map[string]any {
	"oadpflA": 5,
	"cHfvSga": 5,
	"iuFCIOj": 5,
	"FtFBhMR": 5,
	"UuWdHvl": 5,
	"AskLXei": []map[string]string{
		{ "value": "SqszCjdd", "label": "知识：系统地说明课程的核心知识或概念，以及它们之间的逻辑关系"},
		{ "value": "StFnMwbR", "label": "方法：深层次掌握课程中的重要原理、方法或技能"},
		{ "value": "SimUtEdu", "label": "应用：灵活地解释、分析或解决一些现实中的问题"},
		{ "value": "StGGVERJ", "label": "思维：较大程度上获得思维拓展或逻辑思维提升"},
		{ "value": "SzEGpaIQ", "label": "价值观：更深入地认识世界，并能够以积极的心态和视角看待"},
	},
	"nzLUDxN": []map[string]string{ 
		{"value": "SAhKrdxb", "label": "无以上需要"},
	},
	"MfKamDv": "SKNzmbSY",
	"FanwXXp": "SezdRflh",
}

// post issues one POST with a Fresh header map every call (no shared/ liased map, so cookies can't accumulate, and surfaces every error.
// Note that maps in Go are referece like things.
// Structs with it only does a shallow copy and share the hash map
func (z *ZJUAM) post(ctx context.Context, reqURL, jwt string, payload any) ([]byte, error) {
	body, _ := json.Marshal(payload)
	res, err := z.Fetch(ctx, reqURL, &RequestOptions{
		Method: "POST",
		Headers: http.Header {
			"Authorization": []string{"Bearer " + jwt},
			"content-type": []string{"application/json"},
		},
		Body: bytes.NewBuffer(body),
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func (z *ZJUAM) getToken(ctx context.Context) (string, error) {
	res, err := z.Fetch(ctx, authorizeURL, &RequestOptions{
		Method:  "GET",
		Headers: http.Header{"referer": []string{"https://alt.zju.edu.cn/studentEvaluationBackend/list"}},
	})
	if err != nil {
		return "", fmt.Errorf("authorize request failed: %w", err)
	}
	loc := res.Header.Get("Location")
	res.Body.Close()

	for hop := range(20) {
		u, err := url.Parse(loc)
		if err != nil {
			return "", fmt.Errorf("bad redirect URL %q: %w", loc, err)
		}
		if tok := u.Query().Get("token"); tok != "" {
			return tok, nil
		}
		res, err = z.Fetch(ctx, loc, nil)
		if err != nil {
			return "", fmt.Errorf("redirect hop %d failed: %w", hop, err)
		}
		loc = res.Header.Get("Location")
		res.Body.Close()
	}
	return "", fmt.Errorf("token not found after 20 redirects")
}

func (z *ZJUAM) Judge(ctx context.Context) {
	jwtToken, err := z.getToken(ctx)
	if err != nil {
		fmt.Println("token err:", err)
		return
	}
	listBody, err := z.post(ctx, getListURL, jwtToken, map[string]int {
		"pageNum": 0,
		"pageSize": 20,
	})
	if err != nil {
		fmt.Println("authroize request failed:", err)
		return
	}

	list := gjson.GetBytes(listBody, "data.data").Array()
	fmt.Printf("Found %d courses\n", len(list))
	ok, fail := 0, 0
	for i, re := range(list) {
		id := re.Get("id").String()

		courseBody, err := z.post(ctx, getCourseURL, jwtToken, map[string]string{ "planCourseId": id, })
		if err != nil {
			fmt.Printf("[course %d] get course info failed: %v\n", i, err)
			continue
		}

		groupId := gjson.GetBytes(courseBody, "data.groupId").String()
		TeaList := gjson.GetBytes(courseBody, "data.teacherList").Array()
		formBody, err := z.post(ctx, createFormURL, jwtToken, map[string]any{ "groupId": groupId, "value": judgeInfo, })
		if err != nil {
			fmt.Printf("[course %d] create form failed: %v\n", i, err)
			continue
		}

		formId := gjson.GetBytes(formBody, "data").String()
		fmt.Printf("[course %d] id=%s  %d teachers formId=%s\n", i, id, len(TeaList), formId)
		for ti, re := range(TeaList) {
			sid := re.Get("userSid").String()
			saveBody, err := z.post(ctx, saveJudgeURL, jwtToken, map[string]any {
				"planCourseId": id,
				"teaching": true,
				"formId": formId,
				"teaSid": sid,
			})
			if err != nil {
				fail++
				fmt.Printf("	teacher[%d] %s -> submit form failed: %v\n", ti, sid, err)
				continue
			}
			
			code := gjson.GetBytes(saveBody, "code").Int()
			if code == 200 {
				ok++
				fmt.Printf("	teacher[%d] %s Succeed!\n", ti, sid)
			} else {
				fail++

				fmt.Printf("	teacher[%d] %s -> FAILED code=%d msg=%q\n        body=%s\n", ti, 
				sid,
				code, 
				gjson.GetBytes(saveBody, "message").String(), string(saveBody))
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
	fmt.Printf("\nDone. %d succeeded, %d failed.\n", ok, fail)
}
