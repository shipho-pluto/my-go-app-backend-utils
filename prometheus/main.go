package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "app_request_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"nethod", "path"})

	activeUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_active_users",
		Help: "Number of active users",
	})

	temperatureMetric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_temperature_celsius",
		Help: "Current tamperature in Celsius",
	})
)

func main() {
	ctx := context.Background()
	// Запускаем горутину для обновления метрик
	go updateMetrics(ctx)

	// HTTP обработчики
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/data", dataHandler)
	http.HandleFunc("/api/error", errorHandler)

	// Метрики Prometheus
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Имитируем обработку
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	w.Write([]byte("Hello from Go Metrics App!"))

	// Записываем метрики
	duration := time.Since(start).Seconds()
	requestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
	requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	w.Write([]byte(`{"data": "some json response"}`))

	duration := time.Since(start).Seconds()
	requestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
	requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))

	duration := time.Since(start).Seconds()
	requestsTotal.WithLabelValues(r.Method, r.URL.Path, "500").Inc()
	requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
}

func updateMetrics(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
		case <-ticker.C:
			// Обновляем метрики
			activeUsers.Set(float64(rand.Intn(1000)))
			temperatureMetric.Set(20 + rand.Float64()*10 - 5)
		}
	}
}
