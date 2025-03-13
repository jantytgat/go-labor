package labor

import (
	"context"
	"time"
)

type Task interface {
	Handler() Handler
}

var (
	printTaskHandlerMaxConcurrent = 0
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

func printTaskHandlerFunc(ctx context.Context, t Task, data any) Process {
	// fmt.Println("PRINT", t, data)
	return Process{
		Output: Response{
			Data:  data,
			Error: nil,
		}}
}
