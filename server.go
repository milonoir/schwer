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

// newServer returns a new configured http.Server with all endpoints registered to it.
func newServer(port uint64, ld *loadController, l *log.Logger) *http.Server {
	router := http.NewServeMux()
	router.Handle("/", indexHandler())
	router.Handle("/cpu", cpuHandler(ld))
	router.Handle("/mem", memHandler(ld))

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
func cpuHandler(lc *loadController) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			b, err := json.Marshal(lc.cpuUsage())
			if err != nil {
				http.Error(w, fmt.Sprintf("Server error: %s", err), http.StatusInternalServerError)
			}
			w.Write(b)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, fmt.Sprintf("Unable to parse request: %s", err), http.StatusBadRequest)
				return
			}

			v := r.FormValue("pct")
			pct, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				http.Error(w, "Invalid percentage value", http.StatusBadRequest)
				return
			}

			if pct < 0 || pct > 100 {
				http.Error(w, "Percentage value must be between 0-100", http.StatusBadRequest)
				return
			}

			lc.updateCPULoad(pct)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("CPU load percentage updated"))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func memHandler(lc *loadController) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			b, err := json.Marshal(lc.memUsage())
			if err != nil {
				http.Error(w, fmt.Sprintf("Server error: %s", err), http.StatusInternalServerError)
			}
			w.Write(b)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, fmt.Sprintf("Unable to parse request: %s", err), http.StatusBadRequest)
				return
			}

			v := r.FormValue("size")
			size, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				http.Error(w, "Invalid size value", http.StatusBadRequest)
				return
			}

			if size < 0 {
				http.Error(w, "Size value must be positive", http.StatusBadRequest)
				return
			}

			lc.updateMemLoad(int(size))
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Memory allocation size updated"))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
