package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel/trace"
)

const (
	// traceIDKey is used as the trace ID key value in the prometheus.Labels in a prometheus.Exemplar.
	//
	// Its value of `trace_id` complies with the OpenTelemetry specification for metrics' exemplars, as seen in:
	// https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars
	traceIDKey = "trace_id"
)

type Metrics struct {
	// Message broker metrics
	requestsReceivedTotal   *prometheus.CounterVec
	requestsReceivedErrored *prometheus.CounterVec
	requestsLatencySeconds  *prometheus.HistogramVec

	// Third party metrics
	collectors []prometheus.Collector
}

func (m *Metrics) RegisterCollector(collector prometheus.Collector) {
	m.collectors = append(m.collectors, collector)
}

func (m *Metrics) InitRequestsMetrics(minimum, maximum string) {
	m.requestsReceivedTotal.WithLabelValues(minimum, maximum)
	m.requestsReceivedErrored.WithLabelValues(minimum, maximum)
	m.requestsLatencySeconds.WithLabelValues(minimum, maximum)
}

func (m *Metrics) IncRequestsReceivedTotal(minimum, maximum string) {
	m.requestsReceivedTotal.WithLabelValues(minimum, maximum).Inc()
}

func (m *Metrics) IncRequestsReceivedErrored(minimum, maximum string) {
	m.requestsReceivedErrored.WithLabelValues(minimum, maximum).Inc()
}

func (m *Metrics) ObserveRequestLatency(ctx context.Context, minimum, maximum string, duration time.Duration) {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		//nolint:forcetypeassert // prometheus.ExemplarObserver is implemented in the underlying *prometheus.histogram type
		m.requestsLatencySeconds.
			WithLabelValues(minimum, maximum).(prometheus.ExemplarObserver).
			ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				traceIDKey: sc.TraceID().String(),
			})

		return
	}

	m.requestsLatencySeconds.WithLabelValues(minimum, maximum).Observe(duration.Seconds())
}

func (m *Metrics) Registry() (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()

	for _, metric := range []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			ReportErrors: false,
		}),
		m.requestsReceivedTotal,
		m.requestsReceivedErrored,
		m.requestsLatencySeconds,
	} {
		err := reg.Register(metric)
		if err != nil {
			return nil, err
		}
	}

	for _, metric := range m.collectors {
		err := reg.Register(metric)
		if err != nil {
			return nil, err
		}
	}

	return reg, nil
}

func NewMetrics() *Metrics {
	return &Metrics{
		requestsReceivedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "requests_received_total",
			Help: "Count of requests received",
		}, []string{"minimum", "maximum"}),
		requestsReceivedErrored: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "requests_failed_total",
			Help: "Count of all errored requests",
		}, []string{"minimum", "maximum"}),
		requestsLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "request_latency_seconds",
			Help:    "Histogram of request processing times",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"minimum", "maximum"}),
	}
}
