package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PingCollector struct {
	db             *sql.DB
	pingSuccessful *prometheus.Desc
	pingRttSeconds *prometheus.Desc
	timeout        time.Duration
	time           Time
}

type Time interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type realTime struct{}

func (r realTime) Now() time.Time {
	return time.Now()
}

func (r realTime) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func NewPingCollector(db *sql.DB, name string) *PingCollector {
	return &PingCollector{
		db: db,
		pingSuccessful: prometheus.NewDesc(
			"go_sql_ping_success",
			"Exports whether a connection to the database was possible",
			nil, prometheus.Labels{"db_name": name},
		),
		pingRttSeconds: prometheus.NewDesc(
			"go_sql_ping_rtt_seconds",
			"Round-trip-time for the database ping",
			nil, prometheus.Labels{"db_name": name},
		),
		timeout: 1 * time.Second,
		time:    realTime{},
	}
}

func (p PingCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- p.pingSuccessful
	descs <- p.pingRttSeconds
}

func (p PingCollector) Collect(metrics chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	startTime := p.time.Now()

	err := p.db.PingContext(ctx)

	rtt := p.time.Since(startTime)

	var success float64
	if err == nil {
		success = 1
	}

	metrics <- prometheus.MustNewConstMetric(p.pingSuccessful, prometheus.GaugeValue, success)
	metrics <- prometheus.MustNewConstMetric(p.pingRttSeconds, prometheus.GaugeValue, rtt.Seconds())
}
