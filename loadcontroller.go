package main

import (
	"log"
	"runtime"
	"sync"
	"time"
)

const (
	// defaultCPULoadPct is used on start-up.
	defaultCPULoadPct = 10
)

// loadController is responsible for generating the requested amount of CPU load over time.
type loadController struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	logger *log.Logger

	cpuCores  int
	cpuChange []chan time.Duration
}

// newLoadController returns a configured loadController.
func newLoadController(cpuCores int, l *log.Logger) *loadController {
	return &loadController{
		cancel:    make(chan struct{}),
		cpuCores:  cpuCores,
		cpuChange: make([]chan time.Duration, cpuCores),
		logger:    l,
	}
}

// start spins up load goroutines for each CPU core.
func (lc *loadController) start() {
	lc.logger.Println("starting load")

	lc.wg.Add(lc.cpuCores)
	for i := 0; i < lc.cpuCores; i++ {
		ch := make(chan time.Duration, 1)
		lc.cpuChange[i] = ch
		go lc.cpuLoad(i, ch)
	}
}

// stop stops load goroutines.
func (lc *loadController) stop() {
	close(lc.cancel)
	lc.wg.Wait()
}

// cpuLoad is a CPU load goroutine.
func (lc *loadController) cpuLoad(n int, changed <-chan time.Duration) {
	defer lc.wg.Done()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sleep := lc.sleepDuration(defaultCPULoadPct)
	lc.logger.Printf("CPU core %d sleep duration: %s\n", n, sleep)
	for {
		select {
		case <-lc.cancel:
			return
		case sleep = <-changed:
			lc.logger.Printf("CPU core %d new sleep duration: %s\n", n, sleep)
		default:
			time.Sleep(sleep)
		}
	}
}

// sleepDuration returns how long the CPU goroutine should sleep in a second.
func (lc *loadController) sleepDuration(pct int64) time.Duration {
	return time.Duration(100-pct) * 10 * time.Microsecond
}

func (lc *loadController) updateCPULoad(pct int64) {
	lc.logger.Printf("updating cpu load percentage: %d%%\n", pct)
	for _, ch := range lc.cpuChange {
		ch <- lc.sleepDuration(pct)
	}
}
