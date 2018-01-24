package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/shurcooL/graphql"
)

func printJSON(v interface{}) {
	w := json.NewEncoder(os.Stdout)
	w.SetIndent("", "\t")
	err := w.Encode(v)
	if err != nil {
		panic(err)
	}
}

// GetUrls search to domains we must allow to be public
func GetUrls(s Specification) (customDomains []string, mobs []Mobilization) {

	var query struct {
		AllMobilizations struct {
			Edges []struct {
				Node   Mobilization
				Cursor graphql.String
			}
		} `graphql:"allMobilizations"`
	}

	err2 := client.Query(context.Background(), &query, nil)
	if err2 != nil {
		fmt.Println("Error querying api services: ", err2)
	}
	// printJSON(query)

	mobs = make([]Mobilization, 0)
	customDomains = make([]string, 0)

	for _, node := range query.AllMobilizations.Edges {
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

// Load Mobilization Content and Save To Redis
func populateCache(s Specification) {
	log.Println("[populateCache] job started")
	_, mobs := GetUrls(s)

	for _, mob := range mobs {
		var cachedMob = RedisReadMobilization("cached_urls:" + mob.CustomDomain)
		tUpdatedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", mob.UpdatedAt)
		tCachedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", cachedMob.CachedAt)

		if string(cachedMob.Content) == "" {
			readHTMLFromHTTPAndSaveToRedis(mob, s)
		} else if time.Now().Sub(tCachedAt).Hours() >= 168.0 { // 7 days
			readHTMLFromHTTPAndSaveToRedis(mob, s)
		} else if time.Now().Sub(tUpdatedAt).Seconds() <= s.Interval {
			readHTMLFromHTTPAndSaveToRedis(mob, s)
		}
	}
}

func readHTMLFromHTTPAndSaveToRedis(mob Mobilization, s Specification) []*HTTPResponse {
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
		RedisSaveMobilization("cached_urls:"+mob.CustomDomain, mob)
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
