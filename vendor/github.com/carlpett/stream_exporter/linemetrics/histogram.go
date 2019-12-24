package linemetrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type HistogramLineMetric struct {
	BaseLineMetric
	valueIdx int
	metric   prometheus.Histogram
}

func (histogram HistogramLineMetric) MatchLine(s string) {
	matches := histogram.pattern.FindStringSubmatch(s)
	if len(matches) > 0 {
		captures := matches[1:]
		valueStr := captures[histogram.valueIdx]
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			log.Warnf("Unable to convert %s to float\n", valueStr)
			return
		}
		histogram.metric.Observe(value)
	}
}

type HistogramVecLineMetric struct {
	BaseLineMetric
	valueIdx int
	metric   prometheus.HistogramVec
}

func (histogram HistogramVecLineMetric) MatchLine(s string) {
	matches := histogram.pattern.FindStringSubmatch(s)
	if len(matches) > 0 {
		labels, value, err := getLabelsAndValue(matches, histogram.valueIdx)
		if err != nil {
			return
		}
		histogram.metric.WithLabelValues(labels...).Observe(value)
	}
}

func NewHistogramLineMetric(base BaseLineMetric, config HistogramConfig) (LineMetric, prometheus.Collector) {
	valueIdx, err := getValueCaptureIndex(base.labels)
	if err != nil {
		log.Fatalf("Error initializing histogram %s: %s", base.name, err)
	}
	base.labels = append(base.labels[:valueIdx], base.labels[valueIdx+1:]...)

	opts := prometheus.HistogramOpts{
		Name:    base.name,
		Help:    base.name,
		Buckets: config.Buckets,
	}
	var lineMetric LineMetric
	if len(base.labels) > 0 {
		metric := prometheus.NewHistogramVec(opts, base.labels)
		lineMetric = HistogramVecLineMetric{
			BaseLineMetric: base,
			valueIdx:       valueIdx,
			metric:         *metric,
		}
		return lineMetric, metric
	} else {
		metric := prometheus.NewHistogram(opts)
		lineMetric = HistogramLineMetric{
			BaseLineMetric: base,
			valueIdx:       valueIdx,
			metric:         metric,
		}
		return lineMetric, metric
	}
}
