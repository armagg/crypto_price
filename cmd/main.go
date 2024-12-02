package main

import (
	_ "net/http/pprof"
	"crypto_price/pkg/server"
	"crypto_price/pkg/jobs"
	"github.com/getsentry/sentry-go"
	"log"
)


func main(){
	err := sentry.Init(sentry.ClientOptions{
		EnableTracing: true,
		Dsn: "https://4c2fbcc74e453c8e558c3b952b494c59@sentry.hamravesh.com/7641",
		TracesSampleRate: 1.0,
		ProfilesSampleRate: 1.0,
	  })
	  if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	  }
	go jobs.GetData()
	server.StartHTTPServer()
}