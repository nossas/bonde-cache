package main

import (
	"net/http"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"encoding/json"
	"time"
	"os"
  "io/ioutil"
)

type Mobilization struct {
	ID int `json:id`
	Name string `json:name`
	Slug string `json:slug`
	CustomDomain string `json:custom_domain`
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
      e.Logger.Fatal(err)
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

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, 	jsonData[0].Name)
	})

	e.Logger.Fatal(e.Start(":" + port))
}
