package server

import (
	"net/http"
	"log"
	"crypto_price/pkg/controller"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/client_golang/prometheus/collectors"

)


func StartHTTPServer() {
    registerMetrics()
    http.Handle("/metrics", promhttp.Handler())

    http.HandleFunc("/price", controller.HandlePriceRequest) 

    log.Println("Starting HTTP server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}



func registerMetrics() {
    goCollector := collectors.NewGoCollector()
    if err := prometheus.Register(goCollector); err != nil {
        if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
            panic(err) // Or handle the error in another appropriate way
        }
    }
    // Register other metrics similarly if needed
}