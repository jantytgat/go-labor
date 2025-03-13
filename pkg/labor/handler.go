package labor

import (
	"context"
	"time"
)

type Handler struct {
	execute       func(ctx context.Context, t Task, data any)
	maxConcurrent int
	timeout       time.Duration
}
