package main

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type load struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	logger *log.Logger

	cpuCores  int
	cpuPct    int32
	cpuChange []chan struct{}
}

func newLoad(cpuCores int, l *log.Logger) *load {
	return &load{
		cancel:    make(chan struct{}),
		logger:    l,
		cpuCores:  cpuCores,
		cpuPct:    10,
		cpuChange: make([]chan struct{}, cpuCores),
	}
}

func (l *load) start() {
	l.logger.Println("starting load")
	cpuPct := atomic.LoadInt32(&l.cpuPct)

	l.wg.Add(l.cpuCores)
	for i := 0; i < l.cpuCores; i++ {
		l.logger.Printf("set CPU core %d load at %d%%\n", i, cpuPct)
		ch := make(chan struct{}, 1)
		l.cpuChange[i] = ch
		go l.cpu(ch)
	}
}

func (l *load) stop() {
	close(l.cancel)
	l.wg.Wait()
}

func (l *load) cpu(changed <-chan struct{}) {
	defer l.wg.Done()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sleep := l.getSleepDuration()
	for {
		select {
		case <-l.cancel:
			return
		case <-changed:
			sleep = l.getSleepDuration()
		default:
			time.Sleep(sleep)
		}
	}
}

func (l *load) getSleepDuration() time.Duration {
	sleep := time.Duration(100-int(atomic.LoadInt32(&l.cpuPct))) * 10 * time.Microsecond
	return sleep
}

func (l *load) updateCPUPct(pct int32) {
	atomic.StoreInt32(&l.cpuPct, pct)
	l.logger.Printf("updating cpu load percentage: %d%%\n", pct)
	for _, ch := range l.cpuChange {
		ch <- struct{}{}
	}
}
