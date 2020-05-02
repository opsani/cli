package main

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

const metricsPath string = "/metrics"
const subsystemName string = "demo"

// Metrics maintains the values to be emitted to Prometheus
type Metrics struct {
	requestCount     *prometheus.CounterVec
	requestDurations *prometheus.HistogramVec
}

// NewMiddleware returns a new Fiber middleware for tracking basic metrics to Prometheus
func NewMiddleware() func(*fiber.Ctx) {
	var metrics Metrics = Metrics{
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystemName,
				Name:      "requests_total",
				Help:      "The HTTP request counts processed.",
			},
			[]string{"code", "method"},
		),
		requestDurations: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystemName,
				Name:      "request_duration_seconds",
				Help:      "request latencies",
				Buckets:   []float64{.005, .01, .02, 0.04, .06, 0.08, .1, 0.15, .25, 0.4, .6, .8, 1, 1.5, 2, 3, 5},
			},
			[]string{"code", "path"},
		),
	}
	prometheus.MustRegister(metrics.requestCount, metrics.requestDurations)

	return func(ctx *fiber.Ctx) {
		if ctx.Path() == metricsPath {
			ctx.Next()
			return
		}

		// Instrument the request
		start := time.Now()
		ctx.Next()

		status := strconv.Itoa(ctx.Fasthttp.Response.StatusCode())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		ep := string(ctx.Method()) + "_" + ctx.OriginalURL()
		metrics.requestCount.WithLabelValues(status, string(ctx.Method())).Inc()
		metrics.requestDurations.WithLabelValues(status, ep).Observe(elapsed)
	}
}

func prometheusHandler() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
}

func main() {
	app := fiber.New()

	// Wrap all our endpoints with metrics instrumentation
	middleware := NewMiddleware()
	app.Use(middleware)

	// Wrap the Prometheus HTTP handler into FastHTTP and dispatch via Fiber
	// This registers a /metrics endpoint that will publish metrics for Prometheus
	handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	app.Get(metricsPath, func(c *fiber.Ctx) {
		handler(c.Fasthttp)
	})

	app.Get("/health", func(c *fiber.Ctx) {
		c.SendStatus(200)
		c.JSON(fiber.Map{
			"status": "pass",
		})
		log.Println(string(c.Path()))
	})

	app.Get("/", func(c *fiber.Ctx) {
		c.JSON(fiber.Map{
			"hello": "world",
		})
	})

	app.Listen(8080)
}
