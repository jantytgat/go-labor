package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/jantytgat/go-labor/pkg/labor"
)

var (
	logLevel         = slog.LevelInfo
	evenLogLevel     = slog.LevelDebug
	managerName      = "example"
	maxJobs      int = 1000
	maxCustomers     = 200
	maxOperators int = runtime.NumCPU() * maxCustomers * 2
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	mc := labor.ManagerConfig{
		Address:       labor.NewAddress(labor.LocalAddress, "manager", managerName),
		EventLogger:   logger,
		EventLogLevel: evenLogLevel,
		MaxOperators:  maxOperators,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := labor.NewManager(mc)
	m.Enable(ctx)

	var customers = make([]*labor.Customer, maxCustomers)
	var mux sync.Mutex

	wg := &sync.WaitGroup{}
	go func(wg *sync.WaitGroup) {
		for i := 0; i < maxCustomers; i++ {
			customer := labor.NewCustomer(fmt.Sprintf("customer_%d", i+1))
			mux.Lock()
			customers[i] = customer
			mux.Unlock()

			go func(ctx context.Context, c *labor.Customer, id int) {
				for j := 0; j < maxJobs; j++ {
					err := c.Send(
						ctx,
						labor.Job{
							Name: fmt.Sprintf("%s_job_%d", customer.Name, j+1),
							Data: nil,
							Pipeline: labor.Pipeline{
								Sequence: []labor.Process{
									{
										Task:   labor.PrintTask{},
										Data:   fmt.Sprintf("%s_job_%d_1", customer.Name, j+1),
										Output: nil,
									}, {
										Task:   labor.PrintTask{},
										Data:   fmt.Sprintf("%s_job_%d_2", customer.Name, j+1),
										Output: nil,
									},
								},
								Data: nil,
							},
						},
						m)
					if err != nil {
						return
					}
				}
			}(ctx, customer, i)

			go func(ctx context.Context, c *labor.Customer) {
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
			}(context.Background(), customer)
		}
		wg.Done()
	}(wg)
	wg.Add(1)

	wg.Wait()
	startTime := time.Now()

	var requests int
	var responses int

Detect:
	for {
		requests = 0
		responses = 0

		mux.Lock()
		for i := 0; i < maxCustomers; i++ {
			requests = requests + customers[i].RequestsTotal()
			responses = responses + customers[i].ResponsesTotal()
		}
		mux.Unlock()

		if requests != 0 && responses != 0 && requests == responses {
			break Detect
		}
	}

	fmt.Println("Processed requests:", requests)
	fmt.Println("Processed responses:", responses)
	fmt.Println("Total time:", time.Since(startTime))
}
