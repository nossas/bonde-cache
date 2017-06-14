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
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"
)

type Mobilization struct {
	ID            int
	Name          string
	Content       []byte
	CachedAt      time.Time
	Slug          string
	Custom_Domain string
	Updated_At    string
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
		fmt.Println(err)
	}
	defer r.Body.Close()

	jsonDataFromHTTP, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	var jsonData []Mobilization
	err = json.Unmarshal([]byte(jsonDataFromHTTP), &jsonData) // here!
	if err != nil {
		fmt.Println(err)
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

func enqueueCache(mobs []Mobilization, db *bolt.DB, force bool) []*HttpResponse {
	interval, _ := strconv.Atoi(os.Getenv("CACHE_INTERVAL"))
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// customDomains, _ := getUrls()
				// if len(customDomains) > len(mobs) {
				// 	os.Exit(1)
				// }

				refreshCache(mobs, db, force)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return nil
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
	tUpdatedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", mob.Updated_At)

	if time.Now().Sub(tUpdatedAt.UTC()).Seconds() <= interval*3.0 || force {
		go func(mob Mobilization) {
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
		}(mob)
		time.Sleep(1e9)
	}

	for {
		select {
		case r := <-ch:
			fmt.Printf("fetched url %s\n", r.url)
			responses = append(responses, r)
			return responses
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
			return nil
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
			mob.Content = body
			mob.CachedAt = time.Now()
			encoded, err := json.Marshal(mob)
			if err != nil {
				return err
			}

			b.Put([]byte(mob.Custom_Domain), encoded)
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

// StartAutoTLS starts an HTTPS server using certificates automatically installed from https://letsencrypt.org.
// func (e *Echo) MyStartAutoTLS(address string) error {
// 	s := e.TLSServer
// 	cfg := &tls.Config{
// 		MinVersion:               tls.VersionTLS12,
// 		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
// 		PreferServerCipherSuites: true,
// 		CipherSuites: []uint16{
// 			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
// 			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
// 			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
// 			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
// 		},
// 	}
// 	s.TLSConfig = cfg
// 	s.TLSConfig.GetCertificate = e.AutoTLSManager.GetCertificate
// 	return e.startTLS(address)
// }

func main() {
	isdev, err := strconv.ParseBool(os.Getenv("IS_DEV"))

	db, err := bolt.Open("bonde-cache.db", 0600, nil)
	if err != nil {
		fmt.Errorf("open cache: %s", err)
	}

	CustomHTTPErrorHandler := func(err error, c echo.Context) {
		req := c.Request()
		host := req.Host

		// code := http.StatusInternalServerError
		// if he, ok := err.(*echo.HTTPError); ok {
		// 	code = he.Code
		// }

		if err := c.File("error.html"); err != nil {
			c.Logger().Error(err)
		}

		c.Logger().Error(err)
		fmt.Println(err, host)
	}

	if os.Getenv("RESET_CACHE") == "true" {
		_, mobs := getUrls()
		refreshCache(mobs, db, true) // force first time build cache
	}

	finish := make(chan bool)
	e := echo.New()
	if !isdev {
		go func() {
			ee := echo.New()
			ee.Pre(middleware.RemoveTrailingSlash())
			ee.Pre(middleware.HTTPSWWWRedirect())
			ee.Pre(middleware.HTTPSRedirect())
			ee.HTTPErrorHandler = CustomHTTPErrorHandler
			e.Logger.Fatal(ee.Start(":" + os.Getenv("PORT")))
		}()
	}
	go func() {
		customDomains, mobs := getUrls()
		enqueueCache(mobs, db, false)

		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 2}))
		e.Use(middleware.BodyLimit("1M"))
		e.HTTPErrorHandler = CustomHTTPErrorHandler

		e.GET("/", func(c echo.Context) error {
			req := c.Request()
			host := req.Host
			if isdev {
				host, _, _ = net.SplitHostPort(host)
			}

			err := db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("cached_urls"))
				v := b.Get([]byte(host))
				var mob Mobilization
				err := json.Unmarshal(v, &mob)
				if err != nil {
					return err
				}

				noCache := c.QueryParam("nocache")
				if noCache == "1" {
					readCacheContent(mob, db, true)
				}

				c.HTML(http.StatusOK, string(mob.Content)+"<!--"+mob.CachedAt.UTC().Format(time.RFC3339)+"-->")
				return nil
			})
			if err != nil {
				fmt.Errorf("%s", err)
			}
			return nil
		})
		e.Pre(middleware.RemoveTrailingSlash())
		e.Pre(middleware.WWWRedirect())
		if isdev {
			e.Debug = true
			e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
		} else {
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(customDomains...)
			e.AutoTLSManager.Cache = autocert.DirCache("./cache/")
			e.AutoTLSManager.Email = "tech@nossas.org"
			e.DisableHTTP2 = true
			e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
				XFrameOptions:         "",
				HSTSMaxAge:            3600,
				ContentSecurityPolicy: "",
			}))
			e.Logger.Fatal(e.StartAutoTLS(":" + os.Getenv("PORT_SSL")))
		}
	}()
	<-finish

	defer db.Close()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	// quit := make(chan os.Signal)
	// signal.Notify(quit, os.Interrupt)
	// <-quit
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	// if err := e.Shutdown(ctx); err != nil {
	// 	e.Logger.Fatal(err)
	// }
}
