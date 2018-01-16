package main

import (
	"log"

	"github.com/jasonlvhit/gocron"
	"github.com/kelseyhightower/envconfig"
)

// Specification are enviroment variables
type Specification struct {
	Env      string
	Sync     bool
	Interval float64
	Port     string
	PortSsl  string
	Domain   string
	RedisUrl string
}

func main() {
	var s Specification
	err := envconfig.Process("cache", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	pool = newPool(s)

	go webRedirect(s)
	go webCache(s)

	worker(s)
	<-gocron.Start()
}
