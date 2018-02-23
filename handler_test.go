package main

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestWeb_routeRoot(t *testing.T) {
	type fields struct {
		s         Specification
		r         *Redis
		g         *Graphql
		serverSSL *echo.Echo
		server    *echo.Echo
		cache     *CacheManager
		domains   []string
	}
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
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
				cache:     tt.fields.cache,
				domains:   tt.fields.domains,
			}
			web.Setup()
			// e := echo.New()
			req := httptest.NewRequest(echo.GET, "/", nil)
			rec := httptest.NewRecorder()
			c := web.serverSSL.NewContext(req, rec)
			c.SetPath("/")
			// c.SetParamNames("email")
			// c.SetParamValues("jon@labstack.com")
			// h := &handler{mockDB}

			// Assertions
			// if assert.NoError(t, h.getUser(c)) {
			// 	assert.Equal(t, http.StatusOK, rec.Code)
			// 	assert.Equal(t, userJSON, rec.Body.String())
			// }
			if err := web.routeRoot(c); (err != nil) != tt.wantErr {
				t.Errorf("Web.routeRoot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWeb_routeReset(t *testing.T) {
	type fields struct {
		s         Specification
		r         *Redis
		g         *Graphql
		serverSSL *echo.Echo
		server    *echo.Echo
		cache     *CacheManager
		domains   []string
	}
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
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
				cache:     tt.fields.cache,
				domains:   tt.fields.domains,
			}
			web.Setup()
			// e := echo.New()
			req := httptest.NewRequest(echo.GET, "/reset-all", nil)
			rec := httptest.NewRecorder()
			c := web.serverSSL.NewContext(req, rec)
			c.SetPath("/reset-all")
			if err := web.routeReset(c); (err != nil) != tt.wantErr {
				t.Errorf("Web.routeReset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
