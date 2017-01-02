package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

type Metrics map[string]float64

const (
	DefaultApiURL = "https://tls-observatory.services.mozilla.com/api/v1/"
)

// CLI string array args
type arrayArgs []string

func (i *arrayArgs) String() string {
	return strings.Join(*i, " ")
}

func (i *arrayArgs) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func init() {
	prometheus.MustRegister(version.NewCollector("observatory_exporter"))
}

func main() {
	var (
		listenAddr  = flag.String("web.listen-address", ":9229", "The address to listen on for HTTP requests.")
		showVersion = flag.Bool("version", false, "Print version information")
		apiURL      = flag.String("observatory.api-url", DefaultApiURL, "The Observatory API endpoint used.")
		interval    = flag.Int("observatory.interval", 60*60, "Interval used for running checks against the Observatory API")
	)

	var targetURLs arrayArgs
	flag.Var(&targetURLs, "observatory.target-url", "The URL checked via Observatory")

	registerSignals()

	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("observatory_exporter"))
		os.Exit(0)
	}

	if len(targetURLs) == 0 {
		log.Fatalf("No target url set.")
	}

	targetURLs = sanitizeURLs(targetURLs)

	mux := http.NewServeMux()

	cache := NewCache()
	collector := NewCollector(*apiURL)

	exporter := NewExporter(cache)
	prometheus.MustRegister(exporter)

	go func() {
		timer := time.NewTimer(time.Second * time.Duration(*interval))

		for {
			for _, targetURL := range targetURLs {
				go func(targetURL string) {
					// We always try to rescan 'targetURL' and limit 'interval' to the
					// limits from Observatory
					// (see https://github.com/mozilla/tls-observatory#post-/api/v1/scan)
					// In case we still hit the limit (restart, someone else checking the
					// target) we will initiate a scrape without a rescan to get valid data.
					var err error
					var result Metrics

					result, err = collector.Scrape(targetURL, true)

					if err != nil && err.Error() == http.StatusText(http.StatusTooManyRequests) {
						result, err = collector.Scrape(targetURL, false)
					}

					if err == nil {
						cache.Write(targetURL, result)
						log.Printf("Updated result for %s", targetURL)
					} else {
						log.Printf("Failed to get result for %s: %s", targetURL, err)
					}
				}(targetURL)
			}
			<-timer.C
		}
	}()

	mux.Handle("/metrics", prometheus.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Observatory Exporter</title></head>
             <body>
             <h1>Observatory Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})
	http.ListenAndServe(*listenAddr, mux)
}

func registerSignals() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Print("Received SIGTERM, exiting...")
		os.Exit(1)
	}()
}

func sanitizeURLs(urls []string) []string {
	var results []string

	for _, url := range urls {
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")
		results = append(results, url)
	}

	return results
}
