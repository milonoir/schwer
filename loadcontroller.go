package main

import (
	"log"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const (
	// defaultCPULoadPct is used on start-up.
	defaultCPULoadPct = 10

	megaBytes = 1 << 20
)

type memoryStats struct {
	Total     int `json:"total"`
	Available int `json:"available"`
	Used      int `json:"used"`
	UsedPct   int `json:"usedpct"`
}

// loadController is responsible for generating the requested amount of CPU load over time.
type loadController struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	logger *log.Logger

	cpuCores  int
	cpuChange []chan time.Duration
	cpuUtil   []int
	cpuMtx    sync.RWMutex

	memStat memoryStats
	memMtx  sync.RWMutex
}

// newLoadController returns a configured loadController.
func newLoadController(cpuCores int, l *log.Logger) *loadController {
	return &loadController{
		cancel:    make(chan struct{}),
		cpuCores:  cpuCores,
		cpuChange: make([]chan time.Duration, cpuCores),
		cpuUtil:   make([]int, 4),
		logger:    l,
	}
}

// start spins up load goroutines for each CPU core.
func (lc *loadController) start() {
	lc.logger.Println("starting load")

	lc.wg.Add(2)
	go lc.cpuMonitor()
	go lc.memMonitor()

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

// cpuMonitor continuously polls CPU utilisation levels and stores the values.
func (lc *loadController) cpuMonitor() {
	defer lc.wg.Done()

	for {
		select {
		case <-lc.cancel:
			return
		default:
			vals, err := cpu.Percent(time.Second, true)
			if err != nil {
				lc.logger.Printf("error in getting CPU utilisation levels: %s\n", err)
				continue
			}
			lc.saveCPUUtilisation(vals)
		}
	}
}

// saveCPUUtilisation rounds values to the closest integer and stores them in a slice.
func (lc *loadController) saveCPUUtilisation(values []float64) {
	lc.cpuMtx.Lock()
	defer lc.cpuMtx.Unlock()

	for i, v := range values {
		lc.cpuUtil[i] = int(math.Round(v))
	}
}

// cpuUsage returns a copy of the stored CPU utilisation levels.
func (lc *loadController) cpuUsage() []int {
	lc.cpuMtx.RLock()
	defer lc.cpuMtx.RUnlock()

	// Return a deep-copy of values so we're not racy.
	u := make([]int, len(lc.cpuUtil))
	copy(u, lc.cpuUtil)
	return u
}

// cpuLoad is a CPU load goroutine.
func (lc *loadController) cpuLoad(n int, changed <-chan time.Duration) {
	defer lc.wg.Done()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sleep := lc.sleepDuration(defaultCPULoadPct)
	lc.logger.Printf("thread %d sleep duration: %s\n", n, sleep)
	for {
		select {
		case <-lc.cancel:
			return
		case sleep = <-changed:
			lc.logger.Printf("thread %d new sleep duration: %s\n", n, sleep)
		default:
			time.Sleep(sleep)
		}
	}
}

// sleepDuration returns how long the CPU goroutine should sleep in a second.
func (lc *loadController) sleepDuration(pct int64) time.Duration {
	return time.Duration(100-pct) * 10 * time.Microsecond
}

// updateCPULoad updates the sleep duration of all CPU load goroutines.
func (lc *loadController) updateCPULoad(pct int64) {
	lc.logger.Printf("updating cpu load percentage: %d%%\n", pct)
	for _, ch := range lc.cpuChange {
		ch <- lc.sleepDuration(pct)
	}
}

func (lc *loadController) memMonitor() {
	defer lc.wg.Done()

	for {
		select {
		case <-lc.cancel:
			return
		default:
			memStat, err := mem.VirtualMemory()
			if err != nil {
				lc.logger.Printf("error in getting virtual memory stats: %s\n", err)
				continue
			}
			lc.saveMemUsage(memStat.Total, memStat.Available, memStat.Used, memStat.UsedPercent)
			time.Sleep(time.Second)
		}
	}
}

func (lc *loadController) saveMemUsage(total, avail, used uint64, usedPct float64) {
	lc.memMtx.Lock()
	defer lc.memMtx.Unlock()

	lc.memStat.Total = int(total / megaBytes)
	lc.memStat.Available = int(avail / megaBytes)
	lc.memStat.Used = int(used / megaBytes)
	lc.memStat.UsedPct = int(math.Round(usedPct))
}

func (lc *loadController) memUsage() memoryStats {
	lc.cpuMtx.RLock()
	defer lc.cpuMtx.RUnlock()

	return lc.memStat
}
