package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/milonoir/schwer/statik"
	"github.com/rakyll/statik/fs"
)

const (
	tplServerError = "Server error: %s"
	tplParseError  = "Unable to parse request: %s"
)

// newServer returns a new configured http.Server with all endpoints registered to it.
func newServer(port uint64, c *Controller, l *log.Logger) *http.Server {
	router := http.NewServeMux()
	router.Handle("/", indexHandler())
	router.Handle("/cpu", cpuHandler(c))
	router.Handle("/mem", memHandler(c))

	return &http.Server{
		Addr:         ":" + strconv.FormatUint(port, 10),
		Handler:      router,
		ErrorLog:     l,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}

// indexHandler is the main web front-end handler.
func indexHandler() http.Handler {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	return http.FileServer(statikFS)
}

// cpuHandler handles requests for:
// - (GET)  getting current CPU utilisation levels;
// - (POST) updating CPU load percentage.
func cpuHandler(c *Controller) http.Handler {
	return makeHandler(
		c.CPUUtilisationLevels,
		c.UpdateCPULoad,
		"pct",
		func(pct int64) error {
			if pct < 0 || pct > 100 {
				return fmt.Errorf("Percentage value must be between 0-100, got: %d", pct)
			}
			return nil
		},
		"CPU load percentage updated",
	)
}

// memHandler handles requests for:
// - (GET)  getting current memory stats;
// - (POST) updating the size of the allocation in memory load.
func memHandler(c *Controller) http.Handler {
	return makeHandler(
		c.MemStats,
		c.UpdateMemLoad,
		"size",
		func(size int64) error {
			if size < 0 {
				return fmt.Errorf("Size value must be positive: %d", size)
			}
			return nil
		},
		"Memory allocation size updated",
	)
}

func makeHandler(getFunc func() interface{}, setFunc func(int64), formValue string, validator func(int64) error, successMsg string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			b, err := json.Marshal(getFunc())
			if err != nil {
				http.Error(w, fmt.Sprintf(tplServerError, err), http.StatusInternalServerError)
			}
			w.Write(b)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, fmt.Sprintf(tplParseError, err), http.StatusBadRequest)
				return
			}

			v := r.FormValue(formValue)
			intValue, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid %s value", formValue), http.StatusBadRequest)
				return
			}

			if err := validator(intValue); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			setFunc(intValue)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(successMsg))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
