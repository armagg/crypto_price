package main

import (
	"crypto_price/pkg/config"
	"crypto_price/pkg/jobs"
	"crypto_price/pkg/server"
	"log"
	_ "net/http/pprof"

	"github.com/getsentry/sentry-go"
)


func main(){
	cfg := config.GetConfigs()

	if cfg.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			EnableTracing: true,
			Dsn: cfg.SentryDSN,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
		log.Println("Sentry initialized successfully")
	} else {
		log.Println("Sentry DSN not provided, skipping Sentry initialization")
	}

	go jobs.GetData()
	server.StartHTTPServer()
}