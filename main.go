package main

import (
	"net/http"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"encoding/json"
	"time"
	"os"
  "io/ioutil"
	"fmt"
)

type Mobilization struct {
	ID int
	Name string
	Slug string
	Custom_Domain string
}

type HttpResponse struct {
	url      string
	response *http.Response
	err      error
}

func asyncHttpGets(urls []string) []*HttpResponse {
	ch := make(chan *HttpResponse, len(urls)) // buffered
	responses := []*HttpResponse{}
	for _, url := range urls {
		go func(url string) {
			fmt.Printf("Fetching %s \n", url)
			resp, err := http.Get(url)
			if err == nil {
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					fmt.Println("get:\n", string(body))
				}
				defer resp.Body.Close()
			}
			ch <- &HttpResponse{url, resp, err}
		}(url)
	}

	for {
		select {
		case r := <-ch:
			fmt.Printf("%s was fetched\n", r.url)
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

func main() {
	port := os.Getenv("PORT")
	e := echo.New()
	e.Debug = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if port == "" {
		e.Logger.Fatal("$PORT must be set")
	}

	var myClient = &http.Client{Timeout: 10 * time.Second}
  r, err := myClient.Get("https://api-ssl.reboo.org/mobilizations")
  if err != nil {
      panic(err)
  }
  defer r.Body.Close()

	jsonDataFromHttp, err := ioutil.ReadAll(r.Body)
	if err != nil {
					panic(err)
	}

	var jsonData []Mobilization
	err = json.Unmarshal([]byte(jsonDataFromHttp), &jsonData) // here!
	if err != nil {
					panic(err)
	}

	urls := make([]string, 0)
	for _, jd := range jsonData {
		if jd.Custom_Domain != "" {
			urls = append(urls, "http://" + jd.Custom_Domain)
		}
	}

	results := asyncHttpGets(urls)
	for _, result := range results {
		if (result.response != nil) {
			fmt.Printf("%s status: %s\n", result.url,
				result.response.Status)
		}
	}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, 	jsonData[0].Name)
	})

	e.Logger.Fatal(e.Start(":" + port))
}
