package zju

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type ClsrmErr struct {
	err string
}
func (c *ClsrmErr) Error () string {
	return c.err
}
type PayLoad struct {
	Code int `json:"code"`
	Data string `json:"data"`
}

type ClassroomSelection struct {
	PkeyList []string `json:"pkey_list"`
	Xnm      string `json:"xnm"`
	Xqm      string `json:"xqm"`
	XqhID    string `json:"xqh_id"`
	Lxdh     string `json:"lxdh"`
	Jyyydm   string `json:"jyyydm"`
	Xxyy     string `json:"xxyy"`
	Jysxqc   string `json:"jysxqc"`
	Spbh     *string `json:"spbh"`
	Jyfs     int    `json:"jyfs"`
	Sqzt     int    `json:"sqzt"`
	Jylx     string `json:"jylx"`
	Ksrq     string `json:"ksrq"`
	Jsrq     string `json:"jsrq"`
	ArrJcd   []int `json:"arr_jcd"`
}
// InfoplusData holds all fields needed for the Phase 2 infoplus form submission
type InfoplusData struct {
	Email            string
	BorrowReasonText string // e.g. "教学"
	Identity         string // e.g. "学生"
	DepartmentCode   string
	DepartmentName   string
	TimeSlotLabel    string // e.g. "2026-03-21 ~ 2026-03-21 (第6-10节)"

	SupervisorCode     string
	SupervisorName     string
	SupervisorDept     string
	SupervisorDeptName string
	SupervisorPhone    string
	SupervisorAttr     string

	OrganizerCode     string
	OrganizerName     string
	OrganizerDept     string
	OrganizerDeptName string
	OrganizerPhone    string
	OrganizerAttr     string // auto-built if empty

	GuidingTeacherCode     string
	GuidingTeacherName     string
	GuidingTeacherDept     string
	GuidingTeacherDeptName string
	GuidingTeacherPhone    string
	GuidingTeacherAttr     string
}

type Message struct {
	ClassroomSelection *ClassroomSelection
	InfoplusData       *InfoplusData
}

// Classroom holds the authenticated session for classroom booking operations.
type Classroom struct {
	am          *ZJUAM
	username    string
	accessToken string
}

// NewClassroom authenticates and returns a ready-to-use Classroom session (Steps 1-3).
func NewClassroom(username, password string) (*Classroom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	am := NewZJUAM(username, password, WithRedirectsDisabled())
	fmt.Println("Step 1: Get CAS Service Ticket")
	const CAS_BASE = "https://zjuam.zju.edu.cn"
	const JXZYGL_SERVICE_URL = "https://jxzygl.zju.edu.cn/zypt/loading?redirect=/teacher/roomLoan"
	const JXZYGL_BASE = "https://jxzygl.zju.edu.cn"

	u, _ := url.Parse(CAS_BASE + "/cas/login")
	q := u.Query()
	q.Set("service", JXZYGL_SERVICE_URL)
	u.RawQuery = q.Encode()

	casLoginURL := u.String()
	casLocation, _ := am.Att(ctx, casLoginURL)
	//if err != nil {
	//	return &ClsrmErr { err: "Step 1: Access to casLogin pafe failed" }
	//}
	//defer casRes.Body.Close()
	//casLocation := casRes.Header.Get("Location")

	tcktMatch := regexp.MustCompile(`ticket=(ST-[^&]+)`)
	ticketSlice := tcktMatch.FindStringSubmatch(casLocation)
	if ticketSlice == nil {
		return nil, &ClsrmErr { err: "Step 1: Ticket not found" }
	}
	ticket := ticketSlice[1]
	fmt.Println(ticket)

	fmt.Println("Step 2: Exchange Ticket for Access-Token")
	tokenOrgn, err := url.Parse(JXZYGL_BASE + "/service-auth/api/authentication/cas/authenticationToken")
	if err != nil {
		return nil, &ClsrmErr { err: "Step 2: Failed to parse URL in" }
	}
	tokenURL := tokenOrgn.String()
	jsonData, err := json.Marshal( map[string]string {
		"code": ticket,
		"service": JXZYGL_SERVICE_URL,
	})
	if err != nil {
		return nil, &ClsrmErr { err: "Step 2: Failed to generate jsonData in" }
	}
	reqFull := &RequestOptions {
		Method: "POST",
		Headers: http.Header {
			"Content-Type": {"application/json;charset=UTF-8"},
			"Accept": {"application/json, text/plain, */*"},
			"X-Requested-With": {"XMLHttpRequest"},
			"Origin": {JXZYGL_BASE},
		},
		Body: bytes.NewBuffer(jsonData),
	}
	resp, err := am.Fetch(ctx, tokenURL, reqFull)
	if err != nil {
		fmt.Println(err.Error())
		return nil, &ClsrmErr { err: "Step 2: Failed to get accessToken"}
	}
	defer resp.Body.Close()

	var tokenData map[string]any
	err = convertStreamToJSON(resp.Body, &tokenData)
	fmt.Printf("tokenData is: %#v\n", tokenData)
	if err != nil {
		return nil, &ClsrmErr { err: "Step 2: WTF the tokenData is"}
	}
	var accessToken string
	if tokenData["code"] == "0" && tokenData["data"] != "" {
		var ok bool
		accessToken, ok = tokenData["data"].(string)
		if !ok {
			return nil, &ClsrmErr { err: "Step 2: FUCK TOKENDATA"}
		}
		fmt.Println("Step 2: Succeed getting accessToken", accessToken)
	} else {
		return nil, &ClsrmErr { err: "Step 2: Unexpected response"}
	}

	fmt.Println("Step 3: Switch to Student Role")
	roleURLValue, _ := url.Parse(JXZYGL_BASE + "/service-auth/api/authentication/authz/role/student")
	roleURL := roleURLValue.String()

	roleResponse, err := am.Fetch(ctx, roleURL, &RequestOptions {
		Headers: http.Header {
			"Accept": {"application/json", "text/plain", "*/*"},
			"Access-Token": {accessToken},
			"X-Requested-With": {"XMLHttpRequest"},
		},
	})
	if err != nil {
		return nil, &ClsrmErr { err: "Step 3: Access to Role failed"}
	}
	defer roleResponse.Body.Close()
	var roleData map[string]any
	err = convertStreamToJSON(roleResponse.Body, &roleData)
	if err != nil {
		return nil, &ClsrmErr { err: "Step 3: Invalid JSON response"}
	}
	if roleData["code"] != "0" {
		return nil, &ClsrmErr { err: "Step 3: Role switched failed" }
	}
	fmt.Println("Step 3: Role switched successfully")

	return &Classroom{am: am, username: username, accessToken: accessToken}, nil
}

// Book submits a classroom booking application (Steps 4-8).
func (c *Classroom) Book(msg *Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	const JXZYGL_BASE = "https://jxzygl.zju.edu.cn"
	const INFOPLUS_BASE = "https://one.zju.edu.cn"

	fmt.Println("Step 4: Submit Classroom Selection")
	submitURL := JXZYGL_BASE + "/service-zypt/api/jsjy/jsjyjs/zjJsjysq"
	body, err := json.Marshal(*msg.ClassroomSelection)
	fmt.Println(string(body))
	if err != nil {
		return &ClsrmErr { err: "Step 4: invalid msg.ClassroomSelection"}
	}
	opt := &RequestOptions {
		Method: "POST",
		Headers: http.Header {
			"Content-Type": {"application/json;charset=UTF-8"},
			"Accept": {"application/json, text/plain, */*"},
			"Access-Token": {c.accessToken},
			"X-Requested-With": {"XMLHttpRequest"},
			"Origin": {JXZYGL_BASE},
			"Referer": {JXZYGL_BASE + "/zypt/teacher/roomLoan"},
		},
		Body: bytes.NewBuffer(body),
	}
	res, err := c.am.Fetch(ctx, submitURL, opt)
	if err != nil {
		return &ClsrmErr { err: "Step 4: selecetion submition failed" }
	}
	defer res.Body.Close()
	var data map[string]any
	err = convertStreamToJSON(res.Body, &data)
	fmt.Printf("%#v\n", data)
	if err != nil {
		return &ClsrmErr { err: "Step 4: Invalid data returned" }
	}
	// Extract the nested response: data.data is a map containing "tstjurl"
	innerData, ok := data["data"].(map[string]any)
	if data["code"] != "0" || !ok {
		return &ClsrmErr{err: fmt.Sprintf("Step 4: Classroom selection failed: %v", data)}
	}
	formURL, ok := innerData["tstjurl"].(string)
	if !ok || formURL == "" {
		return &ClsrmErr{err: "Step 4: No tstjurl in response"}
	}
	fmt.Println("Step 4: Form URL:", formURL)

	// Extract stepId from the form URL (e.g. /form/123456/render -> "123456")
	stepMatch := regexp.MustCompile(`/form/(\d+)/render`)
	stepSlice := stepMatch.FindStringSubmatch(formURL)
	if stepSlice == nil {
		return &ClsrmErr{err: "Step 4: Could not extract stepId from form URL"}
	}
	stepId := stepSlice[1]
	fmt.Println("Step 4: Step ID:", stepId)

	// Phase 2 — Step 5+6: Follow redirect chain to reach the infoplus form page,
	// then extract the csrfToken from the rendered HTML.
	fmt.Println("Step 5: Follow redirect chain to fetch infoplus form")
	const MAX_REDIRECTS = 20

	var csrfToken string
	currentURL := formURL
	var html string
	for i := range MAX_REDIRECTS {
		chainRes, chainErr := c.am.Fetch(ctx, currentURL, nil)
		if chainErr != nil {
			fmt.Println(chainErr.Error())
			return &ClsrmErr{err: "Step 5: Failed during redirect chain"}
		}
		loc := chainRes.Header.Get("Location")
		fmt.Printf("  [%d] %d %s\n", i+1, chainRes.StatusCode, currentURL)
		if loc != "" {
			fmt.Printf("      → %s\n", loc)
		}

		if chainRes.StatusCode >= 300 && chainRes.StatusCode < 400 && loc != "" {
			chainRes.Body.Close()
			// Resolve relative redirects against the current host
			if strings.HasPrefix(loc, "/") {
				schemeEnd := strings.Index(currentURL, "://")
				hostEnd := strings.Index(currentURL[schemeEnd+3:], "/")
				if hostEnd >= 0 {
					currentURL = currentURL[:schemeEnd+3+hostEnd] + loc
				} else {
					currentURL = currentURL + loc
				}
			} else {
				currentURL = loc
			}
		} else {
			html = convertStreamToString(chainRes.Body)
			chainRes.Body.Close()
			fmt.Printf("  Final page: %d chars\n", len(html))
			break
		}
	}

	// Try several patterns to find csrfToken in the HTML
	csrfPatterns := []struct {
		name string
		re   *regexp.Regexp
	}{
		{"meta itemprop", regexp.MustCompile(`itemprop="csrfToken"[^>]*content="([^"]+)"`)},
		{"meta csrfToken", regexp.MustCompile(`csrfToken[^>]*content="([^"]+)"`)},
		{"csrfToken=value", regexp.MustCompile(`csrfToken\s*[=:]\s*["']([^"']+)["']`)},
	}
	for _, p := range csrfPatterns {
		m := p.re.FindStringSubmatch(html)
		if m != nil {
			csrfToken = m[1]
			fmt.Printf("  csrfToken found via \"%s\": %s\n", p.name, csrfToken)
			break
		}
	}
	if csrfToken == "" {
		return &ClsrmErr{err: "Step 5: Could not extract csrfToken from HTML"}
	}

	// Phase 2 — Step 7: listNextStepsUsers (pre-check before final submission)
	fmt.Println("Step 7: listNextStepsUsers")
	ip := msg.InfoplusData
	formDataMap := buildFormData(c.username, stepId, msg.ClassroomSelection, ip)
	formDataJSON, err := json.Marshal(formDataMap)
	if err != nil {
		return &ClsrmErr{err: "Step 7: Failed to marshal formData"}
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	randVal := fmt.Sprintf("%.14f", rand.Float64()*1000)
	referer := INFOPLUS_BASE + "/infoplus/form/" + stepId + "/render?theme=standard"

	listBody := url.Values{
		"stepId":      {stepId},
		"actionId":    {"35"},
		"formData":    {string(formDataJSON)},
		"timestamp":   {timestamp},
		"rand":        {randVal},
		"boundFields": {getBoundFields()},
		"csrfToken":   {csrfToken},
		"lang":        {"en"},
	}
	listRes, err := c.am.ReqOfRes(ctx, &RequestOptions{
		Method: "POST",
		Headers: http.Header{
			"Content-Type":     {"application/x-www-form-urlencoded; charset=UTF-8"},
			"Accept":           {"application/json, text/javascript, */*; q=0.01"},
			"X-Requested-With": {"XMLHttpRequest"},
			"Origin":           {INFOPLUS_BASE},
			"Referer":          {referer},
		},
		Body: strings.NewReader(listBody.Encode()),
	}, INFOPLUS_BASE+"/infoplus/interface/listNextStepsUsers")
	if err != nil {
		return &ClsrmErr{err: "Step 7: listNextStepsUsers request failed"}
	}
	var listResult map[string]any
	convertStreamToJSON(listRes.Body, &listResult)
	listRes.Body.Close()
	fmt.Printf("  listNextStepsUsers response: %v\n", listResult)

	// Phase 2 — Step 8: doAction (final submission that creates the booking)
	fmt.Println("Step 8: doAction (final submission)")
	timestamp = fmt.Sprintf("%d", time.Now().Unix())
	randVal = fmt.Sprintf("%.14f", rand.Float64()*1000)

	actionBody := url.Values{
		"actionId":    {"35"},
		"formData":    {string(formDataJSON)},
		"remark":      {""},
		"rand":        {randVal},
		"nextUsers":   {"{}"},
		"stepId":      {stepId},
		"timestamp":   {timestamp},
		"boundFields": {getBoundFields()},
		"csrfToken":   {csrfToken},
		"lang":        {"en"},
	}
	actionRes, err := c.am.ReqOfRes(ctx, &RequestOptions{
		Method: "POST",
		Headers: http.Header{
			"Content-Type":     {"application/x-www-form-urlencoded; charset=UTF-8"},
			"Accept":           {"application/json, text/javascript, */*; q=0.01"},
			"X-Requested-With": {"XMLHttpRequest"},
			"Origin":           {INFOPLUS_BASE},
			"Referer":          {referer},
		},
		Body: strings.NewReader(actionBody.Encode()),
	}, INFOPLUS_BASE+"/infoplus/interface/doAction")
	if err != nil {
		return &ClsrmErr{err: "Step 8: doAction request failed"}
	}
	var actionResult map[string]any
	convertStreamToJSON(actionRes.Body, &actionResult)
	actionRes.Body.Close()
	fmt.Printf("  doAction response: %v\n", actionResult)

	errno, _ := actionResult["errno"].(float64)
	if errno != 0 {
		return &ClsrmErr{err: fmt.Sprintf("Step 8: doAction failed: %v", actionResult)}
	}
	fmt.Println("  BOOKING SUBMITTED SUCCESSFULLY!")
	return nil
}

// buildFormData constructs the full infoplus form payload matching the server's expected schema.
func buildFormData(username, stepId string, cs *ClassroomSelection, ip *InfoplusData) map[string]any {
	now := time.Now()
	nowUnix := now.Unix()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayUnix := today.Unix()
	nowTime := now.Hour()*3600 + now.Minute()*60 + now.Second()

	// Auto-build organizer attr if not provided
	organizerAttr := ip.OrganizerAttr
	if organizerAttr == "" {
		attrBytes, _ := json.Marshal(map[string]string{
			"indepOrganize_Name":  ip.OrganizerDeptName,
			"organize_Codes":      ip.OrganizerDept,
			"organizeName":        ip.OrganizerDeptName,
			"indepOrganize_Code":  ip.OrganizerDept,
			"positions":           ip.OrganizerDept + ":UNDERGRADUATE:" + ip.OrganizerCode,
			"indepOrganize_Names": ip.OrganizerDeptName,
			"userCode":            ip.OrganizerCode,
			"formalPositions":     ip.OrganizerDept + ":UNDERGRADUATE:" + ip.OrganizerCode,
			"organize_Names":      ip.OrganizerDeptName,
			"indepOrganize_Codes": ip.OrganizerDept,
			"userCodes":           ip.OrganizerCode,
			"organizeCode":        ip.OrganizerDept,
		})
		organizerAttr = string(attrBytes)
	}

	// JSON attrs for department fields
	supervisorDeptAttr, _ := json.Marshal(map[string]string{
		"indepOrganize_Name": ip.SupervisorDeptName,
		"indepOrganize_Code": ip.SupervisorDept,
	})
	organizerDeptAttr, _ := json.Marshal(map[string]string{
		"indepOrganize_Name": ip.OrganizerDeptName,
		"indepOrganize_Code": ip.OrganizerDept,
	})
	guidingDeptAttr, _ := json.Marshal(map[string]string{
		"indepOrganize_Name": ip.GuidingTeacherDeptName,
		"indepOrganize_Code": ip.GuidingTeacherDept,
	})

	positions := ip.DepartmentCode + ":UNDERGRADUATE:" + username
	formURL := "https://one.zju.edu.cn/infoplus/form/" + stepId + "/render?theme=standard"

	return map[string]any{
		// ── System variables ──
		"_VAR_NOW_TIME":    fmt.Sprintf("%d", nowTime),
		"_VAR_NOW":         fmt.Sprintf("%d", nowUnix),
		"_VAR_TODAY":       fmt.Sprintf("%d", todayUnix),
		"_VAR_NOW_YEAR":    fmt.Sprintf("%d", now.Year()),
		"_VAR_NOW_MONTH":   fmt.Sprintf("%d", int(now.Month())),
		"_VAR_NOW_DAY":     fmt.Sprintf("%d", now.Day()),
		"_VAR_STEP_CODE":   "TXSQ",
		"_VAR_STEP_NUMBER": stepId,
		"_VAR_ENTRY_NUMBER": "",
		"_VAR_RELEASE":     "true",
		"_VAR_ADDR":        "",

		// ── Action account (current user) ──
		"_VAR_ACTION_ACCOUNT":              username,
		"_VAR_ACTION_ACCOUNT_FRIENDLY":     username,
		"_VAR_ACTION_REALNAME":             ip.OrganizerName,
		"_VAR_ACTION_PHONE":                cs.Lxdh,
		"_VAR_ACTION_EMAIL":                ip.Email,
		"_VAR_ACTION_USERCODES":            username,
		"_VAR_ACTION_ORGANIZE":             ip.DepartmentCode,
		"_VAR_ACTION_ORGANIZE_Name":        ip.DepartmentName,
		"_VAR_ACTION_ORGANIZES_Codes":      ip.DepartmentCode,
		"_VAR_ACTION_ORGANIZES_Names":      ip.DepartmentName,
		"_VAR_ACTION_INDEP_ORGANIZE":       ip.DepartmentCode,
		"_VAR_ACTION_INDEP_ORGANIZE_Name":  ip.DepartmentName,
		"_VAR_ACTION_INDEP_ORGANIZES_Codes": ip.DepartmentCode,
		"_VAR_ACTION_INDEP_ORGANIZES_Names": ip.DepartmentName,

		// ── Owner ──
		"_VAR_OWNER_ACCOUNT":          username,
		"_VAR_OWNER_ACCOUNT_FRIENDLY": username,
		"_VAR_OWNER_REALNAME":         ip.OrganizerName,
		"_VAR_OWNER_PHONE":            cs.Lxdh,
		"_VAR_OWNER_EMAIL":            ip.Email,
		"_VAR_OWNER_USERCODES":        username,
		"_VAR_OWNER_ORGANIZES_Codes":  ip.DepartmentCode,
		"_VAR_OWNER_ORGANIZES_Names":  ip.DepartmentName,

		// ── Execute context ──
		"_VAR_EXECUTE_ORGANIZE":             ip.DepartmentCode,
		"_VAR_EXECUTE_ORGANIZE_Name":        ip.DepartmentName,
		"_VAR_EXECUTE_ORGANIZES_Codes":      ip.DepartmentCode,
		"_VAR_EXECUTE_ORGANIZES_Names":      ip.DepartmentName,
		"_VAR_EXECUTE_INDEP_ORGANIZE":       ip.DepartmentCode,
		"_VAR_EXECUTE_INDEP_ORGANIZE_Name":  ip.DepartmentName,
		"_VAR_EXECUTE_INDEP_ORGANIZES_Codes": ip.DepartmentCode,
		"_VAR_EXECUTE_INDEP_ORGANIZES_Names": ip.DepartmentName,
		"_VAR_EXECUTE_USERCODES":            username,
		"_VAR_POSITIONS":                    positions,
		"_VAR_EXECUTE_POSITIONS":            positions,

		"_VAR_URL":        formURL,
		"_VAR_URL_Name":   formURL,
		"_VAR_URL_Attr":   `{"theme":"standard"}`,
		"_VAR_ENTRY_NAME": "_教室借用申请()",
		"_VAR_ENTRY_TAGS": "13-场馆申请",

		// ── Group/array fields ──
		"fieldZYXXZYID":  []any{},
		"groupZYXXList":  []any{0},
		"groupSHYJList":  []any{},

		// ── Application fields ──
		"fieldSQBH":  "",
		"fieldSQRQ":  nowUnix,
		"fieldGH":    username,
		"fieldXM":    username,
		"fieldXM_Name": ip.OrganizerName,
		"fieldLXDH":  cs.Lxdh,
		"fieldDW":    ip.DepartmentCode,
		"fieldDW_Name": ip.DepartmentName,
		"fieldDZYX":  ip.Email,
		"fieldJYYY":  ip.BorrowReasonText,
		"fieldXXLY":  cs.Xxyy,

		"fieldSFSF":      "2",
		"fieldSFSF_Name": "否",
		"fieldJYFY":      "",

		// ── 活动负责人 (supervisor) ──
		"fieldHDFZR":           ip.SupervisorCode,
		"fieldHDFZR_Name":      ip.SupervisorName,
		"fieldHDFZR_Attr":      ip.SupervisorAttr,
		"fieldHDFZRGH":         ip.SupervisorCode,
		"fieldHDFZRSZXY":       ip.SupervisorDept,
		"fieldHDFZRSZXY_Name":  ip.SupervisorDeptName,
		"fieldHDFZRSZXY_Attr":  string(supervisorDeptAttr),
		"fieldHDFZRSJHM":       ip.SupervisorPhone,

		// ── 活动组织者 (organizer) ──
		"fieldHDZZZ":           ip.OrganizerCode,
		"fieldHDZZZ_Name":      ip.OrganizerName,
		"fieldHDZZZ_Attr":      organizerAttr,
		"fieldHDZZZGH":         ip.OrganizerCode,
		"fieldHDZZZSZXY":       ip.OrganizerDept,
		"fieldHDZZZSZXY_Name":  ip.OrganizerDeptName,
		"fieldHDZZZSZXY_Attr":  string(organizerDeptAttr),
		"fieldHDZZZSJHM":       ip.OrganizerPhone,

		"fieldFZSF": ip.Identity,

		// ── 指导老师 (guiding teacher) ──
		"fieldZDLS":           ip.GuidingTeacherCode,
		"fieldZDLS_Name":      ip.GuidingTeacherName,
		"fieldZDLS_Attr":      ip.GuidingTeacherAttr,
		"fieldZDLSGH":         ip.GuidingTeacherCode,
		"fieldZDLSSZXY":       ip.GuidingTeacherDept,
		"fieldZDLSSZXY_Name":  ip.GuidingTeacherDeptName,
		"fieldZDLSSZXY_Attr":  string(guidingDeptAttr),
		"fieldZDLSSJHM":       ip.GuidingTeacherPhone,

		// ── Flags (all "否" / no) ──
		"fieldSFQTXY": "2", "fieldSFQTXY_Name": "否",
		"fieldQTXYRS": "", "fieldQTXYSJXY": "",
		"fieldSFXWRY": "2", "fieldSFXWRY_Name": "否",
		"fieldXWRYRS": "", "fieldXWRYXWDW": "",
		"fieldSFGAT": "2", "fieldSFGAT_Name": "否",
		"fieldGATRS": "", "fieldGATSSDQ": "",
		"fieldSFYWJ": "2", "fieldSFYWJ_Name": "否",
		"fieldWJRS": "", "fieldWJGJ": "",

		"fieldFJ": "",

		// ── Resource info ──
		"fieldZYXXMC":    []any{},
		"fieldZYXXYYSJ":  []any{ip.TimeSlotLabel},
		"fieldFZSFBXY":   "no",
		"fieldFZJSSZXY":  "",
		"fieldFZJS":      "",

		// ── Approval fields (left empty for new submission) ──
		"fieldYC1":        "",
		"fieldTYSHSHYJ":   "",
		"fieldTYSHSHR":    "",
		"fieldTYSHSHRQ":   "",
		"fieldGXSJ":       "",
	}
}

func getBoundFields() string {
	return strings.Join([]string{
		"fieldHDZZZSZXY", "fieldSFXWRY", "fieldGATSSDQ", "fieldXWRYRS",
		"fieldSFSF", "fieldTYSHSHR", "fieldTYSHSHYJ", "fieldHDZZZSJHM",
		"fieldYC1", "fieldHDZZZGH", "fieldZYXXYYSJ", "fieldFJ",
		"fieldHDFZRGH", "fieldWJGJ", "fieldFZSFBXY", "fieldZYXXMC",
		"fieldWJRS", "fieldSHRQ", "fieldDW", "fieldQTXYRS",
		"fieldJYFY", "fieldGATRS", "fieldFZJSSZXY", "fieldZDLS",
		"fieldFZSF", "fieldSHR", "fieldXM", "fieldSHYJ",
		"fieldJYYY", "fieldDZYX", "fieldHDFZR", "fieldTYSHSHRQ",
		"fieldSFYWJ", "fieldGXSJ", "fieldHDZZZ", "fieldXXLY",
		"fieldSFGAT", "fieldZDLSGH", "fieldZDLSSZXY", "fieldHDFZRSZXY",
		"fieldLXDH", "fieldSHRSJHM", "fieldGH", "fieldQTXYSJXY",
		"fieldFZJS", "fieldSQBH", "fieldXWRYXWDW", "fieldSHHJ",
		"fieldZDLSSJHM", "fieldHDFZRSJHM", "fieldSFQTXY", "fieldSQRQ",
	}, ",")
}
