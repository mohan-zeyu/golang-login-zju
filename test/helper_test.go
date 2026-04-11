package test

import (
	"os"
	"strings"
	"strconv"
	"testing"
	"github.com/joho/godotenv"
)
func TestMain (m *testing.M) {
	_ = godotenv.Load("../.env")
	os.Exit(m.Run())
}

func envStringPtr(key string) *string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	return &v
}

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}

func mustSplitInts(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			panic(err)
		}
		out = append(out, n)
	}
	return out
}
