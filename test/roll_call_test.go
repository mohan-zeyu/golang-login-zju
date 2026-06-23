package test

import (
	"os"
	"testing"

	"github.com/mohan-zeyu/golang-login-zju/zju"
)
func Test_roll_call(t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	c := zju.NewCourses(username, password)
	// for i := range(10000) {
	// 	s := strconv.Itoa(i)
	// 	c.AnswerNumberRollcall("190480", s)
	// }
	c.AnswerNumberRollcall("191002", "2025")
}
