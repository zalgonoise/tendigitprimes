package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Noop struct{}

func (m Noop) RegisterCollector(prometheus.Collector)                               {}
func (m Noop) InitRequestsMetrics(string, string)                                   {}
func (m Noop) IncRequestsReceivedTotal(string, string)                              {}
func (m Noop) IncRequestsReceivedErrored(string, string)                            {}
func (m Noop) ObserveRequestLatency(context.Context, string, string, time.Duration) {}
func (m Noop) SetDatabaseReadiness(bool)                                            {}
func (m Noop) Registry() (*prometheus.Registry, error)                              { return prometheus.NewRegistry(), nil }
