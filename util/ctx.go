package util

import (
	"context"
	"time"
)

func CreateCtx(opt ...int) (context.Context, context.CancelFunc) {
	var t int
	if len(opt) == 0 || opt == nil {
		t = 10
	} else {
		t = opt[0]
	}
	return context.WithTimeout(context.Background(), time.Duration(t) * time.Second)
}
