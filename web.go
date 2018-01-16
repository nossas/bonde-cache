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
	"golang.org/x/crypto/acme/autocert"
)

// ServerRedirect Handle Redirect to HTTPS
func webRedirect(s Specification) {
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

// ServerCache Handle HTTPS Certificates
func webCache(spec Specification) {
	customDomains, _ := GetUrls(spec)

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
		populateCache(spec, false)
		return c.String(http.StatusOK, "Resetting cache")
	})

	e.GET("/", func(c echo.Context) error {
		req := c.Request()
		host := req.Host

		var mob = redisRead("cached_urls:" + host)

		// v, err := redis.Values(conn.Do("HGETALL", "cached_urls:"+host))
		// if err != nil {
		// 	fmt.Println(err)
		// 	return nil
		// }
		// var mob RedisMobilization
		// if err := redis.ScanStruct(v, &mob); err != nil {
		// 	fmt.Println(err)
		// 	return nil
		// }
		// var mob2 Mobilization
		// // var f interface{}
		// err2 := json.Unmarshal(v, &mob2)
		// if err2 != nil {
		// 	return err2
		// }

		noCache := c.QueryParam("nocache")
		if noCache == "1" {
			readOriginContent(mob, spec)
			log.Println("Limpando cache..." + mob.Name)
		}

		if mob.Public {
			return c.HTML(http.StatusOK, string(mob.Content)+"<!--"+mob.CachedAt.Format(time.RFC3339)+"-->")
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
