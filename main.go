package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const serverPort = 3333

func main() {
	type Counter struct {
		TotalRequestsCount      int // Число запросов /ping
		TotalRequestErrorsCount int // Число запросов с 400 ответом
		TotalServerErrorsCount  int // Число запросов с 500 ответом
	}
	counter := Counter{}

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "ping_seconds",
		Help: "Time taken to create hashes",
	}, []string{"code"})
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		code := 200
		logger.Info("request:" + r.Method)
		defer func() {
			if r := recover(); r != nil {
				code = 500
				w.WriteHeader(500)
				logger.Error("Internal server error", r)
				counter.TotalServerErrorsCount++
			}
			duration := time.Since(start)
			histogram.WithLabelValues(fmt.Sprintf("%d", code)).Observe(duration.Seconds())
		}()
		randomNum := rand.Intn(10000-100) + 100
		counter.TotalRequestsCount++
		time.Sleep(time.Duration(randomNum) * time.Microsecond)
		if rand.Intn(10) == 8 {
			code = 400
			w.WriteHeader(400)
			counter.TotalRequestErrorsCount++
			logger.Warn("Request error")
		} else if rand.Intn(20) == 8 {
			panic("AAAAAAA")
		}
	})
	mux.HandleFunc("/requests_counter", func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(counter)
		if err != nil {
			logger.Error("Json encode error")
		}
	})
	mux.Handle("/metrics", prometheusHandler())

	prometheus.Register(histogram)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", serverPort),
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error running http server: %s\n", err)
		}
	}
}

func prometheusHandler() http.Handler {
	return promhttp.Handler()
}
