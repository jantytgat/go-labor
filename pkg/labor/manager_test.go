package labor

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func benchmarkManager_AddJob(b *testing.B) int {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	mc := ManagerConfig{
		Address:         NewAddress(LocalAddress, "manager", "main"),
		EnableScheduler: false,
		EnableOperator:  false,
		EventLogger:     logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m := NewManager(mc)
	m.Start(ctx)

	var count int
	for n := 0; n < b.N; n++ {
		_ = m.AddJob("")
		count++
	}
	return count
}

func BenchmarkManager_AddJob1(b *testing.B) { benchmarkManager_AddJob(b) }
func BenchmarkManager_AddJob2(b *testing.B) { benchmarkManager_AddJob(b) }
func BenchmarkManager_AddJob3(b *testing.B) { benchmarkManager_AddJob(b) }
func BenchmarkManager_AddJob4(b *testing.B) { benchmarkManager_AddJob(b) }
func BenchmarkManager_AddJob5(b *testing.B) { benchmarkManager_AddJob(b) }
