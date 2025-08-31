package server

import (
	"crypto_price/pkg/config"
	"crypto_price/pkg/controller"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func StartHTTPServer() {
    cfg := config.GetConfigs()

    registerMetrics()
    http.Handle("/metrics", promhttp.Handler())

    http.HandleFunc("/price", controller.HandlePriceRequest)

    addr := ":" + cfg.ServerPort
    log.Printf("Starting HTTP server on %s", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}



func registerMetrics() {
    goCollector := collectors.NewGoCollector()
    if err := prometheus.Register(goCollector); err != nil {
        if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
            log.Printf("Failed to register Go collector: %v", err)
            return
        }
    }
    log.Println("Prometheus metrics registered successfully")
}