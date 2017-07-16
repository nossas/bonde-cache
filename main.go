package main

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/kelseyhightower/envconfig"
)

// Specification are enviroment variables
type Specification struct {
	Env      string
	Reset    bool
	Interval float64
	Port     string
	PortSsl  string
}

func main() {
	var s Specification
	err := envconfig.Process("cache", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := bolt.Open("bonde-cache.db", 0600, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("cached_urls"))
		return nil
	})

	finish := make(chan bool)
	done := make(chan bool, 1)

	go ServerRedirect(s)
	go ServerCache(db, s)
	go Worker(finish, db, s)
	<-done

	<-finish
	// defer db.Close()
}
