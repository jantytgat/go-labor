package labor

import (
	"context"
	"fmt"
	"time"
)

type Task interface {
	Handler() Handler
}

var (
	printTaskHandlerMaxConcurrent = 100
	printTaskHandlerTimeout       = time.Second * 10
)

type PrintTask struct{}

func (t PrintTask) Handler() Handler {
	return Handler{
		execute:       printTaskHandlerFunc,
		maxConcurrent: printTaskHandlerMaxConcurrent,
		timeout:       printTaskHandlerTimeout,
	}
}

func printTaskHandlerFunc(ctx context.Context, t Task, data any) {
	fmt.Println("PRINT", t, data)
}
