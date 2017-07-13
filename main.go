package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kelseyhightower/envconfig"
)

type Mobilization struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Content      []byte
	CachedAt     time.Time
	Slug         string `json:"slug"`
	CustomDomain string `json:"custom_domain"`
	UpdatedAt    string `json:"updated_at"`
}

type HttpResponse struct {
	url      string
	response *http.Response
	err      error
}

type Specification struct {
	Dev      bool
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
		fmt.Errorf("open cache: %s", err)
	}

	finish := make(chan bool)
	done := make(chan bool, 1)

	if s.Reset {
		_, mobs := GetUrls()
		refreshCache(mobs, db, s) // force first time build cache
	}

	if !s.Dev {
		go ServerRedirect(s)
	}
	go ServerCache(db, s)
	go Worker(finish, db, s)
	<-done

	<-finish
	// defer db.Close()
}
