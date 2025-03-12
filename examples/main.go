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
	runTime      int = 10
	managerName      = "example"
	maxJobs      int = 1000000
	maxCustomers     = 2000
	maxOperators int = runtime.NumCPU() * maxCustomers * 2
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	mc := labor.ManagerConfig{
		Address:       labor.NewAddress(labor.LocalAddress, "manager", managerName),
		EventLogger:   logger,
		EventLogLevel: logLevel,
		MaxOperators:  maxOperators,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runTime)*time.Second)
	m := labor.NewManager(mc)
	m.Enable(ctx)

	var customers = make([]*labor.Customer, maxCustomers)

	for i := 0; i < maxCustomers; i++ {
		customer := labor.NewCustomer(fmt.Sprintf("customer_%d", i+1))
		customers[i] = customer

		go func(ctx context.Context, c *labor.Customer, id int) {
			for j := 0; j < maxJobs; j++ {
				if err := c.Send(
					ctx,
					labor.Job{
						Name: fmt.Sprintf("%s_job_%d", customer.Name, j+1),
						Data: nil,
					},
					m); err != nil {
					// fmt.Println(err)
					return
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
