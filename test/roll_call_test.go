package test
import (
	"testing"
	"os"
	"github.com/mohan-zeyu/golang-login-zju/zju"
)
func Test_roll_call(t *testing.T) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	c := zju.NewCourses(username, password)
	c.AnswerNumberRollcall("188440", "1475")
}
