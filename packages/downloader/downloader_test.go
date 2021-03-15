package downloader

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

const constMockServerPort string = ":9999"
const constFileCID string = "someunrealistichash12345"

var constVarFile []byte = []byte("some file for testing")

func startMockServer() *echo.Echo {
	e := echo.New()
	e.GET("/ipfs/:cid", func(c echo.Context) error {
		var response []byte
		cid := c.Param("cid")
		switch cid {
		case constFileCID:
			response = constVarFile
		default:
			return c.NoContent(http.StatusNotFound)
		}
		return c.Blob(http.StatusOK, "text/plain", response)
	})
	go func() {
		err := e.Start(constMockServerPort)
		if err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()
	return e
}

func stopMockServer(e *echo.Echo) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := e.Shutdown(ctx)
	if err != nil {
		e.Logger.Fatal(err)
	}
}

func TestIpfsDownload(t *testing.T) {
	log := testutil.NewLogger(t)

	downloader := New(log, "http://localhost"+constMockServerPort)
	server := startMockServer()
	defer stopMockServer(server)

	time.Sleep(100 * time.Millisecond) // Time to wait for server start
	hash := hashing.HashData(constVarFile)
	db := dbprovider.NewInMemoryDBProvider(log)
	reg := registry.NewRegistry(nil, log, db)
	result, err := reg.HasBlob(hash)
	require.NoError(t, err)
	require.False(t, result, "The file should not be part of the registry before the download")
	err = downloader.DownloadAndStore(hash, "ipfs://"+constFileCID, reg)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond) // Time to wait for download completion
	result, err = reg.HasBlob(hash)
	require.True(t, result, "The file must be part of the registry after the download")
	require.NoError(t, err)
}
