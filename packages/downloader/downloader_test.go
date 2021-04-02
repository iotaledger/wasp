package downloader

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

const constMockServerPort string = ":9999"
const constFileCID string = "someunrealistichash12345"

var constVarFile []byte = []byte("some file for testing")

func startMockServer() *echo.Echo {
	l, err := net.Listen("tcp", constMockServerPort)
	e := echo.New()
	if err != nil {
		e.Logger.Fatal(err)
	}
	e.Listener = l

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
		err := e.Start("")
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
	log := testlogger.NewLogger(t)

	downloader := New(log, "http://localhost"+constMockServerPort)
	server := startMockServer()
	defer stopMockServer(server)

	hash := hashing.HashData(constVarFile)
	db := dbprovider.NewInMemoryDBProvider(log)
	reg := registry.NewRegistry(nil, log, db)
	result, err := reg.HasBlob(hash)
	chanDownloaded := make(chan bool)
	require.NoError(t, err)
	require.False(t, result, "The file should not be part of the registry before the download")
	err = downloader.DownloadAndStore(hash, "ipfs://"+constFileCID, reg, chanDownloaded)
	require.NoError(t, err)
	select {
	case downloaded := <-chanDownloaded:
		require.True(t, downloaded, "The downloader should successfully download the file")
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("The download job of downloader timeouted")
	}
	result, err = reg.HasBlob(hash)
	require.True(t, result, "The file must be part of the registry after the download")
	require.NoError(t, err)
}
