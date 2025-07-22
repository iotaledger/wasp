package node_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/webapi"
	"github.com/iotaledger/wasp/v2/packages/webapi/controllers/node"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

func TestNodeVersion(t *testing.T) {
	version := "testVersion"
	c := node.NewNodeController(version, nil, nil, nil, nil)
	e := echo.New()
	server := echoswagger.New(e, "/doc", &echoswagger.Info{
		Title:       "Test Wasp API",
		Description: "Test REST API for the Wasp node",
		Version:     version,
	})
	group := server.Group(c.Name(), fmt.Sprintf("/v%d/", 0))
	mocker := webapi.NewMocker()
	c.RegisterPublic(group, mocker)

	req := httptest.NewRequest(http.MethodGet, "/v0/node/version", http.NoBody)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var res models.VersionResponse
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)
	require.Equal(t, models.VersionResponse{Version: version}, res)
}
