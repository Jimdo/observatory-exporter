package main

import (
	"log"
	"sort"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "observatory"
)

type Exporter struct {
	cache   *Cache
	metrics map[string]*prometheus.Desc
}

func NewExporter(c *Cache) *Exporter {
	labels := []string{"target"}
	e := Exporter{
		cache: c,
		metrics: map[string]*prometheus.Desc{
			"tls_enabled":         prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "tls_enabled"), "TLS enabled for domain", labels, nil),
			"compatibility_level": prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "compatibility_level"), "Defines the Mozilla SSL compatibility level for given domain (bad=0, non compliant=1, old=2, intermediate=3, modern=4)", labels, nil),
			"score":               prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "score"), "Defines the score given by Mozilla Observatory's mozillaGradingWorker (0...100)", labels, nil),
			"grade":               prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "grade"), "Grade representation of score, A=4, B=3, C=2, D=1, F=0", labels, nil),
			"cert_is_trusted":     prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "cert_is_trusted"), "Is 1 (aka 'trusted') if certificate is known to be trusted (via truststores)", labels, nil),
			"cert_expiry_date":    prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "cert_expiry_date"), "Expiry date for certificate.", labels, nil),
			"cert_start_date":     prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "cert_start_date"), "Start date for certificate.", labels, nil),
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
	log.Print("Exporting result.")

	data := e.cache.ReadAll()

	for targetURL, metrics := range data {
		keys := []string{}
		for k := range metrics {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, key := range keys {
			ch <- prometheus.MustNewConstMetric(e.metrics[key], prometheus.GaugeValue, metrics[key], targetURL)
		}
	}
}
