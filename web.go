package main

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
)

// Web struct to start server
type Web struct {
	s         Specification
	r         *Redis
	g         *Graphql
	serverSSL *echo.Echo
	server    *echo.Echo
	cache     *CacheManager
	domains   []string
}

// Setup Web to config non ssl and ssl servers
func (web *Web) Setup() {
	web.g = &Graphql{s: web.s}
	web.r = &Redis{s: web.s}

	web.g.CreateClient()
	web.r.CreatePool()

	web.cache = &CacheManager{s: web.s, g: web.g, r: web.r}
	web.domains, _ = web.cache.GetAllowedDomains()

	web.server = echo.New()
	web.serverSSL = echo.New()

	web.SetupNonSSL()
	web.SetupSSL()
}

// SetupNonSSL Handle Redirect to HTTPS
func (web *Web) SetupNonSSL() {

	ee := web.server
	ee.Pre(middleware.RemoveTrailingSlash())
	ee.Pre(middleware.HTTPSWWWRedirect())
	ee.Pre(middleware.HTTPSRedirect())
	ee.HTTPErrorHandler = web.CustomHTTPErrorHandler

	Log := log.New(os.Stdout, "[server]: ", log.Ldate|log.Ltime|log.Lshortfile)
	gracehttp.SetLogger(Log)

	ee.Server.Addr = ":" + web.s.Port
	// ee.Logger.Fatal(gracehttp.Serve(ee.Server))
}

// SetupSSL Handle HTTPS Certificates - Show Mob or Error Page
func (web *Web) SetupSSL() {

	e := web.serverSSL
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 2}))
	e.Use(middleware.BodyLimit("1M"))
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.WWWRedirect())
	e.HTTPErrorHandler = web.CustomHTTPErrorHandler

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/reset-all", web.routeReset)
	e.GET("/", web.routeRoot)

	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(web.domains...)
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
	web.serverSSL = e
	// e.Logger.Fatal(gracehttp.Serve(e.TLSServer))
}
