# observatory-exporter [![Build Status](https://travis-ci.com/Jimdo/observatory-exporter.svg?token=1djnvUyMgtcVefCz54T4&branch=master)](https://travis-ci.com/Jimdo/observatory-exporter)
Mozilla Observatory Exporter for Prometheus

This is a simple server that calls Mozilla Observatory for given target URLs and exports them via HTTP/JSON for
Prometheus consumption.

### Build
```
make
```

### Run
```
./observatory-exporter --observatory.target-url=google.de --observatory.target-url=google.com
```

### Docker
You can deploy this exporter using the [jimdo/observatory-exporter](https://hub.docker.com/r/jimdo/observatory-exporter/) Docker Image.

Example
```
docker pull jimdo/observatory-exporter
docker run -p 9229:9229 observatory-exporter:latest --observatory.target-url google.de
```

## Exposed metrics
Name | Description
-----|-----
observatory_cert_expiry_date | Expiry date for certificate.
observatory_cert_is_trusted | Is 1 (aka 'trusted') if certificate is known to be trusted (via truststores)
observatory_cert_start_date | Start date for certificate.
observatory_compatibility_level | Defines the Mozilla SSL compatibility level for given domain (bad=0, non compliant=1, old=2, intermediate=3, modern=4)
observatory_grade | Grade representation of score, A=4, B=3, C=2, D=1, F=0
observatory_score | Defines the score given by Mozilla Observatory's mozillaGradingWorker (0...100)
observatory_tls_enabled | TLS enabled for domain

## Further reading on Mozilla Observatory
* https://observatory.mozilla.org
* https://github.com/mozilla/tls-observatory
