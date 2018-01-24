// Worker has two main tasks:
// - populate redis with cache html content from mobilizations published
// - validate certificates with dns challenge
package main

import (
	"net"
	"net/http"
	"time"

	"github.com/jasonlvhit/gocron"
)

// Mobilization is cached at Redis
type Mobilization struct {
	ID              int    `json:"id" redis:"id"`
	Name            string `json:"name" redis:"name"`
	Content         []byte `json:"content" redis:"content"`
	CachedAt        string `json:"cached_at" redis:"cached_at"`
	Slug            string `json:"slug" redis:"slug"`
	CustomDomain    string `json:"custom_domain" redis:"custom_domain"`
	UpdatedAt       string `json:"updated_at" redis:"updated_at"`
	Public          bool   `json:"public" redis:"public"`
	CertificateRoot bool   `json:"certificate_root" redis:"certificate_root"`
}

// Certificate is cached at Redis
type Certificate struct {
	ID   int    `json:"id" redis:"id"`
	Name string `json:"name" redis:"name"`
}

// HTTPResponse help handle output from requests
type HTTPResponse struct {
	url      string
	response *http.Response
	err      error
}

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

var netClient = &http.Client{
	Timeout:   time.Second * 20,
	Transport: netTransport,
}

// Worker are entry point to recurring tasks
func Worker(s Specification) {
	populateCache(s)
	populateCertificates(s)
	gocron.Every(30).Seconds().Do(populateCache, s)
	gocron.Every(30).Seconds().Do(populateCertificates, s)
}
