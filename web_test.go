package main

import (
	"testing"

	"github.com/labstack/echo"
)

// 1. Count domains to grab html and create progress to cache and certificate
// 2. Page showing status before cache ready ( loading )
// 3.

// check custom http error handler
// check log output when endpoint accessed
// check open port specified into s Specification

// check custom http error handler
// check log output when endpoint accessed
// check open port specified into s Specification
// get /metrics
// get /
// check certificates configuration

// func TestWeb(t *testing.T) {
// 	// Setup
// 	var s = Specification{}
// 	g := &Graphql{s: s}
// 	r := &Redis{s: s}

// 	web := &Web{s: s, g: g, r: r}
// 	web.server = echo.New()
// 	req := httptest.NewRequest(echo.GET, "/", nil)
// 	rec := httptest.NewRecorder()
// 	c := web.server.NewContext(req, rec)
// 	c.SetPath("/")
// 	// c.SetParamNames("email")
// 	// c.SetParamValues("jon@labstack.com")
// 	// h := &handler{mockDB}

// 	// Assertions
// 	// if assert.NoError(t, h.getUser(c)) {
// 	assert.Equal(t, http.StatusOK, rec.Code)
// 	// assert.Equal(t, userJSON, rec.Body.String())
// 	// }
// }

func TestWeb_Setup(t *testing.T) {
	type fields struct {
		s         Specification
		r         *Redis
		g         *Graphql
		serverSSL *echo.Echo
		server    *echo.Echo
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			web := &Web{
				s:         tt.fields.s,
				r:         tt.fields.r,
				g:         tt.fields.g,
				serverSSL: tt.fields.serverSSL,
				server:    tt.fields.server,
			}
			web.Setup()
		})
	}
}

// func TestWeb_StartNonSSL(t *testing.T) {
// 	type fields struct {
// 		s         Specification
// 		r         *Redis
// 		g         *Graphql
// 		serverSSL *echo.Echo
// 		server    *echo.Echo
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 	}{
// 		{
// 			name: "empty",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// tt.fields.s.Env = "testing"
// 			// tt.fields.s.Sync = false
// 			// tt.fields.s.Interval = 30
// 			// tt.fields.s.Port = "4555"
// 			// tt.fields.s.Port = "4556"
// 			// tt.fields.s.Domain = ""
// 			// tt.fields.s.RedisURL = ""
// 			// tt.fields.s.APIServiceURL = "http://local:3002"
// 			// tt.fields.s.APIServiceToken = ""

// 			web := &Web{
// 				s:         tt.fields.s,
// 				r:         tt.fields.r,
// 				g:         tt.fields.g,
// 				serverSSL: tt.fields.serverSSL,
// 				server:    tt.fields.server,
// 			}
// 			web.Setup()
// 		})
// 	}
// }

// func TestWeb_StartSSL(t *testing.T) {
// 	type fields struct {
// 		s         Specification
// 		r         *Redis
// 		g         *Graphql
// 		serverSSL *echo.Echo
// 		server    *echo.Echo
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 	}{
// 		{
// 			name: "empty",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			web := &Web{
// 				s:         tt.fields.s,
// 				r:         tt.fields.r,
// 				g:         tt.fields.g,
// 				serverSSL: tt.fields.serverSSL,
// 				server:    tt.fields.server,
// 			}
// 			web.Setup()
// 			web.StartSSL()
// 		})
// 	}
// }
