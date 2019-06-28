package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/milonoir/schwer/resource/cpu"
	"github.com/milonoir/schwer/resource/memory"
)

const (
	minPort     = 1024
	maxPort     = 65535
	defaultPort = 9999

	serverShutdownTimeout = 30 * time.Second
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// _main sets up all moving parts. It is not done in main() in order to be able to
// execute deferred functions and return an error.
func _main() error {
	// Parse command line args.
	port := flag.Uint64("port", defaultPort, fmt.Sprintf("the port number (%d-%d) the server binds to", minPort, maxPort))
	flag.Parse()

	// Validate port.
	if *port < minPort || *port > maxPort {
		flag.Usage()
		return errors.New("invalid port number")
	}

	// Setup logger.
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Setup load and monitoring.
	cores := runtime.NumCPU()
	c := NewController(
		cpu.NewLoad(cores, logger),
		memory.NewLoad(logger),
		cpu.NewMonitor(cores, logger),
		memory.NewMonitor(logger),
	)
	c.Start()
	defer c.Stop()

	// Setup HTTP server.
	server := newServer(*port, c, logger)

	// Setup signal handler.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		logger.Println("shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("could not shutdown server gracefully: %s\n", err)
		}
	}()

	// Run server.
	logger.Printf("starting server on :%d\n", *port)
	return server.ListenAndServe()
}
