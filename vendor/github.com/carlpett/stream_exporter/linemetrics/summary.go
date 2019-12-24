package linemetrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type SummaryLineMetric struct {
	BaseLineMetric
	valueIdx int
	metric   prometheus.Summary
}

func (summary SummaryLineMetric) MatchLine(s string) {
	matches := summary.pattern.FindStringSubmatch(s)
	if len(matches) > 0 {
		captures := matches[1:]
		valueStr := captures[summary.valueIdx]
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			log.Warnf("Unable to convert %s to float\n", valueStr)
			return
		}
		summary.metric.Observe(value)
	}
}

type SummaryVecLineMetric struct {
	BaseLineMetric
	valueIdx int
	metric   prometheus.SummaryVec
}

func (summary SummaryVecLineMetric) MatchLine(s string) {
	matches := summary.pattern.FindStringSubmatch(s)
	if len(matches) > 0 {
		labels, value, err := getLabelsAndValue(matches, summary.valueIdx)
		if err != nil {
			return
		}
		summary.metric.WithLabelValues(labels...).Observe(value)
	}
}

func NewSummaryLineMetric(base BaseLineMetric, config SummaryConfig) (LineMetric, prometheus.Collector) {
	valueIdx, err := getValueCaptureIndex(base.labels)
	if err != nil {
		log.Fatalf("Error initializing summary %s: %s", base.name, err)
	}
	base.labels = append(base.labels[:valueIdx], base.labels[valueIdx+1:]...)

	opts := prometheus.SummaryOpts{
		Name:       base.name,
		Help:       base.name,
		AgeBuckets: config.AgeBuckets,
		BufCap:     config.BufCap,
		MaxAge:     config.MaxAge,
		Objectives: config.Objectives,
	}
	var lineMetric LineMetric
	if len(base.labels) > 0 {
		metric := prometheus.NewSummaryVec(opts, base.labels)
		lineMetric = SummaryVecLineMetric{
			BaseLineMetric: base,
			valueIdx:       valueIdx,
			metric:         *metric,
		}
		return lineMetric, metric
	} else {
		metric := prometheus.NewSummary(opts)
		lineMetric = SummaryLineMetric{
			BaseLineMetric: base,
			valueIdx:       valueIdx,
			metric:         metric,
		}
		return lineMetric, metric
	}
}
