package main

import (
	"context"
	"fmt"
	"github.com/jantytgat/go-labor/pkg/labor"
	"log/slog"
	"os"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	mc := labor.ManagerConfig{
		Address:         labor.NewAddress(labor.LocalAddress, "manager", "main"),
		EnableScheduler: false,
		EnableOperator:  false,
		EventLogger:     logger,
		EventLogLevel:   slog.LevelDebug,
		MaxOperators:    5000,
	}

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	m := labor.NewManager(mc)
	m.Start(ctx)

	c := labor.NewCustomer("main")

	sleep := false
	var requestCounter int
	go func(sleep bool) {
		for j := 0; j < 10000000000; j++ {
			if err := c.Send(
				ctx,
				labor.Request{
					Name: fmt.Sprintf("job_%d", j),
					Data: nil,
				},
				m); err != nil {
				break
			}
			requestCounter++
		}
	}(sleep)

	var responseCounter int
	go func(ctx context.Context, c *labor.Customer) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Println(c.Receive(ctx))
				responseCounter++
			}
		}
	}(ctx, c)
	time.Sleep(21 * time.Second)
	cancel()
	fmt.Printf("Processed jobs %d/%d in %s", responseCounter, requestCounter, time.Since(startTime))

}
