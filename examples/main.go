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
	logLevel         = slog.LevelInfo
	runTime      int = 1
	managerName      = "example"
	maxJobs      int = 10000
	maxCustomers     = 12
	maxOperators int = runtime.NumCPU() * maxCustomers
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	mc := labor.ManagerConfig{
		Address:                labor.NewAddress(labor.LocalAddress, "manager", managerName),
		ManagerEventLogLevel:   slog.LevelInfo,
		RouterEventLogLevel:    slog.LevelDebug,
		SchedulerEventLogLevel: slog.LevelDebug,
		OperatorEventLogLevel:  slog.LevelDebug,
		MaxOperators:           maxOperators,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runTime)*time.Second)
	m := labor.NewManager(ctx, mc, logger)
	m.Enable(ctx)

	var customers = make([]*labor.Customer, maxCustomers)

	for i := 0; i < maxCustomers; i++ {
		customer := labor.NewCustomer(fmt.Sprintf("customer_%d", i+1))
		customers[i] = customer

		go func(ctx context.Context, c *labor.Customer, id int) {
			for j := 0; j < maxJobs; j++ {
				if err := c.Send(
					ctx,
					labor.Request{
						Name: fmt.Sprintf("%s_job_%d", customer.Name, j+1),
						Data: nil,
					},
					m); err != nil {
					fmt.Println(err)
					break
				}
			}
		}(ctx, customer, i)

		go func(ctx context.Context, c *labor.Customer, id int) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if res := c.Receive(ctx); res != nil {
						logger.Log(ctx, slog.LevelInfo, "received reply", slog.Any("reply", res))
					}
				}
			}
		}(context.Background(), customer, i)
	}

	time.Sleep(time.Duration(runTime+1) * time.Second)
	var requests int
	var responses int

	for i := 0; i < maxCustomers; i++ {
		requests = requests + customers[i].Requests
		responses = responses + customers[i].Responses
	}
	cancel()
	fmt.Println("Processed requests:", requests)
	fmt.Println("Processed responses:", responses)
}
