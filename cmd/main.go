package main

import (
	"crypto_price/pkg/server"
	"crypto_price/pkg/jobs"
)


func main(){
	go jobs.GetData()
	server.StartHTTPServer()
}