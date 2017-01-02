package main

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestRequestScan(t *testing.T) {
	c := NewCollector("ulfr.io", DefaultApiURL)
	scanid, err := c.requestScan(false)
	if err != nil {
		t.Error("requestScan returned an error: ", err)
	}
	if scanid <= 0 {
		t.Error("invalid scanid ", scanid)
	}
}

func TestGetResult(t *testing.T) {
	c := NewCollector("ulfr.io", DefaultApiURL)

	scanid, err := c.requestScan(false)
	if err != nil {
		t.Fatal("requestScan returned an error: ", err)
	}

	res, err := c.getResult(scanid)
	if err != nil {
		t.Fatal("getResult returned an error: ", err)
	}
	if !res.Has_tls {
		t.Fatal("Oh no! No TLS")
	}
}

func TestGetCertificate(t *testing.T) {
	c := NewCollector("ulfr.io", DefaultApiURL)

	scanid, err := c.requestScan(false)
	if err != nil {
		t.Fatal("requestScan returned an error: ", err)
	}

	res, err := c.getResult(scanid)
	if err != nil {
		t.Fatal("getResult returned an error: ", err)
	}

	cert, err := c.getCertificate(res.Cert_id)
	if err != nil {
		t.Fatal("getCertificate returned an error: ", err)
	}

	now := time.Now().UTC()
	if now.After(cert.Validity.NotAfter) || now.Before(cert.Validity.NotBefore) {
		t.Fatal("certificate outdated")
	}
}

func readGauge(m prometheus.Metric) float64 {
	pb := &dto.Metric{}
	m.Write(pb)
	return pb.GetGauge().GetValue()
}

func TestMetricsExport(t *testing.T) {
	targetURL := "dummy-url.com"
	cached := Metrics{}
	e := NewExporter(targetURL, func() *Metrics { return &cached })

	tomorrow := float64(time.Now().Unix()) + (time.Hour * 24).Seconds()
	yesterday := float64(time.Now().Unix()) - (time.Hour * 24).Seconds()

	// ordering is important.
	cached["cert_expiry_date"] = tomorrow
	cached["cert_is_trusted"] = 0
	cached["cert_start_date"] = yesterday
	cached["grade"] = 3
	cached["score"] = 85
	cached["ssl_level"] = 1
	cached["tls_enabled"] = 1

	ch := make(chan prometheus.Metric)
	go func() {
		defer close(ch)
		e.Collect(ch)
	}()

	if expect, got := cached["cert_expiry_date"], readGauge(<-ch); expect != got {
		t.Errorf("cert_expiry_date: expected %f, got %f", expect, got)
	}
	if expect, got := cached["cert_is_trusted"], readGauge(<-ch); expect != got {
		t.Errorf("cert_is_trusted: expected %f, got %f", expect, got)
	}
	if expect, got := cached["cert_start_date"], readGauge(<-ch); expect != got {
		t.Errorf("cert_start_date: expected %f, got %f", expect, got)
	}
	if expect, got := cached["grade"], readGauge(<-ch); expect != got {
		t.Errorf("grade: expected %f, got %f", expect, got)
	}
	if expect, got := cached["score"], readGauge(<-ch); expect != got {
		t.Errorf("score: expected %f, got %f", expect, got)
	}
	if expect, got := cached["ssl_level"], readGauge(<-ch); expect != got {
		t.Errorf("ssl_level: expected %f, got %f", expect, got)
	}
	if expect, got := cached["tls_enabled"], readGauge(<-ch); expect != got {
		t.Errorf("tls_enabled: expected %f, got %f", expect, got)
	}
}
