package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
)

// Web struct to start server
type Web struct {
	s Specification
	r *Redis
	g *Graphql
}

// StartNonSSL Handle Redirect to HTTPS
func (web *Web) StartNonSSL() {

	ee := echo.New()
	ee.Pre(middleware.RemoveTrailingSlash())
	ee.Pre(middleware.HTTPSWWWRedirect())
	ee.Pre(middleware.HTTPSRedirect())
	ee.HTTPErrorHandler = web.CustomHTTPErrorHandler

	Log := log.New(os.Stdout, "[server]: ", log.Ldate|log.Ltime|log.Lshortfile)
	gracehttp.SetLogger(Log)

	ee.Server.Addr = ":" + web.s.Port
	ee.Logger.Fatal(gracehttp.Serve(ee.Server))
}

// StartSSL Handle HTTPS Certificates - Show Mob or Error Page
func (web *Web) StartSSL() {
	var cache = &CacheManager{
		g: web.g,
		s: web.s,
		r: web.r,
	}

	customDomains, _ := cache.GetAllowedDomains()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 2}))
	e.Use(middleware.BodyLimit("1M"))
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.WWWRedirect())
	e.HTTPErrorHandler = web.CustomHTTPErrorHandler

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/reset-all", func(c echo.Context) error {
		// evacuateCache(spec)
		return c.String(http.StatusOK, "Resetting cache")
	})

	e.GET("/", web.routeRoot)

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
		PreferServerCipherSuites: true,
	}

	s.TLSConfig = cfg
	if web.s.Env == "production" || web.s.Env == "staging" {
		s.TLSConfig.GetCertificate = e.AutoTLSManager.GetCertificate
	} else {
		e.Debug = true
		s.TLSConfig.Certificates = make([]tls.Certificate, 1)
		s.TLSConfig.Certificates[0], _ = tls.LoadX509KeyPair("./data/certificates/server.crt", "./data/certificates/server.key")
	}

	LogSsl := log.New(os.Stdout, "[server_ssl]: ", log.Ldate|log.Ltime|log.Lshortfile)
	gracehttp.SetLogger(LogSsl)

	s.Addr = ":" + web.s.PortSsl
	e.Logger.Fatal(gracehttp.Serve(e.TLSServer))
}
