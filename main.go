package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/common/log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func healthz(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(response, "ok")
}

func main() {
	c, err := NewCollector("https://dadata.ru/api/v2", os.Getenv("DADATA_TOKEN"), os.Getenv("DADATA_X_SECRET"))
	if err != nil {
		log.Fatalf("Can't create collector: %v", err)
	}
	prometheus.MustRegister(c)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Dadata Exporter</title></head>
			<body>
			<h1>Dadata Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})
	log.Infoln("Starting dadata-exporter")
	log.Fatal(http.ListenAndServe(":9501", nil))
}
