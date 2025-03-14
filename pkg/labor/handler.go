package labor

import (
	"context"
	"time"
)

type Handler struct {
	Name          string
	Execute       func(ctx context.Context, p Pipeline) Pipeline
	MaxConcurrent int
	Timeout       time.Duration
}
