package labor

import (
	"fmt"
)

type Job struct {
	Name     string
	Data     any
	Pipeline Pipeline
}

type Pipeline struct {
	Sequence []Task
}

type Task interface {
	Execute() func() error
}
type PrintTask struct{}

func (t PrintTask) Execute() func() error {
	return func() error {
		fmt.Println("executing task")
		return nil
	}
}
