package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

// CacheManager is responsible to manager mob content
type CacheManager struct {
	s Specification
	g *Graphql
	r *Redis
}

// HTTPResponse help handle output from requests
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

// Populate will load Mobilization Content and Save To Redis if necessary
func (c *CacheManager) Populate() {
	log.Println("[populate_cache] job started")
	_, mobs := c.GetAllowedDomains()

	for _, mob := range mobs {
		var cachedMob = c.r.ReadMobilization("cached_urls:" + mob.CustomDomain)
		tUpdatedAt, _ := time.Parse(
			"2006-01-02T15:04:05.000-07:00",
			mob.UpdatedAt)
		tCachedAt, _ := time.Parse(
			"2006-01-02T15:04:05.000-07:00",
			cachedMob.CachedAt)

		if string(cachedMob.Content) == "" {
			c.readHTMLFromHTTPAndSaveToRedis(mob, c.s)
		} else if time.Now().Sub(tCachedAt).Hours() >= 168.0 { // 7 days
			c.readHTMLFromHTTPAndSaveToRedis(mob, c.s)
		} else if time.Now().Sub(tUpdatedAt).Seconds() <= c.s.Interval {
			c.readHTMLFromHTTPAndSaveToRedis(mob, c.s)
		}
	}
}

// GetAllowedDomains search to domains we must allow to be public
func (c *CacheManager) GetAllowedDomains() (customDomains []string, mobs []Mobilization) {

	c.g.GetAllMobilizations()
	var q = c.g.queryAllMobilizations

	mobs = make([]Mobilization, 0)
	customDomains = make([]string, 0)

	for _, node := range q.AllMobilizations.Edges {
		var jd = node.Node
		jd.Public = false
		if jd.CustomDomain != "" {
			customDomains = append(customDomains, jd.CustomDomain)
			jd.Public = true
			mobs = append(mobs, jd)
		}
	}
	return customDomains, mobs
}

func (c *CacheManager) readHTMLFromHTTPAndSaveToRedis(mob Mobilization, s Specification) []*HTTPResponse {
	ch := make(chan *HTTPResponse, 1) // buffered
	responses := []*HTTPResponse{}

	resp, err := netClient.Get("http://" + mob.Slug + "." + s.Domain)
	// defer resp.Body.Close()

	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("%s", err)
		}
		mob.Content = body
		mob.CachedAt = time.Now().Format("2006-01-02T15:04:05.000-07:00")
		c.r.SaveMobilization("cached_urls:"+mob.CustomDomain, mob)
	} else {
		log.Printf("error read response http: %s", err)
	}
	ch <- &HTTPResponse{mob.CustomDomain, resp, err}

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
