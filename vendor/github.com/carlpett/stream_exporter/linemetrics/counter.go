package linemetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CounterVecLineMetric struct {
	BaseLineMetric
	metric prometheus.CounterVec
}

func (counter CounterVecLineMetric) MatchLine(s string) {
	matches := counter.pattern.FindStringSubmatch(s)
	if len(matches) > 0 {
		captures := matches[1:]
		counter.metric.WithLabelValues(captures...).Inc()
	}
}

type CounterLineMetric struct {
	BaseLineMetric
	metric prometheus.Counter
}

func (counter CounterLineMetric) MatchLine(s string) {
	matches := counter.pattern.MatchString(s)
	if matches {
		counter.metric.Inc()
	}
}

func NewCounterLineMetric(base BaseLineMetric) (LineMetric, prometheus.Collector) {
	opts := prometheus.CounterOpts{
		Name: base.name,
		Help: base.name,
	}
	var lineMetric LineMetric
	if len(base.labels) > 0 {
		metric := prometheus.NewCounterVec(opts, base.labels)
		lineMetric = CounterVecLineMetric{
			BaseLineMetric: base,
			metric:         *metric,
		}
		return lineMetric, metric
	} else {
		metric := prometheus.NewCounter(opts)
		lineMetric = CounterLineMetric{
			BaseLineMetric: base,
			metric:         metric,
		}
		return lineMetric, metric
	}
}
