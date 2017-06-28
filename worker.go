package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

// Worker enquee refresh cache to run between intervals
func Worker(done chan bool, db *bolt.DB) {
	interval, _ := strconv.Atoi(os.Getenv("CACHE_INTERVAL"))
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	quit := make(chan struct{})
	fmt.Print("Worker is up! \n")

	go func() {
		for {
			select {
			case <-ticker.C:
				_, mobs := GetUrls()
				// if len(customDomains) > len(mobs) {
				//  err := p.Signal(os.Interrupt)
				// 	os.Exit(1)
				// }

				refreshCache(mobs, db, false)
			case <-quit:
				ticker.Stop()
				// done <- true
				// return
			}
		}
	}()

	done <- true
}

// GetUrls serach to domains we must allow to be served
func GetUrls() (customDomains []string, mobs []Mobilization) {
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
		// fmt.Println(jd.UpdatedAt)
	}
	return customDomains, mobs
}

func refreshCache(mobs []Mobilization, db *bolt.DB, force bool) []*HttpResponse {
	for _, mob := range mobs {
		results := readCacheContent(mob, db, force)
		for _, result := range results {
			if result.response != nil {
				fmt.Printf("%s status: %s\n", result.url, result.response.Status)
			}
			time.Sleep(1e9)
		}
	}
	return nil
}

func readCacheContent(mob Mobilization, db *bolt.DB, force bool) []*HttpResponse {
	ch := make(chan *HttpResponse, 1) // buffered
	responses := []*HttpResponse{}
	interval, _ := strconv.ParseFloat(os.Getenv("CACHE_INTERVAL"), 64)

	// fmt.Printf("Checking if %s || %s <= %f \n", time.Now().Format("2006-01-02T15:04:05.000-07:00"), mob.UpdatedAt, interval*3.0)
	tUpdatedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", mob.UpdatedAt)
	if time.Now().Sub(tUpdatedAt).Seconds() <= interval*3.0 || force {
		go func(mob Mobilization) {
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
			fmt.Printf("\nWorker: Content from slug %s will be served as %s\n", mob.Slug, mob.CustomDomain)
			return nil
		})
		if err != nil {
			fmt.Errorf("error save content %s", err)
		}
	} else {
		fmt.Errorf("error read response body: %s", err)
	}
}
