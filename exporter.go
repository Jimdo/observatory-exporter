package main

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "observatory"
)

type Exporter struct {
	targetURL string
	mutex     sync.Mutex
	metrics   map[string]*prometheus.Desc
	fetch     func() *Metrics
}

func NewExporter(targetURL string, fetch func() *Metrics) *Exporter {
	targetLabel := prometheus.Labels{"target": targetURL}
	e := Exporter{
		targetURL: targetURL,
		fetch:     fetch,
		metrics: map[string]*prometheus.Desc{
			"tls_enabled":      prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "tls_enabled"), "TLS enabled for domain", nil, targetLabel),
			"is_valid":         prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "is_valid"), "Is 1 (aka 'valid') if any of the per truststore valitities is valid", nil, targetLabel),
			"ssl_level":        prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "ssl_level"), "Defines the Mozilla SSL level for given domain (old=0, intermediate=1, modern=2)", nil, targetLabel),
			"score":            prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "score"), "Defines the score given by Mozilla Observatory's mozillaGradingWorker (0...100)", nil, targetLabel),
			"grade":            prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "grade"), "Grade representation of score, A=4, B=3, C=2, D=1, F=0", nil, targetLabel),
			"cert_expiry_date": prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "cert_expiry_date"), "Expiry date for certificate.", nil, targetLabel),
			"cert_start_date":  prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "cert_start_date"), "Start date for certificate.", nil, targetLabel),
		},
	}
	return &e
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.metrics {
		ch <- m
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	log.Printf("Exporting result for %s", e.targetURL)

	data := e.fetch()
	for key, val := range *data {
		ch <- prometheus.MustNewConstMetric(e.metrics[key], prometheus.GaugeValue, val)
	}
}
