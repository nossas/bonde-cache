package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"
)

// Handle Redirect to HTTPS
func ServerRedirect(s Specification) {
	ee := echo.New()
	ee.Pre(middleware.RemoveTrailingSlash())
	ee.Pre(middleware.HTTPSWWWRedirect())
	ee.Pre(middleware.HTTPSRedirect())
	ee.HTTPErrorHandler = CustomHTTPErrorHandler
	ee.Server.Addr = ":" + s.Port
	ee.Logger.Fatal(gracehttp.Serve(ee.Server))
	// gracehttp.SetLogger(ee.Logger)
	// if err := ee.Start(":" + os.Getenv("PORT")); err != nil {
	// 	ee.Logger.Info("Server Redirect: DOWN")
	// } else {
	// 	ee.Logger.Info("Server Redirect: UP")
	// }
}

// Handle HTTPS Certificates
func ServerCache(db *bolt.DB, spec Specification) {
	customDomains, _ := GetUrls()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 2}))
	e.Use(middleware.BodyLimit("1M"))
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.WWWRedirect())
	e.HTTPErrorHandler = CustomHTTPErrorHandler

	e.GET("/reset-all", func(c echo.Context) error {
		_, mobs := GetUrls()
		refreshCache(mobs, db, spec) // force first time build cache

		return c.String(http.StatusOK, "Resetting cache")
	})

	e.GET("/", func(c echo.Context) error {
		req := c.Request()
		host := req.Host
		host, _, _ = net.SplitHostPort(host)

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
				readCacheContent(mob, db, spec)
			}

			c.HTML(http.StatusOK, string(mob.Content)+"<!--"+mob.CachedAt.Format(time.RFC3339)+"-->")
			return nil
		})
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		return nil
	})

	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(customDomains...)
	e.AutoTLSManager.Cache = autocert.DirCache("./cache/")
	e.AutoTLSManager.Email = "tech@nossas.org"
	e.AutoTLSManager.ForceRSA = true
	e.DisableHTTP2 = true
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XFrameOptions:         "",
		HSTSMaxAge:            63072000,
		ContentSecurityPolicy: "",
	}))
	s := e.TLSServer
	cfg := &tls.Config{
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		PreferServerCipherSuites: true,
	}
	s.TLSConfig = cfg
	s.TLSConfig.GetCertificate = e.AutoTLSManager.GetCertificate
	s.Addr = ":" + spec.PortSsl
	e.Logger.Fatal(gracehttp.Serve(e.Server))

	// if err := e.StartServer(e.TLSServer); err != nil {
	// 	e.Logger.Info("Server Cache: DOWN")
	// } else {
	// 	e.Logger.Info("Server Cache: UP")
	// }
}

// Echo HTTP Error Handler
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
