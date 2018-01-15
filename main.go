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
	Reset    bool
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

	// restoreCertificates(s)

	// if !s.Reset {
	// 	restoreDb(s)
	// }

	db, err := bolt.Open("./data/db/bonde-cache.db", 0600, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	gocron.Every(60).Seconds().Do(populateCache, db, s)

	// if s.Reset {
	// 	_, mobs := GetUrls(s)
	// 	refreshCache(mobs, db, s)
	// }

	// finish := make(chan bool)
	// done := make(chan bool, 1)

	go ServerRedirect(s)
	go ServerCache(db, s)
	// go Worker(finish, db, s)

	// <-finish
	<-gocron.Start()
	// <-done
}
