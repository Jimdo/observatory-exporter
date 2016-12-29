package main

import (
	"testing"
	"time"
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
