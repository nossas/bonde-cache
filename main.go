package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
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

func main() {
	isdev, err := strconv.ParseBool(os.Getenv("IS_DEV"))
	finish := make(chan bool)
	done := make(chan bool, 1)

	db, err := bolt.Open("bonde-cache.db", 0600, nil)
	if err != nil {
		fmt.Errorf("open cache: %s", err)
	}

	if os.Getenv("RESET_CACHE") == "true" {
		_, mobs := GetUrls()
		refreshCache(mobs, db, true) // force first time build cache
	}

	if !isdev {
		go ServerRedirect()
	}
	go ServerCache(db, isdev)
	go Worker(finish, db)
	<-done

	<-finish
	// defer db.Close()
}
