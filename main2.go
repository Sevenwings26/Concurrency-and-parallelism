package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Job struct {
	ID int
}

func main() {
	numWorkers := 3
	numJobs := 10

	// 1. Context for Cancellation
	// We create a context that can be cancelled if an OS signal is received.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Buffered Channel for Backpressure
	// By giving the channel a buffer of 5, the producer can send up to 5 jobs
	// without waiting for a worker to pick them up. This absorbs traffic spikes.
	startJobs := make(chan Job, 5)

	// 3. Graceful Shutdown with WaitGroup
	// WaitGroup keeps track of how many workers are still active.
	var wg sync.WaitGroup

	// Start workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go advancedWorker(ctx, w, startJobs, &wg)
	}

	// 4. Handling System Signals (Graceful Shutdown)
	// We listen for SIGINT/SIGTERM (e.g., Ctrl+C or Kubernetes shutdown)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n[SYSTEM] Received shutdown signal. Initiating graceful shutdown...")
		// Tell workers to stop accepting new work and exit
		cancel()
	}()

	// Producer generating jobs
	go func() {
		for jobID := 1; jobID <= numJobs; jobID++ {
			select {
			case <-ctx.Done():
				// If context is cancelled, stop producing jobs
				fmt.Println("[PRODUCER] Stopped producing jobs due to cancellation.")
				close(startJobs)
				return
			case startJobs <- Job{ID: jobID}:
				fmt.Printf("[PRODUCER] Sent job %d\n", jobID)
			}
		}
		// Close channel when all jobs are sent normally
		close(startJobs)
	}()

	// Wait for all workers to finish their current tasks gracefully
	wg.Wait()
	fmt.Println("[SYSTEM] All workers gracefully shut down. Exiting application.")
}

func advancedWorker(ctx context.Context, id int, jobs <-chan Job, wg *sync.WaitGroup) {
	// Defer ensures wg.Done() is called when the worker exits, no matter what.
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			// 5. Handling Cancellation and Outages
			// If context is cancelled, the worker abandons the queue and returns.
			fmt.Printf("Worker %d: Context cancelled, stopping.\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				// Channel is closed, no more jobs
				fmt.Printf("Worker %d: No more jobs, shutting down.\n", id)
				return
			}
			
			// Simulate processing the job
			fmt.Printf("Worker %d started job %d\n", id, job.ID)
			
			// Simulate a network call or DB query that might take time.
			// Using select with a timeout to prevent hanging forever (Circuit Breaker / Timeout pattern).
			select {
			case <-time.After(1 * time.Second):
				fmt.Printf("Worker %d finished job %d\n", id, job.ID)
			case <-ctx.Done():
				fmt.Printf("Worker %d: Context cancelled while processing job %d! Dropping job.\n", id, job.ID)
				return
			}
		}
	}
}

/*
Hardware Impact Notes for Advanced Go Workers:
- Context Switches: Go's scheduler multiplexes thousands of goroutines onto a few OS threads, keeping CPU cache hot and avoiding expensive OS context switches.
- Memory: Buffer channels (like `make(chan Job, 5)`) use bounded memory. The system won't crash from OOM if traffic spikes.
- CPU Utilization: `select` statements don't spin the CPU. When a goroutine is waiting on a channel or timer, the scheduler puts it to sleep, freeing the CPU core for other tasks.
*/
