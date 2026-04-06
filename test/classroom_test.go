package test

import (
	"fmt"
	"os"
	"testing"
	"github.com/mohan-zeyu/golang-login-zju/zju"
)

func TestClassroom (t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fmt.Println("Start testing for Classroom")
	c, err := zju.NewClassroom(username, password)
	c.Book(&zju.Message{
      ClassroomSelection: &zju.ClassroomSelection{
          PkeyList: []string{"05E69DBB8B3140AF9D506A8AFD8E9709"}, // 紫金港北4-203
          Xnm:      "2025",
          Xqm:      "16",
          XqhID:    "04C12561BA75AB98E0630CA6CA0A8F0F",
          Lxdh:     "YOUR_PHONE_NUMBER",
          Jyyydm:   "01",
          Xxyy:     "活动事项",
          Jysxqc:   "活动标题",
          Spbh:     nil,
          Jyfs:     2,
          Sqzt:     1,
          Jylx:     "1",
          Ksrq:     "2025-04-25",
          Jsrq:     "2025-04-25",
          ArrJcd:   []int{6,7,8,9,10},
      },
      InfoplusData: &zju.InfoplusData{
          Email:            "mohanzeyu@qq.com",
          BorrowReasonText: "教学",
          Identity:         "学生",
          DepartmentCode:   "521000",
          DepartmentName:   "计算机科学与技术学院",
          TimeSlotLabel:    "2025-04-25 ~ 2025-04-25 (第6-10节)",

          SupervisorCode:     "教职工号",
          SupervisorName:     "Supervisor name",
          SupervisorDept:     "DeptCode",
          SupervisorDeptName: "DeptName",
          SupervisorPhone:    "Supervisor Phone Number",
          SupervisorAttr:     `{"indepOrganize_Name":"党委宣传部（含网络信息办公室）","organize_Codes":"042000\n006000","organizeName":"党委学生工作部","o
  rganizeFiltered_Codes":"042000","indepOrganize_Code":"006000","userCodesFiltered":"0017573","positions":"042000:FACULTY:0017573\n006000:HDSQ_XCB:0017573
  ","indepOrganize_Names":"党委宣传部（含网络信息办公室）\n党委学生工作部","userCode":"0017573","formalPositions":"042000:FACULTY:0017573\n006000:HDSQ_XCB
  :0017573","organize_Names":"党委学生工作部\n党委宣传部（含网络信息办公室）","cloned":"","indepOrganize_Codes":"006000\n042000","userCodes":"0017573","in
  depOrganizeFiltered_Codes":"042000","organizeCode":"042000","organizeFiltered_Names":"党委学生工作部","indepOrganizeFiltered_Names":"党委学生工作部"}`,

          OrganizerCode:     "StudentID",
          OrganizerName:     "StudentName",
          OrganizerDept:     "StudentDeptCode",
          OrganizerDeptName: "StudentDeptName",
          OrganizerPhone:    "StudentPhoneNumber",
          OrganizerAttr:     "", // auto-built from above fields

          GuidingTeacherCode:     "TeaCode",
          GuidingTeacherName:     "teaname",
          GuidingTeacherDept:     "TeaDeptCode",
          GuidingTeacherDeptName: "TeaDeptName",
          GuidingTeacherPhone:    "TeaPhone",
          GuidingTeacherAttr:
  `{"indepOrganize_Name":"党委宣传部（含网络信息办公室）","organize_Codes":"042000\n006000","organizeName":"党委学生工作部","organizeFiltered_Codes":"0420
  00","indepOrganize_Code":"006000","userCodesFiltered":"0017573","positions":"042000:FACULTY:0017573\n006000:HDSQ_XCB:0017573","indepOrganize_Names":"党
  委宣传部（含网络信息办公室）\n党委学生工作部","userCode":"0017573","formalPositions":"042000:FACULTY:0017573\n006000:HDSQ_XCB:0017573","organize_Names":
  "党委学生工作部\n党委宣传部（含网络信息办公室）","cloned":"","indepOrganize_Codes":"006000\n042000","userCodes":"0017573","indepOrganizeFiltered_Codes":
  "042000","organizeCode":"042000","organizeFiltered_Names":"党委学生工作部","indepOrganizeFiltered_Names":"党委学生工作部"}`,
      },
  })
	if err != nil {
		t.Errorf("%s", err)
	}
}
