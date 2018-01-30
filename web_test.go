package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
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

func TestWeb(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/")
	// c.SetParamNames("email")
	// c.SetParamValues("jon@labstack.com")
	// h := &handler{mockDB}

	// Assertions
	// if assert.NoError(t, h.getUser(c)) {
	assert.Equal(t, http.StatusOK, rec.Code)
	// assert.Equal(t, userJSON, rec.Body.String())
	// }
}
