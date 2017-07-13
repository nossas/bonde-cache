package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

// Worker enquee refresh cache to run between intervals
func Worker(done chan bool, db *bolt.DB, s Specification) {
	ticker := time.NewTicker(time.Duration(s.Interval) * time.Second)
	quit := make(chan struct{})
	fmt.Print("Worker is up! \n")

	go func() {
		for {
			select {
			case <-ticker.C:
				_, mobs := GetUrls()
				for _, mob := range mobs {
					db.View(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte("cached_urls"))
						v := b.Get([]byte(mob.CustomDomain))

						if string(v) == "" {
							fmt.Printf("New domain created %s to %s at %s.\n", mob.CustomDomain, mob.Slug, mob.UpdatedAt)
							readCacheContent(mob, db, s)
							// os.Exit(1)
							// err := p.Signal(os.Interrupt)
						}
						return nil
					})
				}

				refreshCache(mobs, db, s)
			case <-quit:
				ticker.Stop()
				// done <- true
				return
			}
		}
	}()

	done <- true
}

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}
var netClient = &http.Client{
	Timeout:   time.Second * 10,
	Transport: netTransport,
}

// GetUrls serach to domains we must allow to be served
func GetUrls() (customDomains []string, mobs []Mobilization) {
	var jsonData []Mobilization

	var r, _ = netClient.Get("https://api.bonde.org/mobilizations")
	defer r.Body.Close()

	var jsonDataFromHTTP, _ = ioutil.ReadAll(r.Body)
	json.Unmarshal([]byte(jsonDataFromHTTP), &jsonData) // here!

	mobs = make([]Mobilization, 0)
	customDomains = make([]string, 0)

	for _, jd := range jsonData {
		if jd.CustomDomain != "" {
			customDomains = append(customDomains, jd.CustomDomain)
			mobs = append(mobs, jd)
		}
	}
	return customDomains, mobs
}

func refreshCache(mobs []Mobilization, db *bolt.DB, s Specification) []*HttpResponse {
	for _, mob := range mobs {
		results := readCacheContent(mob, db, s)
		for _, result := range results {
			if result.response != nil {
				fmt.Printf("%s status: %s\n", result.url, result.response.Status)
			}
			time.Sleep(1e9)
		}
	}
	return nil
}

func readCacheContent(mob Mobilization, db *bolt.DB, s Specification) []*HttpResponse {
	ch := make(chan *HttpResponse, 1) // buffered
	responses := []*HttpResponse{}

	// fmt.Printf("Checking if %s || %s <= %f \n", time.Now().Format("2006-01-02T15:04:05.000-07:00"), mob.UpdatedAt, interval*3.0)
	tUpdatedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", mob.UpdatedAt)
	if time.Now().Sub(tUpdatedAt).Seconds() <= s.Interval*3.0 || s.Reset {
		go func(mob Mobilization) {
			resp, err := netClient.Get("http://" + mob.Slug + ".bonde.org")
			// defer resp.Body.Close()

			if err == nil {
				saveCacheContent(mob, resp, db)
			} else {
				fmt.Errorf("error read response http: %s", err)
			}
			ch <- &HttpResponse{mob.CustomDomain, resp, err}
		}(mob)
		time.Sleep(1e9)
	}

	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			return responses
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
			return nil
		}
	}

	// return responses
}

func saveCacheContent(mob Mobilization, resp *http.Response, db *bolt.DB) {
	body, err := ioutil.ReadAll(resp.Body)

	if err == nil {
		err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("cached_urls"))
			if err != nil {
				fmt.Errorf("error create cache bucket: %s", err)
			}
			mob.Content = body
			mob.CachedAt = time.Now()
			encoded, err := json.Marshal(mob)
			if err != nil {
				return err
			}

			b.Put([]byte(mob.CustomDomain), encoded)
			fmt.Printf("\nWorker: [cached in %s] from www.%s.bonde.org and served at %s \n", mob.CachedAt, mob.Slug, mob.CustomDomain)
			return nil
		})
		if err != nil {
			fmt.Errorf("error save content %s", err)
		}
	} else {
		fmt.Errorf("error read response body: %s", err)
	}
}
