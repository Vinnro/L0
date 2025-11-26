package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OrdersProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_processed_total",
			Help: "Total number of successfully processed orders",
		})

	OrdersErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_errors_total",
			Help: "Number of errors during order processing",
		})

	RetryCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_retry_total",
			Help: "How many messages were sent to retry",
		})

	DlqCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_dlq_total",
			Help: "How many messages were sent to DLQ",
		})

	ProcessingTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "order_processing_seconds",
			Help:    "Time spent processing a single order",
			Buckets: prometheus.LinearBuckets(0.01, 0.05, 20),
		})
)

func Register() {
	prometheus.MustRegister(OrdersProcessed)
	prometheus.MustRegister(OrdersErrors)
	prometheus.MustRegister(RetryCount)
	prometheus.MustRegister(DlqCount)
	prometheus.MustRegister(ProcessingTime)
}
