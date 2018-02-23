package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
)

func (web *Web) routeRoot(c echo.Context) error {
	req := c.Request()
	host := req.Host
	mob := web.r.ReadMobilization("cached_urls:" + host)
	// noCache := c.QueryParam("nocache")
	tCachedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", mob.CachedAt)

	// if noCache == "1" {
	// 	readHTMLFromHTTPAndSaveToRedis(mob, spec)
	// 	log.Println("Limpando cache..." + mob.Name)
	// }

	if mob.Public {
		return c.HTML(http.StatusOK, string(mob.Content)+"<!--"+tCachedAt.Format(time.RFC3339)+"-->")
	}
	return c.HTML(http.StatusOK, string("Página não encontrada!"))
}

func (web *Web) routeReset(c echo.Context) error {
	// evacuateCache(spec)
	return c.String(http.StatusOK, "Resetting cache")
}

// CustomHTTPErrorHandler Echo HTTP Error Handler
func (web *Web) CustomHTTPErrorHandler(err error, c echo.Context) {
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
