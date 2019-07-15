package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prometheus_demo_request_count",
		Help: "Number of Requests to endpoint",
	}, []string{"code", "path"})

	requestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "prometheus_demo_request_latency_seconds",
		Help:    "Request Latency",
		Buckets: []float64{0.001, 0.002, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"code", "path"})
)

type PrometheusDemo struct {
	redisClient *redis.Client
}

func main() {
	log.Printf("Starting Server")
	redisAddr := os.Getenv("REDIS_ADDR")
	promDemo := newPrometheusDemo(redisAddr)
	http.HandleFunc("/sayhi", promDemo.HelloServer)
	http.HandleFunc("/count", promDemo.NameCount)
	http.HandleFunc("/base64", promDemo.Base64)
	http.HandleFunc("/health", promDemo.Health)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func newPrometheusDemo(redisAddr string) *PrometheusDemo {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &PrometheusDemo{
		redisClient: client,
	}
}

func (p *PrometheusDemo) AddToRedis(key string) error {
	val, err := p.redisClient.Get(key).Result()

	//Value is not found so add it
	if val == "" {
		err := p.redisClient.Set(key, "1", 0).Err()
		if err != nil {
			return err
		}
		return nil

	} else if err != nil {
		return err
	}
	intVal, err := strconv.Atoi(val)

	if err != nil {
		return fmt.Errorf("Value not an Integer")

	}

	// Increment count and update in redis
	newValue := intVal + 1
	err = p.redisClient.Set(key, newValue, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p *PrometheusDemo) GetName(key string) (string, error) {
	val, err := p.redisClient.Get(key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return val, nil
	}
	return "", nil
}

func (p *PrometheusDemo) HelloServer(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s\t\t%s\t\t",
		r.Method,
		r.RequestURI,
	)
	start := time.Now()
	keys, ok := r.URL.Query()["name"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'name' is missing")
		http.Error(w, "Url Param 'name' is missing", 500)
		requestCount.WithLabelValues("500", "/sayhi").Inc()
		requestTime := float64(time.Since(start).Nanoseconds()) / 1000000000.0
		requestLatency.WithLabelValues("500", "/sayhi").Observe(requestTime)
		return
	}
	data := keys[0]
	err := p.AddToRedis(data)
	if err != nil {
		http.Error(w, "Redis Error", 500)
		requestCount.WithLabelValues("500", "/sayhi").Inc()
		requestTime := float64(time.Since(start).Nanoseconds()) / 1000000000.0
		requestLatency.WithLabelValues("500", "/sayhi").Observe(requestTime)
		return

	}

	fmt.Fprintf(w, "Hello %v!", data)
	requestCount.WithLabelValues("200", "/sayhi").Inc()
	requestTime := float64(time.Since(start).Nanoseconds()) / 1000000000.0
	requestLatency.WithLabelValues("200", "/sayhi").Observe(requestTime)
}

func (p *PrometheusDemo) NameCount(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s\t\t%s\t\t",
		r.Method,
		r.RequestURI,
	)
	start := time.Now()
	keys, ok := r.URL.Query()["name"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'name' is missing")
		http.Error(w, "Url Param 'name' is missing", 500)
		requestCount.WithLabelValues("500", "/count").Inc()
		return
	}
	data := keys[0]
	count, err := p.GetName(data)
	if err != nil {
		http.Error(w, "Redis Error", 500)
		requestCount.WithLabelValues("500", "/count").Inc()
		requestTime := float64(time.Since(start).Nanoseconds()) / 1000000000.0
		requestLatency.WithLabelValues("500", "/count").Observe(requestTime)
		return

	}
	if count == "" {
		fmt.Fprintf(w, "%v: not found", data)

	} else {
		fmt.Fprintf(w, "%v: %v", data, count)

	}
	requestCount.WithLabelValues("200", "/count").Inc()
	requestTime := float64(time.Since(start).Nanoseconds()) / 1000000000.0
	requestLatency.WithLabelValues("200", "/count").Observe(requestTime)
}

func (p *PrometheusDemo) Health(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s\t\t%s\t\t",
		r.Method,
		r.RequestURI,
	)
	start := time.Now()
	w.Write([]byte("ok"))
	requestCount.WithLabelValues("200", "/health").Inc()
	requestTime := float64(time.Since(start).Nanoseconds()) / 1000000000.0
	requestLatency.WithLabelValues("200", "/health").Observe(requestTime)
}

func (p *PrometheusDemo) Base64(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s\t\t%s\t\t",
		r.Method,
		r.RequestURI,
	)
	start := time.Now()
	keys, ok := r.URL.Query()["data"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'data' is missing")
		http.Error(w, "Url Param 'data' is missing", 500)
		requestCount.WithLabelValues("500", "/base64").Inc()
		requestTime := float64(time.Since(start).Nanoseconds()) / 1000000000.0
		requestLatency.WithLabelValues("500", "/base64").Observe(requestTime)
		return
	}

	data := keys[0]
	encodedData := b64.StdEncoding.EncodeToString([]byte(data))
	w.Write([]byte(encodedData))
	requestCount.WithLabelValues("200", "/base64").Inc()
	requestTime := float64(time.Since(start).Nanoseconds()) / 1000000000.0
	requestLatency.WithLabelValues("200", "/base64").Observe(requestTime)
}
