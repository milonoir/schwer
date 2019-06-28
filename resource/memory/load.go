package memory

import (
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	megaBytes = 1 << 20
)

// LoadController represents a CPU load controller.
type LoadController struct {
	cancel chan struct{}
	wg     sync.WaitGroup
	l      *log.Logger

	alloc    [][]byte
	change   chan int
	pageSize int
}

// NewLoadController returns a configured memory load controller.
func NewLoadController(l *log.Logger) *LoadController {
	return &LoadController{
		change:   make(chan int, 1),
		pageSize: os.Getpagesize(),
		l:        l,
	}
}

// Start starts up the load goroutine.
func (c *LoadController) Start() {
	c.cancel = make(chan struct{})

	c.wg.Add(1)
	go c.load()
}

// Stop signals the load goroutine to stop and waits for it to return.
func (c *LoadController) Stop() {
	close(c.cancel)
	c.wg.Wait()
}

// Update updates the allocated memory size.
func (c *LoadController) Update(size int64) {
	c.l.Printf("updating mem load to %d MB\n", size)
	c.change <- int(size)
}

func (c *LoadController) load() {
	defer c.wg.Done()
	defer func() {
		c.alloc = nil
		runtime.GC()
	}()

	for {
		// Do not use default branch in select as we don't want a busy loop.
		select {
		case <-c.cancel:
			return
		case size := <-c.change:
			c.alloc = nil
			runtime.GC()
			for page := 0; page < size*megaBytes/c.pageSize; page++ {
				// Allocate memory in page-sized chunks.
				chunk := make([]byte, c.pageSize)
				c.alloc = append(c.alloc, chunk)
			}
			c.l.Printf("mem alloc - page size: %d bytes, pages: %d, size: %d MB\n", c.pageSize, len(c.alloc), len(c.alloc)*c.pageSize/megaBytes)
		case <-time.After(time.Second):
			// Make sure we use the allocated memory, so it won't get swapped.
			if c.alloc != nil {
				for page := 0; page < len(c.alloc); page++ {
					c.alloc[page][rand.Intn(c.pageSize)]++
				}
			}
		}
	}
}
