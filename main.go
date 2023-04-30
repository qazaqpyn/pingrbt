package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qazaqpyn/pingrbt/workerpool"
)

const (
	INTERVAL       = time.Second * 5
	REQUET_TIMEOUT = time.Second * 2
	WORKERS_COUNT  = 3
)

var urls = []string{
	"https://workshop.zhashkevych.com/",
	"https://golang-ninja.com/",
	"https://zhashkevych.com/",
	"https://google.com/",
	"https://golang.org/",
}

func main() {
	results := make(chan workerpool.Result)
	workerPool := workerpool.New(WORKERS_COUNT, REQUET_TIMEOUT, results)

	// run our workers that range jobs channel that was created by workerPool
	workerPool.Init()

	// generateJobs running in goroutine concurrently because it creates forever loop => if not concurrenly then our program won't go further because this function never ends
	go generateJobs(workerPool)

	// processResults running in the goroutine concurrenly because if not our function never goes  below it, and we can't make graceful shutdown
	go processResults(results)

	// implementation of graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	workerPool.Stop()
}

// processResults that starts the goroutine for ranging the results channel until its closed for results from workers
func processResults(results chan workerpool.Result) {
	go func() {
		for result := range results {
			fmt.Println(result.Info())
		}
	}()
}

// function that runs forever, only stops when our main func terminates, this function puts to the channel all urls that need to be checked
// one by one by Push() which pushes the job into channel and sleep for some interval and repeats all steps again
func generateJobs(wp *workerpool.Pool) {
	for {
		for _, url := range urls {
			wp.Push(workerpool.Job{URL: url})
		}

		time.Sleep(INTERVAL)
	}
}
