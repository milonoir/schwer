package cpu

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/milonoir/schwer/resource"
	"github.com/shirou/gopsutil/cpu"
)

// Monitor represents a CPU load monitor.
type Monitor struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	l      *log.Logger

	usage resource.CPULevels
	mtx   sync.RWMutex
}

// NewMonitor returns a configured CPU load monitor.
func NewMonitor(cores int, l *log.Logger) *Monitor {
	return &Monitor{
		l:     l,
		usage: make(resource.CPULevels, cores),
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

// Usage returns the latest set of CPU utilisation levels.
func (m *Monitor) Usage() interface{} {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	// Return a deep-copy of values so we're not racy.
	u := make(resource.CPULevels, len(m.usage))
	copy(u, m.usage)
	return u
}

// monitor is the CPU load monitoring goroutine.
func (m *Monitor) monitor() {
	defer m.wg.Done()

	for {
		select {
		case <-m.cancel:
			return
		default:
			vals, err := cpu.Percent(time.Second, true)
			if err != nil {
				m.l.Printf("error in getting CPU utilisation levels: %s\n", err)
				continue
			}
			m.saveUsage(vals)
		}
	}
}

// saveUsage rounds values to the closest integer and stores them in a slice.
func (m *Monitor) saveUsage(values []float64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for i, v := range values {
		m.usage[i] = int(math.Round(v))
	}
}
