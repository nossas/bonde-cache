package main

import (
    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
    "golang.org/x/crypto/acme/autocert"
    "github.com/boltdb/bolt"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "time"
    "os"
    "fmt"
    "net"
    "strconv"
)

type Mobilization struct {
    ID int
    Name string
    Slug string
    Custom_Domain string
}

type HttpResponse struct {
    url string
    response *http.Response
    err error
}

func getUrls() []string {
    var myClient = & http.Client { Timeout: 10 * time.Second }
    r, err := myClient.Get("https://api-ssl.reboo.org/mobilizations")
    if err != nil {
        panic(err)
    }
    defer r.Body.Close()

    jsonDataFromHttp, err := ioutil.ReadAll(r.Body)
    if err != nil {
        panic(err)
    }

    var jsonData[] Mobilization
    err = json.Unmarshal([]byte(jsonDataFromHttp), & jsonData) // here!
    if err != nil {
        panic(err)
    }

    urls := make([]string, 0)
    for _, jd := range jsonData {
        if jd.Custom_Domain != "" {
            urls = append(urls, "http://" + jd.Custom_Domain)
        }
    }
    return urls
}

func refreshCache(urls[]string, db*bolt.DB, interval string)[]*HttpResponse {
    i, _ := strconv.Atoi(interval)
    ticker := time.NewTicker(time.Duration(i) * time.Second)
    quit := make(chan struct{})
    go func() {
        for {
           select {
            case <- ticker.C:
                results := readCacheContent(urls, db)
                for _, result := range results {
                    if (result.response != nil) {
                        fmt.Printf("%s status: %s\n", result.url, result.response.Status)
                    }
                }
            case <- quit:
                ticker.Stop()
                return
            }
        }
    }()
    return nil
}

func readCacheContent(urls[]string, db*bolt.DB)[]*HttpResponse {
    ch := make(chan *HttpResponse, len(urls)) // buffered
    responses := []* HttpResponse {}

    for _, url := range urls {
        go func(url string) {
            fmt.Printf("fetch url %s \n", url)
            resp, err := http.Get(url)
            if err == nil {
                saveCacheContent(url, resp, db)
            } else {
                fmt.Errorf("error read response http: %s", err)
            }
            ch <- & HttpResponse { url, resp, err }
        }(url)
    }

    for {
        select {
            case r := <-ch:
                fmt.Printf("fetched url %s\n", r.url)
                responses = append(responses, r)
                if len(responses) == len(urls) {
                    return responses
                }
            case <-time.After(50 * time.Millisecond):
                fmt.Printf(".")
        }
    }

    return responses
}

func saveCacheContent(url string, resp *http.Response, db *bolt.DB) {
    body, err := ioutil.ReadAll(resp.Body)
    if err == nil {
        err := db.Update(func(tx * bolt.Tx) error {
            b, err := tx.CreateBucketIfNotExists([]byte("cached_urls"))
            if err != nil {
                return fmt.Errorf("error create cache bucket: %s", err)
            }
            b.Put([]byte(url), body)
            fmt.Printf("saved body content url %s \n", url)
            return nil
        })
        if err != nil {
            fmt.Errorf("error save content %s", err)
        }
    } else {
        fmt.Errorf("error read response body: %s", err)
    }
    defer resp.Body.Close()
}

func main() {
    db, err := bolt.Open("bonde-cache.db", 0600, nil)
    if err != nil {
        fmt.Errorf("open cache: %s", err)
    }

    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Static("/dist", "public/dist")
    e.Static("/wysihtml", "public/wysihtml")

    urls := getUrls()
    readCacheContent(urls, db) // force first time build cache
    refreshCache(urls, db, os.Getenv("CACHE_INTERVAL"))

    e.GET("/", func(c echo.Context) error {
        req := c.Request()
        host, _, _ := net.SplitHostPort(req.Host)

        err := db.View(func(tx *bolt.Tx) error {
            b := tx.Bucket([]byte("cached_urls"))
            v := b.Get([]byte("http://" + host))
            fmt.Printf("saved body content url %s \n", "http://" + host)
            c.HTML(http.StatusOK, string(v))
            return nil
        })
        if err != nil {
            fmt.Errorf("%s", err)
        }
        return nil
    })

    isdev, err := strconv.ParseBool(os.Getenv("IS_DEV"))
    if isdev {
        e.Debug = true
        e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
    } else {
        e.Pre(middleware.HTTPSWWWRedirect())
        e.AutoTLSManager.Cache = autocert.DirCache("/tmp/.cache")
        e.Logger.Fatal(e.StartAutoTLS(":443"))
    }
    defer db.Close()
}
