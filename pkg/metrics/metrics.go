package metrics

import "github.com/prometheus/client_golang/prometheus"

var HTTPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "http_requests_total",
	Help: "Total HTTP requests",
}, []string{"method", "path", "status"})

var HTTPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_request_duration_seconds",
	Help: "HTTP request duration in seconds",
	Buckets: prometheus.DefBuckets,
}, []string{"method", "path", "status"})

var KafkaMessagesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "kafka_messages_processed_total",
	Help: "Total kafka messages processed",
}, []string{"topic", "status"})

var KafkaMessageDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "kafka_message_duration_seconds",
	Help: "Kafka message duration in seconds",
	Buckets: prometheus.DefBuckets,
}, []string{"topic"})

func Register() {
	prometheus.MustRegister(HTTPRequestsTotal, HTTPRequestDuration, KafkaMessagesTotal, KafkaMessageDuration)
}