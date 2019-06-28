package memory

import (
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"
)

// Load represents a memory load.
type Load struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	l      *log.Logger

	alloc    [][]byte
	change   chan int
	pageSize int
}

// NewLoad returns a configured memory load.
func NewLoad(l *log.Logger) *Load {
	return &Load{
		change:   make(chan int, 1),
		pageSize: os.Getpagesize(),
		l:        l,
	}
}

// Start starts up the load goroutine.
func (l *Load) Start() {
	l.cancel = make(chan struct{})

	l.wg.Add(1)
	go l.load()
}

// Stop signals the load goroutine to stop and waits for it to return.
func (l *Load) Stop() {
	close(l.cancel)
	l.wg.Wait()
}

// Update updates the allocated memory size.
func (l *Load) Update(size int64) {
	l.l.Printf("updating mem load to %d MB\n", size)
	l.change <- int(size)
}

func (l *Load) load() {
	defer l.wg.Done()
	defer func() {
		l.alloc = nil
		runtime.GC()
	}()

	for {
		// Do not use default branch in select as we don't want a busy loop.
		select {
		case <-l.cancel:
			return
		case size := <-l.change:
			l.alloc = nil
			runtime.GC()
			for page := 0; page < size*megaBytes/l.pageSize; page++ {
				// Allocate memory in page-sized chunks.
				chunk := make([]byte, l.pageSize)
				l.alloc = append(l.alloc, chunk)
			}
			l.l.Printf("mem alloc - page size: %d bytes, pages: %d, size: %d MB\n", l.pageSize, len(l.alloc), len(l.alloc)*l.pageSize/megaBytes)
		case <-time.After(time.Second):
			// Make sure we use the allocated memory, so it won't get swapped.
			if l.alloc != nil {
				for page := 0; page < len(l.alloc); page++ {
					l.alloc[page][rand.Intn(l.pageSize)]++
				}
			}
		}
	}
}
