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
)

const (
	minPort = 1024
	maxPort = 65535
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func _main() error {
	// Parse command line args.
	port := flag.Uint64("port", 9999, "the port number (1024-65535) the server binds to")
	flag.Parse()

	// Validate port.
	if *port < minPort || *port > maxPort {
		flag.Usage()
		return errors.New("invalid port number")
	}

	// Setup logger.
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Setup load handlers.
	load := newLoad(runtime.NumCPU(), logger)
	load.start()
	defer load.stop()

	// Setup HTTP server.
	server := newServer(*port, load, logger)

	// Setup signal handler.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		logger.Println("shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
