package tasks

import (
	"context"
	"fmt"
	"github.com/jantytgat/go-labor/pkg/labor"
	"time"
)

var (
	PrintHandler = labor.Handler{
		Name:          "print",
		Execute:       printTaskHandlerFunc,
		MaxConcurrent: 0,
		Timeout:       time.Second * 10,
	}
)

func printTaskHandlerFunc(ctx context.Context, p labor.Pipeline) labor.Pipeline {
	return labor.Pipeline{
		Input:  p.Input,
		Output: fmt.Sprintf("PRINT - %s", p.Input),
	}
}
