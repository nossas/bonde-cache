package main

import (
    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
    "golang.org/x/crypto/acme/autocert"
    "github.com/boltdb/bolt"
    rice "github.com/GeertJohan/go.rice"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "time"
    "os"
    "fmt"
    "net"
    "bytes"
    "strconv"
    "io"
    "golang.org/x/net/html"
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

func getUrls() (url[]string, customDomains[]string) {
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
    customDomains = make([]string, 0)
    for _, jd := range jsonData {
        if jd.Custom_Domain != "" {
            urls = append(urls, "http://" + jd.Slug + ".reboo.org")
            customDomains = append(customDomains, jd.Custom_Domain)
        }
    }
    return urls, customDomains
}

func refreshCache(urls[]string, customDomains[]string, db*bolt.DB, interval string)[]*HttpResponse {
    i, _ := strconv.Atoi(interval)
    ticker := time.NewTicker(time.Duration(i) * time.Second)
    quit := make(chan struct{})
    go func() {
        for {
           select {
            case <- ticker.C:
                results := readCacheContent(urls, customDomains, db)
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

func readCacheContent(urls[]string, customDomains[]string, db*bolt.DB)[]*HttpResponse {
    ch := make(chan *HttpResponse, len(urls)) // buffered
    responses := []* HttpResponse {}

    for i, url := range urls {
        go func(url string,i int) {
            fmt.Printf("fetch url %s \n", customDomains[i])
            var myClient = & http.Client { Timeout: 10 * time.Second }
            resp, err := myClient.Get(url)
            // defer resp.Body.Close()
            if err == nil {
                saveCacheContent(url, customDomains[i], resp, db)
            } else {
                fmt.Errorf("error read response http: %s", err)
            }
            ch <- & HttpResponse { customDomains[i], resp, err }
        }(url, i)
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

func saveCacheContent(url string, customDomain string, resp *http.Response, db *bolt.DB) {
    body, err := ioutil.ReadAll(resp.Body)
    readContentAssets(body, url)
    if err == nil {
        err := db.Update(func(tx * bolt.Tx) error {
            b, err := tx.CreateBucketIfNotExists([]byte("cached_urls"))
            if err != nil {
                return fmt.Errorf("error create cache bucket: %s", err)
            }
            b.Put([]byte(customDomain), body)
            fmt.Printf("saved body content url %s \n", url)
            return nil
        })
        if err != nil {
            fmt.Errorf("error save content %s", err)
        }
    } else {
        fmt.Errorf("error read response body: %s", err)
    }
}

// Helper function to pull the href attribute from a Token
func getJS(t html.Token, url string) (ok bool) {
    // ok = false
    // Iterate over all of the Token's attributes until we find an "href"
    for _, a := range t.Attr {
        if a.Key == "src" {
            downloadFile("./public" + a.Val, url + a.Val)
            // href = a.Val
            ok = true
        }
    }

    // "bare" return will return the variables (ok, href) as defined in
    // the function definition
    return
}

// Helper function to pull the href attribute from a Token
func getCSS(t html.Token, url string) (ok bool) {
    // ok = false
    // Iterate over all of the Token's attributes until we find an "href"
    for _, a := range t.Attr {
        if a.Key == "href" {
            downloadFile("./public" + a.Val, url + a.Val)
            // href = a.Val
            ok = true
        }
    }

    // "bare" return will return the variables (ok, href) as defined in
    // the function definition
    return
}


func downloadFile(filepath string, url string) (err error) {
    // Create the file
    out, err := os.Create(filepath)
    if err != nil  {
        return err
    }
    defer out.Close()

    var myClient = & http.Client { Timeout: 10 * time.Second }
    // Get the data
    resp, err := myClient.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Writer the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil  {
        return err
    }

    return nil
}

// https://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html
func readContentAssets(body []byte, url string) {
    r := bytes.NewReader(body)
    z := html.NewTokenizer(r)
    for {
        tt := z.Next()
        t := z.Token()
        switch {
        case tt == html.ErrorToken:
            // End of the document, we're done
            return
        case tt == html.StartTagToken:
            if (t.Data == "script") {
                getJS(t, url)
            }
        case tt == html.SelfClosingTagToken:
            if (t.Data == "link") {
                getCSS(t, url)
            }
        }
    }
}

func main() {
    isdev, err := strconv.ParseBool(os.Getenv("IS_DEV"))

    db, err := bolt.Open("bonde-cache.db", 0600, nil)
    if err != nil {
        fmt.Errorf("open cache: %s", err)
    }

    urls, customDomains := getUrls()
    readCacheContent(urls, customDomains, db) // force first time build cache
    refreshCache(urls, customDomains, db, os.Getenv("CACHE_INTERVAL"))

    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.GzipWithConfig(middleware.GzipConfig{ Level: 2 }))
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
        e.Debug = true
        ee := echo.New()
        ee.Pre(middleware.RemoveTrailingSlash())
        ee.Pre(middleware.WWWRedirect())
        ee.Pre(middleware.HTTPSRedirect())

        e.Pre(middleware.HTTPSWWWRedirect())
        e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(customDomains...)
        e.AutoTLSManager.Cache = autocert.DirCache("./cache/")

        finish := make(chan bool)
        go func() {
            e.Logger.Fatal(ee.Start(":80"))
        }()
        go func() {
            e.Logger.Fatal(e.StartAutoTLS(":443"))
        }()
        <-finish

    }
    defer db.Close()
}
