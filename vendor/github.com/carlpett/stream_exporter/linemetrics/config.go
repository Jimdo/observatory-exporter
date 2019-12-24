package linemetrics

import (
	"time"
)

type metricKind string

const (
	counter   metricKind = "counter"
	gauge     metricKind = "gauge"
	histogram metricKind = "histogram"
	summary   metricKind = "summary"
)

type MetricsConfig struct {
	Name    string
	Kind    metricKind
	Pattern string

	HistogramConfig `yaml:",inline"`
	SummaryConfig   `yaml:",inline"`
}

type HistogramConfig struct {
	Buckets []float64
}

type SummaryConfig struct {
	Objectives map[float64]float64
	MaxAge     time.Duration `yaml:"max_age"`
	AgeBuckets uint32        `yaml:"age_buckets"`
	BufCap     uint32        `yaml:"buf_cap"`
}
