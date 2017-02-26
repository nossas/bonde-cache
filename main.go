package main

import (
	"net/http"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"encoding/json"
	"time"
	"os"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJson(url string, target interface{}) error {
    r, err := myClient.Get(url)
    if err != nil {
        return err
    }
    defer r.Body.Close()

    return json.NewDecoder(r.Body).Decode(target)
}

type Mobilization struct {
    ID int
		Name string
		Slug string
		CustomDomain string
}

func main() {
	port := os.Getenv("PORT")
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if port == "" {
		e.Logger.Fatal("$PORT must be set")
	}

	e.GET("/", func(c echo.Context) error {
		mob := new(Mobilization) // or &Foo{}
		getJson("https://api-ssl.reboo.org/mobilizations", mob)
		return c.String(http.StatusOK, mob.Name)
	})

	e.Logger.Fatal(e.Start(":" + port))
}
