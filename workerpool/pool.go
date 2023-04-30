package workerpool

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Job type is the channel type for creation of jobs
type Job struct {
	URL string
}

// Result is the type of final check result of website status
type Result struct {
	URL          string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
}

func (r Result) Info() string {
	if r.Error != nil {
		return fmt.Sprintf("[ERROR] - [%s] - %s", r.URL, r.Error.Error())
	}

	return fmt.Sprintf("[SUCCESS] - [%s] - Status: %d, Response Time: %s", r.URL, r.StatusCode, r.ResponseTime.String())
}

// Pool type is the type of our thread pool that will check webstatus concurrently
type Pool struct {
	worker      *worker
	workerCount int

	jobs    chan Job
	results chan Result

	wg      *sync.WaitGroup
	stopped bool
}

// New function is the function to initialize and create Pool of workers and channels
func New(workerCount int, timeout time.Duration, results chan Result) *Pool {
	return &Pool{
		worker:      newWorker(timeout),
		workerCount: workerCount,
		jobs:        make(chan Job),
		results:     results,
		wg:          new(sync.WaitGroup),
	}
}

// Init() function that starts the workerCount number of workers
func (p *Pool) Init() {
	for i := 0; i < p.workerCount; i++ {
		go p.initWorker(i)
	}
}

func (p *Pool) Push(j Job) {
	// if the process stopped we don't wan't to start new jobs => we just skip and don't enter to next checking phase
	if p.stopped {
		return
	}

	// push created job into jobs channel so worker can receive it and add wg one job that we need to wait until its complete
	p.jobs <- j
	p.wg.Add(1)
}

func (p *Pool) Stop() {
	// pool is stopeed => no more jobs can be pushed to pool
	p.stopped = true
	// close jobs channel so all range loops in initWorker can stop their iterations in channel
	close(p.jobs)
	// we still wait until all workers in initWorker() finishes their jobs and after that our main program shutsdown
	p.wg.Wait()
}

// initWorker function that starts the worker by listening for jobs channel and giving results according to received job from channel
func (p *Pool) initWorker(id int) {
	for job := range p.jobs {
		time.Sleep(time.Second)
		p.results <- p.worker.process(job)
		p.wg.Done()
	}
	log.Printf("[worker ID %d] finished processing\n", id)
}
