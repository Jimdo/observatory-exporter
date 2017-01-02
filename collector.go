package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"log"

	"github.com/mozilla/tls-observatory/certificate"
	"github.com/mozilla/tls-observatory/database"
)

type Scan struct {
	ID int64 `json:"scan_id"`
}

type MozillaEvalData struct {
	Level string `json:"level"`
}

type MozillaGradeData struct {
	Score       float64 `json:"grade"`
	LetterGrade string  `json:"lettergrade"`
}

type Collector struct {
	ApiURL string
	client *http.Client
	mu     sync.Mutex
}

func NewCollector(apiURL string) *Collector {
	c := &Collector{}

	c.ApiURL = strings.TrimSuffix(apiURL, "/")

	c.client = &http.Client{
		Timeout: time.Second * 10,
	}

	return c
}

func (c *Collector) Scrape(targetURL string, enforceRescan bool) (Metrics, error) {
	scanID, err := c.requestScan(targetURL, enforceRescan)
	if err != nil {
		return nil, err
	}

	scan, err := c.getResult(targetURL, scanID)
	if err != nil {
		return nil, err
	}

	cert, err := c.getCertificate(targetURL, scan.Cert_id)
	if err != nil {
		return nil, err
	}

	metrics := exportMetrics(scan, cert)
	return metrics, nil
}

func (c *Collector) requestScan(targetURL string, enforceRescan bool) (int64, error) {
	apiURL := c.ApiURL + "/scan"

	prms := url.Values{}
	prms.Add("target", targetURL)
	if enforceRescan {
		prms.Add("rescan", "true")
	}

	resp, err := c.client.PostForm(apiURL, prms)

	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, errors.New(http.StatusText(resp.StatusCode))
	}

	buf, _ := ioutil.ReadAll(resp.Body)
	var scan Scan
	err = json.Unmarshal(buf, &scan)

	if err != nil {
		return -1, err
	}

	return scan.ID, err
}

func (c *Collector) getResult(targetURL string, scanid int64) (*database.Scan, error) {
	var res database.Scan

	apiURL := fmt.Sprintf("%s/results?id=%d", c.ApiURL, scanid)

	// Try to get the result from observatory for n seconds until
	// we call this a failed attempt.
	endtime := time.Now().Add(time.Second * 120)

	for {
		resp, err := c.client.Get(apiURL)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &res, errors.New(http.StatusText(resp.StatusCode))
		}

		buf, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(buf, &res)

		if err != nil {
			return &res, err
		}

		if res.Complperc < 100 {
			if time.Now().After(endtime) {
				return nil, fmt.Errorf("Failed to retrieve results in time for %s", targetURL)
			}

			fmt.Print(".")
			time.Sleep(time.Second * 1)
			continue
		}

		break
	}

	return &res, nil
}

func (c *Collector) getCertificate(targetURL string, certid int64) (*certificate.Certificate, error) {
	apiURL := fmt.Sprintf("%s/certificate?id=%d", c.ApiURL, certid)

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to access certificate. HTTP %d: %s", resp.StatusCode, targetURL)
	}

	buf, _ := ioutil.ReadAll(resp.Body)

	var cert certificate.Certificate
	err = json.Unmarshal(buf, &cert)

	if err != nil {
		return nil, err
	}

	return &cert, nil
}

func exportMetrics(scan *database.Scan, cert *certificate.Certificate) (res Metrics) {
	res = Metrics{
		"tls_enabled":      boolToFloat(scan.Has_tls),
		"cert_is_trusted":  boolToFloat(scan.Is_valid),
		"cert_expiry_date": float64(cert.Validity.NotAfter.Unix()),
		"cert_start_date":  float64(cert.Validity.NotBefore.Unix()),
	}

	for _, a := range scan.AnalysisResults {
		if a.Success {
			switch a.Analyzer {
			case "mozillaEvaluationWorker":
				var d MozillaEvalData
				err := json.Unmarshal(a.Result, &d)
				if err != nil {
					log.Printf("Failed to unmarshal analyzer 'mozillaEvaluationWorker': %s", err)
					continue
				}
				res["ssl_level"] = levelToInt(d.Level)

			case "mozillaGradingWorker":
				var d MozillaGradeData
				err := json.Unmarshal(a.Result, &d)
				if err != nil {
					log.Printf("Failed to unmarshal analyzer 'mozillaGradingWorker': %s", err)
					continue
				}
				res["score"] = float64(d.Score)
				res["grade"] = gradeLetterToInt(d.LetterGrade)
			}
		}
	}

	return
}

func levelToInt(str string) float64 {
	mapping := map[string]float64{
		"bad":           0,
		"non compliant": 1,
		"old":           2,
		"intermediate":  3,
		"modern":        4,
	}

	str = strings.ToLower(str)
	l, ok := mapping[str]
	if !ok {
		l = -1
	}
	return l
}

func gradeLetterToInt(str string) float64 {
	mapping := map[string]float64{
		"A": 4,
		"B": 3,
		"C": 2,
		"D": 1,
		"F": 0,
	}

	str = strings.ToUpper(str)
	l, ok := mapping[str]
	if !ok {
		l = 0
	}
	return l
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
