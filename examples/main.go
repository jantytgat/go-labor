package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/jantytgat/go-labor/pkg/labor"
)

var (
	logLevel         = slog.LevelDebug
	runTime      int = 1
	maxJobs      int = 100
	maxCustomers     = 1
	maxOperators int = runtime.NumCPU()
	sleep            = false
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	mc := labor.ManagerConfig{
		Address:         labor.NewAddress(labor.LocalAddress, "manager", "main"),
		EnableScheduler: false,
		EnableOperator:  false,
		EventLogger:     logger,
		EventLogLevel:   slog.LevelDebug,
		MaxOperators:    maxOperators,
	}

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runTime)*time.Second)
	m := labor.NewManager(mc)
	m.Start(ctx)

	var customers = make([]*labor.Customer, maxCustomers)
	var requestCounters = make([]int, maxCustomers)
	var responseCounters = make([]int, maxCustomers)

	for i := 0; i < maxCustomers; i++ {
		customer := labor.NewCustomer(fmt.Sprintf("customer_%d", i+1))
		customers[i] = customer

		go func(c *labor.Customer, id int) {
			for j := 0; j < maxJobs; j++ {
				if err := c.Send(
					ctx,
					labor.Request{
						Name: fmt.Sprintf("job_%d", j),
						Data: nil,
					},
					m); err != nil {
					break
				}
				requestCounters[id]++

				if sleep {
					time.Sleep(1 * time.Second)
				}
			}
		}(customer, i)

		go func(ctx context.Context, c *labor.Customer, id int) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if res := c.Receive(ctx); res != nil {
						responseCounters[id]++
					}
				}
			}
		}(ctx, customer, i)
	}

	time.Sleep(time.Duration(runTime+1) * time.Second)
	cancel()

	var requestCounter int
	for reqC := range requestCounters {
		requestCounter = requestCounter + reqC
	}

	var responseCounter int
	for resC := range responseCounters {
		responseCounter = responseCounter + resC
	}
	fmt.Printf("Processed jobs %d/%d in %s\n", responseCounter, requestCounter, time.Since(startTime))

}
