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
	}

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	m := labor.NewManager(mc)
	m.Start(ctx)

	var count1 int
	var count2 int
	var count3 int
	sleep := false
	go func(sleep bool) {
		for j := 0; j < 10000000000; j++ {
			if m.Enabled() {
				count1++
				_ = m.AddJob(fmt.Sprintf("job_%d", j))
				if sleep {
					time.Sleep(1 * time.Second)
				}
				continue
			}
			break
		}
	}(sleep)

	go func(sleep bool) {
		for j := 0; j < 10000000000; j++ {
			if m.Enabled() {
				count2++
				_ = m.AddJob(fmt.Sprintf("job_%d", j))
				if sleep {
					time.Sleep(1 * time.Second)
				}
				continue
			}
			break
		}
	}(sleep)

	go func(sleep bool) {
		for j := 0; j < 10000000000; j++ {
			if m.Enabled() {
				count3++
				_ = m.AddJob(fmt.Sprintf("job_%d", j))
				if sleep {
					time.Sleep(1 * time.Second)
				}
				continue
			}
			break
		}
	}(sleep)
	time.Sleep(21 * time.Second)
	cancel()
	fmt.Println("Processed jobs", count1+count2+count3, time.Since(startTime))

}
