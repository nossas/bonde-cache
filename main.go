package main

import (
	"log"

	"github.com/boltdb/bolt"
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
	ApiUrl   string
}

func main() {
	var s Specification
	err := envconfig.Process("cache", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	if s.Sync {
		syncRestoreCertificates(s)
		syncRestoreDb(s)
	}

	db, err := bolt.Open("./data/db/bonde-cache.db", 0666, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	go webRedirect(s)
	go webCache(db, s)

	worker(db, s)
	<-gocron.Start()
}
