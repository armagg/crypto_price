package main

import (
	_ "net/http/pprof"
	"crypto_price/pkg/server"
	"crypto_price/pkg/jobs"
	"github.com/getsentry/sentry-go"
)


func main(){
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://4c2fbcc74e453c8e558c3b952b494c59@sentry.hamravesh.com/7641",
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	  })
	  if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	  }
	go jobs.GetData()
	server.StartHTTPServer()
}