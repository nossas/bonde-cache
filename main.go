package main

import (
	"log"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/jasonlvhit/gocron"
	"github.com/kelseyhightower/envconfig"
)

// Specification are enviroment variables
type Specification struct {
	Env             string
	Sync            bool
	Interval        float64
	Port            string
	PortSsl         string
	Domain          string
	RedisURL        string
	APIServiceURL   string
	APIServiceToken string
}

func main() {
	var s Specification

	err := envconfig.Process("cache", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	web := &Web{s: s}
	worker := &Worker{s: s}

	web.Setup()
	worker.Setup()

	gracehttp.Serve(web.server.Server, web.serverSSL.TLSServer)
	<-gocron.Start()
}
