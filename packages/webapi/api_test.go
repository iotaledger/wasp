package webapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iotaledger/wasp/packages/webapi"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	version := "testVersion"
	e := echo.New()
	server := echoswagger.New(e, "/doc", &echoswagger.Info{
		Title:       "Test Wasp API",
		Description: "Test REST API for the Wasp node",
		Version:     version,
	})
	webapi.AddHealthEndpoint(server)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
