package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/thoas/stats"
	"golang.org/x/crypto/acme/autocert"
)

// ServerRedirect Handle Redirect to HTTPS
func ServerRedirect(s Specification) {
	ee := echo.New()
	ee.Pre(middleware.RemoveTrailingSlash())
	ee.Pre(middleware.HTTPSWWWRedirect())
	ee.Pre(middleware.HTTPSRedirect())
	ee.HTTPErrorHandler = CustomHTTPErrorHandler

	Log := log.New(os.Stdout, "[server]: ", log.Ldate|log.Ltime|log.Lshortfile)
	gracehttp.SetLogger(Log)

	ee.Server.Addr = ":" + s.Port
	ee.Logger.Fatal(gracehttp.Serve(ee.Server))
}

func routeStats(c echo.Context) error {
	statsMiddleware := stats.New()
	stats := statsMiddleware.Data()

	return c.JSON(http.StatusOK, stats)
}

// func routeRoot
// func routeResetAll

// ServerCache Handle HTTPS Certificates
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

	e.GET("/stats", routeStats)

	e.GET("/reset-all", func(c echo.Context) error {
		_, mobs := GetUrls()
		spec.Reset = true
		refreshCache(mobs, db, spec)
		spec.Reset = false
		return c.String(http.StatusOK, "Resetting cache")
	})

	e.GET("/", func(c echo.Context) error {
		req := c.Request()
		host := req.Host

		db.View(func(tx *bolt.Tx) error {
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
		return nil
	})

	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(customDomains...)
	e.AutoTLSManager.Cache = autocert.DirCache("./data/certificates/")
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
		// MinVersion:               tls.VersionTLS12,
		// CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256, tls.X25519},
		PreferServerCipherSuites: true,
		// CipherSuites: []uint16{
		// 	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		// 	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		// 	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		// 	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		// },
	}
	s.TLSConfig = cfg
	if spec.Env == "production" {
		s.TLSConfig.GetCertificate = e.AutoTLSManager.GetCertificate
	} else {
		e.Debug = true
		s.TLSConfig.Certificates = make([]tls.Certificate, 1)
		s.TLSConfig.Certificates[0], _ = tls.LoadX509KeyPair("./data/certificates/server.crt", "./data/certificates/server.key")
	}

	LogSsl := log.New(os.Stdout, "[server_ssl]: ", log.Ldate|log.Ltime|log.Lshortfile)
	gracehttp.SetLogger(LogSsl)

	s.Addr = ":" + spec.PortSsl
	e.Logger.Fatal(gracehttp.Serve(e.TLSServer))

}

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
