package main

import (
	"net/http"
	"github.com/labstack/echo"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	e := echo.New()

	if port == "" {
		e.Logger.Fatal("$PORT must be set")
	}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.Logger.Fatal(e.Start(":" + port))
}
