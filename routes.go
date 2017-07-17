package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/thoas/stats"
)

func routeStats(c echo.Context) error {
	statsMiddleware := stats.New()
	stats := statsMiddleware.Data()

	return c.JSON(http.StatusOK, stats)
}

// func routeRoot
// func routeResetAll

// CustomHTTPErrorHandler Echo HTTP Error Handler
func CustomHTTPErrorHandler(err error, c echo.Context) {
	req := c.Request()
	host := req.Host

	// code := http.StatusInternalServerError
	// if he, ok := err.(*echo.HTTPError); ok {
	// 	code = he.Code
	// }

	if err := c.File("error.html"); err != nil {
		c.Logger().Error(err)
	}

	c.Logger().Error(err, host)
}
