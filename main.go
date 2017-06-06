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

	rice "github.com/GeertJohan/go.rice"
	"github.com/boltdb/bolt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"
)

type Mobilization struct {
	ID            int
	Name          string
	Slug          string
	Custom_Domain string
	UpdatedAt     time.Time
}

type HttpResponse struct {
	url      string
	response *http.Response
	err      error
}

func getUrls() (customDomains []string, mobs []Mobilization) {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get("https://api.bonde.org/mobilizations")
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var jsonData []Mobilization
	err = json.Unmarshal([]byte(jsonDataFromHTTP), &jsonData) // here!
	if err != nil {
		panic(err)
	}

	mobs = make([]Mobilization, 0)
	customDomains = make([]string, 0)
	for _, jd := range jsonData {
		if jd.Custom_Domain != "" {
			customDomains = append(customDomains, jd.Custom_Domain)
			mobs = append(mobs, jd)
		}
	}
	return customDomains, mobs
}

func refreshCache(mobs []Mobilization, db *bolt.DB, interval string) []*HttpResponse {
	i, _ := strconv.Atoi(interval)
	ticker := time.NewTicker(time.Duration(i) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				results := readCacheContent(mobs, db)
				for _, result := range results {
					if result.response != nil {
						fmt.Printf("%s status: %s\n", result.url, result.response.Status)
					}
					time.Sleep(1e9)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func readCacheContent(mobs []Mobilization, db *bolt.DB) []*HttpResponse {
	ch := make(chan *HttpResponse, len(mobs)) // buffered
	responses := []*HttpResponse{}

	for i, mob := range mobs {
		go func(mob Mobilization, i int) {
			fmt.Printf("fetch url %s \n", mob.Custom_Domain)
			var myClient = &http.Client{Timeout: 10 * time.Second}
			resp, err := myClient.Get("http://" + mob.Slug + ".bonde.org")

			// defer resp.Body.Close()
			if err == nil {
				saveCacheContent(mob, resp, db)
			} else {
				fmt.Errorf("error read response http: %s", err)
			}
			ch <- &HttpResponse{mob.Custom_Domain, resp, err}
		}(mob, i)
		time.Sleep(1e9)
	}

	for {
		select {
		case r := <-ch:
			fmt.Printf("fetched url %s\n", r.url)
			responses = append(responses, r)
			if len(responses) == len(mobs) {
				return responses
			}
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
		}
	}

	return responses
}

func saveCacheContent(mob Mobilization, resp *http.Response, db *bolt.DB) {
	body, err := ioutil.ReadAll(resp.Body)

	if err == nil {
		err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("cached_urls"))
			if err != nil {
				fmt.Errorf("error create cache bucket: %s", err)
			}
			b.Put([]byte(mob.Custom_Domain), body)
			fmt.Printf("saved body content url %s \n", mob.Slug)
			return nil
		})
		if err != nil {
			fmt.Errorf("error save content %s", err)
		}
	} else {
		fmt.Errorf("error read response body: %s", err)
	}
}

func main() {
	isdev, err := strconv.ParseBool(os.Getenv("IS_DEV"))

	db, err := bolt.Open("bonde-cache.db", 0600, nil)
	if err != nil {
		fmt.Errorf("open cache: %s", err)
	}

	customDomains, mobs := getUrls()
	readCacheContent(mobs, db) // force first time build cache
	refreshCache(mobs, db, os.Getenv("CACHE_INTERVAL"))

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 2}))
	e.Use(middleware.BodyLimit("1M"))

	assetHandler := http.FileServer(rice.MustFindBox("./public/").HTTPBox())
	e.GET("/dist/*", echo.WrapHandler(assetHandler))
	e.GET("/wysihtml/*", echo.WrapHandler(assetHandler))

	e.GET("/", func(c echo.Context) error {
		req := c.Request()
		host := req.Host
		if isdev {
			host, _, _ = net.SplitHostPort(host)
		}
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("cached_urls"))
			v := b.Get([]byte(host))
			c.HTML(http.StatusOK, string(v))
			return nil
		})
		if err != nil {
			fmt.Errorf("%s", err)
		}
		return nil
	})

	if isdev {
		e.Debug = true
		e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
	} else {
		e.Pre(middleware.RemoveTrailingSlash())
		e.Pre(middleware.WWWRedirect())
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(customDomains...)
		e.AutoTLSManager.Cache = autocert.DirCache("./cache/")
		e.AutoTLSManager.Email = "tech@nossas.org"

		finish := make(chan bool)
		go func() {
			ee := echo.New()
			ee.Pre(middleware.RemoveTrailingSlash())
			ee.Pre(middleware.HTTPSWWWRedirect())
			ee.Pre(middleware.HTTPSRedirect())
			e.Logger.Fatal(ee.Start(":80"))
		}()
		go func() {
			e.Logger.Fatal(e.StartAutoTLS(":443"))
		}()
		<-finish

	}
	defer db.Close()
}
