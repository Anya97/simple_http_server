package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Anya97/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	type Counter struct {
		TotalRequestsCount      int64 `json:"total_requests_count"`       // Число запросов /ping
		TotalRequestErrorsCount int64 `json:"total_request_errors_count"` // Число запросов с 400 ответом
		TotalServerErrorsCount  int64 `json:"total_server_errors_count"`  // Число запросов с 500 ответом
	}
	counter := Counter{}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := config.GetConfig()

	requestsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_counter",
		Help: "Total count of requests",
	}, []string{"status_code", "endpoint"})

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_requests_seconds",
		Help: "Time taken to create hashes",
	}, []string{"status_code", "endpoint"})

	mux := http.NewServeMux()
	logger.Info("Started application")
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		code := http.StatusOK
		logger.Info("request:" + r.Method)
		defer func() {
			if r := recover(); r != nil {
				code = http.StatusInternalServerError
				w.WriteHeader(http.StatusInternalServerError)
				logger.With(r).Error("Internal server error")
				atomic.AddInt64(&counter.TotalServerErrorsCount, 1)
				requestsCounter.WithLabelValues(strconv.Itoa(code), "ping").Inc()
			}
			duration := time.Since(start)
			histogram.WithLabelValues(strconv.Itoa(code), "ping").Observe(duration.Seconds())
		}()
		randomNum := rand.Intn(10000-100) + 100
		atomic.AddInt64(&counter.TotalRequestsCount, 1)
		requestsCounter.WithLabelValues(strconv.Itoa(code), "ping").Inc()
		time.Sleep(time.Duration(randomNum) * time.Millisecond)

		v := rand.Intn(100)

		if v < 5 {
			code = http.StatusBadRequest
			w.WriteHeader(code)
			atomic.AddInt64(&counter.TotalRequestErrorsCount, 1)
			requestsCounter.WithLabelValues(strconv.Itoa(code), "ping").Inc()
			logger.With(code).Warn("Request error in /ping")
		} else if v < 15 {
			panic("AAAAAAA")
		}
	})

	mux.HandleFunc("/requests_counter", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		err := json.NewEncoder(w).Encode(counter)
		if err != nil {
			logger.With(err).Error("Json encode error")
		}
		histogram.WithLabelValues(strconv.Itoa(http.StatusOK), "requests_counter").Observe(time.Since(start).Seconds())
		requestsCounter.WithLabelValues(strconv.Itoa(http.StatusOK), "requests_counter").Inc()
	})

	mux.Handle("/metrics", promhttp.Handler())

	err := prometheus.Register(histogram)
	if err != nil {
		logger.With(err).Error("Register error")
	}

	server := http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.With(err).Error("error running http server")
			}
		}
	}()

	<-ctx.Done()
	logger.Info("got interruption signal")
	if err := server.Shutdown(context.TODO()); err != nil {
		logger.With(err).Info("server shutdown returned an err")
	}

	log.Println("final")
}
