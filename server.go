package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/milonoir/schwer/statik"
	"github.com/rakyll/statik/fs"
)

func newServer(port uint64, ld *load, l *log.Logger) *http.Server {
	router := http.NewServeMux()
	router.Handle("/", indexHandler())
	router.Handle("/hello", helloHandler())
	router.Handle("/cpu", updateCPUPctHandler(ld))

	return &http.Server{
		Addr:         ":" + strconv.FormatUint(port, 10),
		Handler:      router,
		ErrorLog:     l,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}

func indexHandler() http.Handler {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	return http.FileServer(statikFS)
}

func helloHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	})
}

func updateCPUPctHandler(l *load) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("unable to parse request: %s", err), http.StatusBadRequest)
			return
		}

		v := r.FormValue("pct")
		pct, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			http.Error(w, "invalid percentage value", http.StatusBadRequest)
			return
		}

		if pct < 0 || pct > 100 {
			http.Error(w, "percentage must be between 0-100", http.StatusBadRequest)
			return
		}

		l.updateCPUPct(int32(pct))
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("cpu percentage updated"))
	})
}
