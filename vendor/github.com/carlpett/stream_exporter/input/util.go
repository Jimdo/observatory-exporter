package input

import (
	"bytes"
	"fmt"
	"io"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/log"
)

func writeMetrics(w io.Writer) {
	metfam, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		log.Fatal(err)
	}
	out := &bytes.Buffer{}
	for _, met := range metfam {
		if _, err := expfmt.MetricFamilyToText(out, met); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprintln(w, out)
}
