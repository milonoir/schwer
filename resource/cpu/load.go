package cpu

import (
	"log"
	"runtime"
	"sync"
	"time"
)

// Load represents a CPU load.
type Load struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	l      *log.Logger

	cores  int
	change []chan time.Duration
}

// NewLoad returns a configured CPU load.
func NewLoad(cores int, l *log.Logger) *Load {
	return &Load{
		cores:  cores,
		change: make([]chan time.Duration, cores),
		l:      l,
	}
}

// Start starts up the load goroutines.
func (l *Load) Start() {
	l.cancel = make(chan struct{})

	l.wg.Add(l.cores)
	for i := 0; i < l.cores; i++ {
		ch := make(chan time.Duration, 1)
		l.change[i] = ch
		go l.load(i, ch)
	}

}

// Stop signals all goroutines to stop and waits for them to return.
func (l *Load) Stop() {
	close(l.cancel)
	l.wg.Wait()
}

// Update updates the load percentage of all goroutines.
func (l *Load) Update(pct int64) {
	l.l.Printf("updating cpu load percentage: %d%%\n", pct)
	for _, ch := range l.change {
		ch <- l.sleepDuration(pct)
	}
}

// cpuLoad is a CPU load goroutine.
func (l *Load) load(n int, changed <-chan time.Duration) {
	defer l.wg.Done()

	// Bind the goroutine to an OS thread, so the scheduler won't move it around.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sleep := l.sleepDuration(0)

	// TODO: improve busy loop!
	for {
		select {
		case <-l.cancel:
			return
		case sleep = <-changed:
			l.l.Printf("thread %d sleep duration: %s\n", n, sleep)
		case <-time.After(time.Millisecond - sleep):
			time.Sleep(sleep)
		default:
			// Default branch is required to keep the infinite loop busy.
		}
	}
}

// sleepDuration returns how long the CPU goroutine should sleep in a second.
func (l *Load) sleepDuration(pct int64) time.Duration {
	return time.Duration(100-pct) * 10 * time.Millisecond
}
