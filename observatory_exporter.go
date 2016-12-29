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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

// todo:
// other logger??
// yml config + CLA support...

// https://github.com/mozilla/tls-observatory

type Metrics map[string]float64

func registerSignals() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Print("Received SIGTERM, exiting...")
		os.Exit(1)
	}()
}

const (
	DefaultApiURL = "https://tls-observatory.services.mozilla.com/api/v1/"
)

func init() {
	prometheus.MustRegister(version.NewCollector("observatory_exporter"))
}

func main() {
	var (
		listenAddr  = flag.String("web.listen-address", ":9229", "The address to listen on for HTTP requests.")
		showVersion = flag.Bool("version", false, "Print version information")
		targetURL   = flag.String("observatory.target-url", "", "The URL checked via Observatory")
		apiURL      = flag.String("observatory.api-url", DefaultApiURL, "The Observatory API endpoint used.")
		// observatory allows rescans only every 3 minutes
		interval = flag.Int("observatory.interval", 60*60, "Interval used for running checks against the Observatory API")
	)

	registerSignals()

	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("observatory_exporter"))
		os.Exit(0)
	}

	if *targetURL == "" {
		log.Fatalf("No target url set.")
	}

	cached := Metrics{}
	collector := NewCollector(*targetURL, *apiURL)

	exporter := NewExporter(*targetURL, func() *Metrics { return &cached })
	prometheus.MustRegister(exporter)

	go func() {
		timer := time.NewTimer(time.Second * time.Duration(*interval))

		for {
			var err error
			cached, err = collector.Scrape(false)

			if err != nil {
				// do something!! but dont die please
				log.Printf("Failed to get result for %s: %s", *targetURL, err)
			} else {
				log.Printf("Updated result for %s", *targetURL)
			}

			<-timer.C
		}
	}()

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Observatory Exporter</title></head>
             <body>
             <h1>Observatory Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})
	http.ListenAndServe(*listenAddr, nil)
}
