package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/jasonlvhit/gocron"
)

// Mobilization is saved as columns into to file
type Mobilization struct {
	ID           int    `json:"id" redis:"id"`
	Name         string `json:"name" redis:"name"`
	Content      []byte `json: "content" redis:"content"`
	CachedAt     string `json:"cached_at" redis:"cached_at"`
	Slug         string `json:"slug" redis:"slug"`
	CustomDomain string `json:"custom_domain" redis:"custom_domain"`
	UpdatedAt    string `json:"updated_at" redis:"updated_at"`
	Public       bool   `json:"public" redis:"public"`
}

// HTTPResponse helper handle output from requests
type HTTPResponse struct {
	url      string
	response *http.Response
	err      error
}

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

var netClient = &http.Client{
	Timeout:   time.Second * 20,
	Transport: netTransport,
}

func worker(s Specification) {
	if s.Sync {
		syncRestoreCertificates(s)
	}
	populateCache(s)
	checkNewCertificates(s)
	gocron.Every(30).Seconds().Do(populateCache, s)
	gocron.Every(30).Seconds().Do(checkNewCertificates, s)
}

func checkNewCertificates(s Specification) {
	log.Println("[checkNewCertificates] job started")
	_, mobs := GetUrls(s)
	for _, mob := range mobs {
		var cachedMob = redisRead("cached_urls:" + mob.CustomDomain)
		if string(cachedMob.Name) == "" {
			log.Println("[checkNewCertificate] NEW CERT FOUND")
		}

	}
	// time.Sleep(30 * time.Second)
	// pid := os.Getpid()
	// proc, _ := os.FindProcess(pid)
	// proc.Signal(os.Interrupt)
}

func populateCache(s Specification) {
	log.Println("[populateCache] job started")
	_, mobs := GetUrls(s)

	for _, mob := range mobs {
		var cachedMob = redisRead("cached_urls:" + mob.CustomDomain)
		tUpdatedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", mob.UpdatedAt)
		tCachedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", cachedMob.CachedAt)

		if string(cachedMob.Content) == "" {
			writeOriginToCache(mob, s)
		} else if time.Now().Sub(tCachedAt).Hours() >= 168.0 { // 7 days
			writeOriginToCache(mob, s)
		} else if time.Now().Sub(tUpdatedAt).Seconds() <= s.Interval {
			writeOriginToCache(mob, s)
		}
	}
}

// GetUrls serach to domains we must allow to be served
func GetUrls(s Specification) (customDomains []string, mobs []Mobilization) {
	var jsonData []Mobilization

	var r, err = netClient.Get("https://api." + s.Domain + "/mobilizations")
	if err != nil {
		log.Println("[worker] couldn't reach api server")
	}
	defer r.Body.Close()

	var jsonDataFromHTTP, _ = ioutil.ReadAll(r.Body)
	json.Unmarshal([]byte(jsonDataFromHTTP), &jsonData) // here!

	mobs = make([]Mobilization, 0)
	customDomains = make([]string, 0)

	for _, jd := range jsonData {
		jd.Public = false
		if jd.CustomDomain != "" {
			customDomains = append(customDomains, jd.CustomDomain)
			jd.Public = true
			mobs = append(mobs, jd)
		}
	}
	return customDomains, mobs
}

func writeOriginToCache(mob Mobilization, s Specification) []*HTTPResponse {
	results := readOriginContent(mob, s)
	for _, result := range results {
		if result.response != nil {
			log.Printf("[worker] updated cache to %s, http status code: %s", result.url, result.response.Status)
		}
		time.Sleep(1e9)
	}
	return nil
}

func readOriginContent(mob Mobilization, s Specification) []*HTTPResponse {
	ch := make(chan *HTTPResponse, 1) // buffered
	responses := []*HTTPResponse{}

	go func(mob Mobilization) {
		resp, err := netClient.Get("http://" + mob.Slug + "." + s.Domain)
		// defer resp.Body.Close()

		if err == nil {
			saveCacheContent(mob, resp)
		} else {
			log.Printf("error read response http: %s", err)
		}
		ch <- &HTTPResponse{mob.CustomDomain, resp, err}
	}(mob)
	time.Sleep(1e9)
	// }

	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			return responses
		case <-time.After(50 * time.Millisecond):
			// log.Printf(".")
			return nil
		}
	}
}

func saveCacheContent(mob Mobilization, resp *http.Response) {
	body, err := ioutil.ReadAll(resp.Body)

	if err == nil {
		mob.Content = body
		mob.CachedAt = time.Now().Format("2006-01-02T15:04:05.000-07:00")
		// encoded, err2 := json.Marshal(mob)
		// if err2 != nil {
		// 	log.Printf("[worker] cache can't decode mob %s ", err)
		// }
		redisSave("cached_urls:"+mob.CustomDomain, mob)
	} else {
		log.Printf("error read response body: %s", err)
	}
}
