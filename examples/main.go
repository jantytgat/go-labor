package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/jantytgat/go-labor/pkg/tasks"

	"github.com/jantytgat/go-labor/pkg/labor"
)

var (
	logLevel          = slog.LevelInfo
	eventLogLevel     = slog.LevelDebug
	managerName       = "example"
	maxJobs       int = 2000
	maxCustomers      = 10
	// maxOperators  int = runtime.NumCPU() * maxCustomers * 2
	maxOperators = 2
)

func main() {
	// var cpuProfile, memProfile *os.File
	// var err error
	// cpuProfile, err = os.Create("examples/cpu.profile.log")
	// if err != nil {
	//	panic(err)
	// }
	// defer cpuProfile.Close()
	// if err = pprof.StartCPUProfile(cpuProfile); err != nil {
	//	panic(err)
	// }
	// defer pprof.StopCPUProfile()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	mc := labor.ManagerConfig{
		Address:       labor.NewAddress(labor.LocalAddress, "manager", managerName),
		EventLogger:   logger,
		EventLogLevel: eventLogLevel,
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
							Sequence: []labor.Process{
								{
									Handler: tasks.PrintHandler,
								}, {
									Handler: tasks.PrintHandler,
								},
							},
							Data: labor.Pipeline{},
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
							// logger.Log(ctx, slog.LevelInfo, "received reply", slog.Any("reply", res))
						}
					}
				}
			}(context.Background(), customer)
		}
		wg.Done()
	}(wg)
	wg.Add(1)

	wg.Wait()

	var requests int
	var responses int
	var duration time.Duration
	startTime := time.Now()

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
			duration = time.Since(startTime)
			m.Disable()
			break Detect
		}
	}

	fmt.Println("Processed requests:", requests)
	fmt.Println("Processed responses:", responses)
	fmt.Println("Total time:", duration)
	//
	// memProfile, err = os.Create("examples/mem.profile.log")
	// if err != nil {
	//	panic(err)
	// }
	// defer memProfile.Close()
	// runtime.GC()
	// if err = pprof.Lookup("allocs").WriteTo(memProfile, 2); err != nil {
	//	panic(err)
	// }
	// if err = pprof.Lookup("heap").WriteTo(memProfile, 2); err != nil {
	//	panic(err)
	// }
}
