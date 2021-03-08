package downloader

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

const MOCK_SERVER_PORT string = ":9999"
const FILE_CID string = "someunrealistichash12345"

var FILE []byte = []byte{'s', 'o', 'm', 'e', ' ', 'f', 'i', 'l', 'e', ' ', 'f', 'o', 'r', ' ', 't', 'e', 's', 't', 'i', 'n', 'g'} // should be constant, but...

func startMockServer() *echo.Echo {
	var e *echo.Echo
	e = echo.New()
	e.GET("/ipfs/:cid", func(c echo.Context) error {
		var cid string
		var response []byte
		cid = c.Param("cid")
		switch cid {
		case FILE_CID:
			response = FILE
		default:
			return c.NoContent(http.StatusNotFound)
		}
		return c.Blob(http.StatusOK, "text/plain", response)
	})
	go func() {
		var err error = e.Start(MOCK_SERVER_PORT)
		if err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()
	return e
}

func stopMockServer(e *echo.Echo) {
	var ctx context.Context
	var cancel context.CancelFunc
	var err error

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = e.Shutdown(ctx)
	if err != nil {
		e.Logger.Fatal(err)
	}
}

func TestIpfsDownload(t *testing.T) {
	var log *logger.Logger = testutil.NewLogger(t)

	var downloader *Downloader = New(log, "http://localhost"+MOCK_SERVER_PORT)
	var server *echo.Echo = startMockServer()
	time.Sleep(100 * time.Millisecond) // Time to wait for server start

	var hash hashing.HashValue = hashing.HashData(FILE)

	var db *dbprovider.DBProvider = dbprovider.NewInMemoryDBProvider(log)
	var reg *registry.Impl = registry.NewRegistry(nil, log, db)
	var result bool
	var err error

	result, err = reg.HasBlob(hash)
	require.NoError(t, err)
	require.False(t, result, "The file should not be part of the registry before the download")
	err = downloader.DownloadAndStore(hash, "ipfs://"+FILE_CID, reg)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond) // Time to wait for download completion
	result, err = reg.HasBlob(hash)
	require.True(t, result, "The file must be part of the registry after the download")
	require.NoError(t, err)

	stopMockServer(server)
}
