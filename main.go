/*
How Goroutine/workers and Job
How the worker handles multiple jobs

*/
package main

import (
	"fmt"
	"time"
)

// struct is used to pass information between functions, whereby parameters are defined
type Job struct {
	ID int
}

// step 1: define variable 
func main(){
	numJobs := 5
	numWorkers := 3

	// step 2
	// channels: 2 channels
	startJobs := make(chan Job)
	jobDone := make(chan bool)
	

	// step 3 
	// for loop 
	for w := 0; w < numWorkers; w++{
		// call goroutine 
		go worker(w, startJobs, jobDone)
	}

	// step 5 
	go func(){
		for job := 0; job < numJobs; job++{
			startJobs <- Job{ID: job}
		}
		close(startJobs)
	}()

	// step 6: For all jobs to complete
	for job := 0; job < numJobs; job++{
		<-jobDone
	}
	fmt.Println("All jobs completed")
}

// step 4 
// create the go function 
func worker(id int, startJobs <-chan Job, jobDone chan<- bool){
	for job := range startJobs{
		fmt.Printf("Worker %d started processing job %d\n", id, job.ID)
		time.Sleep(time.Second)
		fmt.Printf("Worker %d finished processing job %d\n", id, job.ID)
		jobDone <- true
	}
}


/*

I think this is working like multithreading in python using semaphores...  

[Expanded Explanation]
Yes, this is structurally similar to Python's ThreadPoolExecutor or manually using threading.Semaphore to cap concurrency. However, conceptually and under the hood, it's very different.

1. Python Multi-threading vs. Goroutines:
   - Python threads are OS-level threads (pthreads), and due to the Global Interpreter Lock (GIL), they don't achieve true parallelism for CPU tasks. They also have a large memory footprint (~1MB+).
   - Go uses an M:N scheduler (M goroutines multiplexed over N OS threads). Goroutines are lightweight (starting at ~2KB) and bypass OS overhead, allowing you to spawn hundreds of thousands easily.
   - Go allows true parallelism on multiple CPU cores.

2. Shared State vs. Channels (CSP):
   - In Python, you typically share memory (a list or queue) and use locks/semaphores to control access to it.
   - In Go, you communicate by passing messages over Channels (`startJobs`, `jobDone`). "Do not communicate by sharing memory; instead, share memory by communicating."

3. Hardware Impact:
   - Python (Threads): High memory consumption. Context switching between OS threads incurs high CPU overhead. True CPU parallelism is blocked by GIL.
   - Python (AsyncIO): Low memory (single thread), but CPU-bound tasks block the entire event loop, freezing the server.
   - Go (Goroutines): Extremely low memory footprint. The Go runtime optimally schedules execution on available hardware CPU cores without blocking other tasks during I/O. 

For advanced systems (under load, failure, cancellation), you need:
- Backpressure (buffered channels, load shedding).
- Cancellation (context timeouts).
- Graceful shutdown (sync.WaitGroup).

(See main2.go and main3.py for advanced implementations of these patterns)

Output for :=
5 jobs
3 workers
Worker 0 started processing job 0
Worker 2 started processing job 2
Worker 1 started processing job 1
Worker 0 finished processing job 0
Worker 0 started processing job 3
Worker 2 finished processing job 2
Worker 1 finished processing job 1
Worker 1 started processing job 4
Worker 1 finished processing job 4
Worker 0 finished processing job 3
All jobs completed
*/
