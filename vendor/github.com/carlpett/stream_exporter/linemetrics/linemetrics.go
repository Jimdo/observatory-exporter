package linemetrics

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v2"
)

type LineMetric interface {
	MatchLine(s string)
	Name() string
}

type BaseLineMetric struct {
	name    string
	pattern regexp.Regexp
	labels  []string
}

func (m BaseLineMetric) Name() string {
	return m.name
}

type config struct {
	Metrics []MetricsConfig
}

func ReadPatternConfig(path string) ([]MetricsConfig, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return config.Metrics, nil
}

func NewLineMetric(config MetricsConfig) (LineMetric, prometheus.Collector) {
	pattern := regexp.MustCompile(config.Pattern)
	labels := pattern.SubexpNames()[1:] // First element is entire expression

	var lineMetric LineMetric
	base := BaseLineMetric{
		name:    config.Name,
		pattern: *pattern,
		labels:  labels,
	}
	var collector prometheus.Collector
	switch config.Kind {
	case counter:
		lineMetric, collector = NewCounterLineMetric(base)
	case gauge:
		lineMetric, collector = NewGaugeLineMetric(base)
	case histogram:
		lineMetric, collector = NewHistogramLineMetric(base, config.HistogramConfig)
	case summary:
		lineMetric, collector = NewSummaryLineMetric(base, config.SummaryConfig)
	}

	return lineMetric, collector
}

func getValueCaptureIndex(labels []string) (int, error) {
	foundValue := false
	valueIdx := 0
	for idx, l := range labels {
		if l == "value" {
			foundValue = true
			valueIdx = idx
			break
		}
	}
	if !foundValue {
		return valueIdx, errors.New("No named capture group for 'value'")
	}

	return valueIdx, nil
}

func getLabelsAndValue(matches []string, valueIdx int) ([]string, float64, error) {
	captures := matches[1:]
	valueStr := captures[valueIdx]
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Warnf("Unable to convert %s to float\n", valueStr)
		return nil, 0, err
	}
	labels := append(captures[0:valueIdx], captures[valueIdx+1:]...)
	return labels, value, nil
}
