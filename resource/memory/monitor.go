package memory

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/milonoir/schwer/resource"
	"github.com/shirou/gopsutil/mem"
)

// Monitor represents a memory load monitor.
type Monitor struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	l      *log.Logger

	usage resource.MemStats
	mtx   sync.RWMutex
}

// NewMonitor returns a configured memory load monitor.
func NewMonitor(l *log.Logger) *Monitor {
	return &Monitor{
		l: l,
	}
}

// Start starts up the monitoring goroutine.
func (m *Monitor) Start() {
	m.cancel = make(chan struct{})

	m.wg.Add(1)
	go m.monitor()
}

// Stop signals the monitoring goroutine to stop and waits for it to return.
func (m *Monitor) Stop() {
	close(m.cancel)
	m.wg.Wait()
}

// Usage returns the latest set of memory stats.
func (m *Monitor) Usage() interface{} {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	return m.usage
}

// monitor is the memory load monitoring goroutine.
func (m *Monitor) monitor() {
	defer m.wg.Done()

	for {
		select {
		case <-m.cancel:
			return
		default:
			usage, err := mem.VirtualMemory()
			if err != nil {
				m.l.Printf("error in getting virtual memory stats: %s\n", err)
				continue
			}
			m.saveUsage(usage.Total, usage.Available, usage.Used, usage.UsedPercent)
			time.Sleep(time.Second)
		}
	}
}

// saveUsage stores memory stats in MB and rounds used percentage to the closest integer value.
func (m *Monitor) saveUsage(total, avail, used uint64, usedPct float64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.usage.Total = int(total / megaBytes)
	m.usage.Available = int(avail / megaBytes)
	m.usage.Used = int(used / megaBytes)
	m.usage.UsedPct = int(math.Round(usedPct))
}
