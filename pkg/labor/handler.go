package labor

import (
	"context"
	"time"
)

type Handler struct {
	execute       func(ctx context.Context, t Task, data any) Process
	maxConcurrent int
	timeout       time.Duration
}
