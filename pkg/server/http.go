package server

import (
	"net/http"
	"log"
	"crypto_price/pkg/controller"
)


func StartHTTPServer() {
    http.HandleFunc("/price", controller.HandlePriceRequest) 

    log.Println("Starting HTTP server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}