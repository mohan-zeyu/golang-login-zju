package test

import (
	"fmt"
	"strings"
	"os"
	"testing"
	"github.com/mohan-zeyu/golang-login-zju/zju"
)

func TestClassroom (t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fmt.Println("Start testing for Classroom")
	c, err := zju.NewClassroom(username, password)
	if err != nil {
		t.Error(err)
	}
	c.Book(&zju.Message{
		ClassroomSelection: &zju.ClassroomSelection{
			PkeyList: strings.Split(os.Getenv("CLASSROOM_PKEY_LIST"), ","),
			Xnm:      os.Getenv("CLASSROOM_XNM"),
			Xqm:      os.Getenv("CLASSROOM_XQM"),
			XqhID:    os.Getenv("CLASSROOM_XQH_ID"),
			Lxdh:     os.Getenv("CLASSROOM_LXDH"),
			Jyyydm:   os.Getenv("CLASSROOM_JYYYDM"),
			Xxyy:     os.Getenv("CLASSROOM_XXYY"),
			Jysxqc:   os.Getenv("CLASSROOM_JYSXQC"),
			Spbh:     envStringPtr("CLASSROOM_SPBH"),
			Jyfs:     mustAtoi(os.Getenv("CLASSROOM_JYFS")),
			Sqzt:     mustAtoi(os.Getenv("CLASSROOM_SQZT")),
			Jylx:     os.Getenv("CLASSROOM_JYLX"),
			Ksrq:     os.Getenv("CLASSROOM_KSRQ"),
			Jsrq:     os.Getenv("CLASSROOM_JSRQ"),
			ArrJcd:   mustSplitInts(os.Getenv("CLASSROOM_ARR_JCD")),
		},
		InfoplusData: &zju.InfoplusData{
			Email:            os.Getenv("INFOPLUS_EMAIL"),
			BorrowReasonText: os.Getenv("INFOPLUS_BORROW_REASON_TEXT"),
			Identity:         os.Getenv("INFOPLUS_IDENTITY"),
			DepartmentCode:   os.Getenv("INFOPLUS_DEPARTMENT_CODE"),
			DepartmentName:   os.Getenv("INFOPLUS_DEPARTMENT_NAME"),

			SupervisorCode:     os.Getenv("INFOPLUS_SUPERVISOR_CODE"),
			SupervisorName:     os.Getenv("INFOPLUS_SUPERVISOR_NAME"),
			SupervisorDept:     os.Getenv("INFOPLUS_SUPERVISOR_DEPT"),
			SupervisorDeptName: os.Getenv("INFOPLUS_SUPERVISOR_DEPT_NAME"),
			SupervisorPhone:    os.Getenv("INFOPLUS_SUPERVISOR_PHONE"),
			SupervisorAttr:     os.Getenv("INFOPLUS_SUPERVISOR_ATTR"),

			OrganizerCode:     os.Getenv("INFOPLUS_ORGANIZER_CODE"),
			OrganizerName:     os.Getenv("INFOPLUS_ORGANIZER_NAME"),
			OrganizerDept:     os.Getenv("INFOPLUS_ORGANIZER_DEPT"),
			OrganizerDeptName: os.Getenv("INFOPLUS_ORGANIZER_DEPT_NAME"),
			OrganizerPhone:    os.Getenv("INFOPLUS_ORGANIZER_PHONE"),
			OrganizerAttr:     os.Getenv("INFOPLUS_ORGANIZER_ATTR"),

			GuidingTeacherCode:     os.Getenv("INFOPLUS_GUIDING_TEACHER_CODE"),
			GuidingTeacherName:     os.Getenv("INFOPLUS_GUIDING_TEACHER_NAME"),
			GuidingTeacherDept:     os.Getenv("INFOPLUS_GUIDING_TEACHER_DEPT"),
			GuidingTeacherDeptName: os.Getenv("INFOPLUS_GUIDING_TEACHER_DEPT_NAME"),
			GuidingTeacherPhone:    os.Getenv("INFOPLUS_GUIDING_TEACHER_PHONE"),
			GuidingTeacherAttr:     os.Getenv("INFOPLUS_GUIDING_TEACHER_ATTR"),
		},
	})
	if err != nil {
		t.Errorf("%s", err)
	}
}
