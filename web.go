package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
)

// WebRedirect Handle Redirect to HTTPS
func WebRedirect(s Specification) {
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

// WebCache Handle HTTPS Certificates - Show Mob or Error Page
func WebCache(spec Specification) {
	customDomains, _ := GetUrls(spec)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 2}))
	e.Use(middleware.BodyLimit("1M"))
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.WWWRedirect())
	e.HTTPErrorHandler = CustomHTTPErrorHandler

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/reset-all", func(c echo.Context) error {
		// evacuateCache(spec)
		return c.String(http.StatusOK, "Resetting cache")
	})

	e.GET("/", func(c echo.Context) error {
		req := c.Request()
		host := req.Host
		mob := RedisReadMobilization("cached_urls:" + host)
		noCache := c.QueryParam("nocache")
		tCachedAt, _ := time.Parse("2006-01-02T15:04:05.000-07:00", mob.CachedAt)

		if noCache == "1" {
			readHTMLFromHTTPAndSaveToRedis(mob, spec)
			log.Println("Limpando cache..." + mob.Name)
		}

		if mob.Public {
			return c.HTML(http.StatusOK, string(mob.Content)+"<!--"+tCachedAt.Format(time.RFC3339)+"-->")
		}
		return c.HTML(http.StatusOK, string("Página não encontrada!"))
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
	// 1.
	s.TLSConfig = cfg
	if spec.Env == "production" || spec.Env == "staging" {
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
