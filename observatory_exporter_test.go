package main

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestScrape(t *testing.T) {
	c := NewCollector(DefaultApiURL)

	metrics, err := c.Scrape("ulfr.io", false)

	if err != nil {
		t.Fatalf("Scrape returned an error: %s", err)
	}

	expected := []string{
		"tls_enabled",
		"cert_is_trusted",
		"cert_expiry_date",
		"cert_start_date",
		"ssl_level",
		"score",
		"grade",
	}

	for _, k := range expected {
		if _, ok := metrics[k]; !ok {
			t.Fatalf("Missing metrics %s", k)
		}
	}
}

func readGauge(m prometheus.Metric) float64 {
	pb := &dto.Metric{}
	m.Write(pb)
	return pb.GetGauge().GetValue()
}

func TestMetricsExport(t *testing.T) {
	targetURL := "dummy-url.com"
	cache := NewCache()
	e := NewExporter(cache)

	tomorrow := float64(time.Now().Unix()) + (time.Hour * 24).Seconds()
	yesterday := float64(time.Now().Unix()) - (time.Hour * 24).Seconds()

	// ordering is important.
	metrics := Metrics{}
	metrics["cert_expiry_date"] = tomorrow
	metrics["cert_is_trusted"] = 0
	metrics["cert_start_date"] = yesterday
	metrics["grade"] = 3
	metrics["score"] = 85
	metrics["ssl_level"] = 1
	metrics["tls_enabled"] = 1

	cache.Write(targetURL, metrics)

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		e.Collect(ch)
	}()

	if expect, got := metrics["cert_expiry_date"], readGauge(<-ch); expect != got {
		t.Errorf("cert_expiry_date: expected %f, got %f", expect, got)
	}
	if expect, got := metrics["cert_is_trusted"], readGauge(<-ch); expect != got {
		t.Errorf("cert_is_trusted: expected %f, got %f", expect, got)
	}
	if expect, got := metrics["cert_start_date"], readGauge(<-ch); expect != got {
		t.Errorf("cert_start_date: expected %f, got %f", expect, got)
	}
	if expect, got := metrics["grade"], readGauge(<-ch); expect != got {
		t.Errorf("grade: expected %f, got %f", expect, got)
	}
	if expect, got := metrics["score"], readGauge(<-ch); expect != got {
		t.Errorf("score: expected %f, got %f", expect, got)
	}
	if expect, got := metrics["ssl_level"], readGauge(<-ch); expect != got {
		t.Errorf("ssl_level: expected %f, got %f", expect, got)
	}
	if expect, got := metrics["tls_enabled"], readGauge(<-ch); expect != got {
		t.Errorf("tls_enabled: expected %f, got %f", expect, got)
	}
}
