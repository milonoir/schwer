package cpu

import (
	"log"
	"runtime"
	"sync"
	"time"
)

// LoadController represents a CPU load controller.
type LoadController struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	l      *log.Logger

	cores  int
	change []chan time.Duration
}

// NewLoadController returns a configured CPU load controller.
func NewLoadController(cores int, l *log.Logger) *LoadController {
	return &LoadController{
		cores:  cores,
		change: make([]chan time.Duration, cores),
		l:      l,
	}
}

// Start starts up the load goroutines.
func (c *LoadController) Start() {
	c.cancel = make(chan struct{})

	c.wg.Add(c.cores)
	for i := 0; i < c.cores; i++ {
		ch := make(chan time.Duration, 1)
		c.change[i] = ch
		go c.load(i, ch)
	}

}

// Stop signals all goroutines to stop and waits for them to return.
func (c *LoadController) Stop() {
	close(c.cancel)
	c.wg.Wait()
}

// Update updates the load percentage of all goroutines.
func (c *LoadController) Update(pct int64) {
	c.l.Printf("updating cpu load percentage: %d%%\n", pct)
	for _, ch := range c.change {
		ch <- c.sleepDuration(pct)
	}
}

// cpuLoad is a CPU load goroutine.
func (c *LoadController) load(n int, changed <-chan time.Duration) {
	defer c.wg.Done()

	// Bind the goroutine to an OS thread, so the scheduler won't move it around.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sleep := c.sleepDuration(0)

	// TODO: improve busy loop!
	for {
		select {
		case <-c.cancel:
			return
		case sleep = <-changed:
			c.l.Printf("thread %d sleep duration: %s\n", n, sleep)
		case <-time.After(time.Millisecond - sleep):
			time.Sleep(sleep)
		default:
			// Default branch is required to keep the infinite loop busy.
		}
	}
}

// sleepDuration returns how long the CPU goroutine should sleep in a second.
func (c *LoadController) sleepDuration(pct int64) time.Duration {
	return time.Duration(100-pct) * 10 * time.Millisecond
}
