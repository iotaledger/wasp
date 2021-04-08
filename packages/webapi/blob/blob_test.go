package blob

import (
	"net/http"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

func TestPutBlob(t *testing.T) {
	blobCache := coretypes.NewInMemoryBlobCache()
	b := &blobWebAPI{func() coretypes.BlobCache { return blobCache }}

	data := []byte{1, 3, 3, 7}
	hash := hashing.HashData(data)

	var res model.BlobInfo
	testutil.CallHTTPRequestHandler(
		t,
		b.handlePutBlob,
		http.MethodPost,
		routes.PutBlob(),
		nil,
		model.NewBlobData(data),
		&res,
	)
	require.EqualValues(t, hash, res.Hash.HashValue())

	d, _, _ := blobCache.GetBlob(hash)
	require.EqualValues(t, data, d)
}

func TestGetBlob(t *testing.T) {
	blobCache := coretypes.NewInMemoryBlobCache()
	b := &blobWebAPI{func() coretypes.BlobCache { return blobCache }}

	data := []byte{1, 3, 3, 7}

	hash, err := blobCache.PutBlob(data)
	require.NoError(t, err)

	var res model.BlobData
	testutil.CallHTTPRequestHandler(
		t,
		b.handleGetBlob,
		http.MethodGet,
		routes.GetBlob(":hash"),
		map[string]string{"hash": hash.Base58()},
		nil,
		&res,
	)
	require.EqualValues(t, data, res.Data.Bytes())
}

func TestHasBlob(t *testing.T) {
	blobCache := coretypes.NewInMemoryBlobCache()
	b := &blobWebAPI{func() coretypes.BlobCache { return blobCache }}

	data := []byte{1, 3, 3, 7}

	hash, err := blobCache.PutBlob(data)
	require.NoError(t, err)

	var res model.BlobInfo
	testutil.CallHTTPRequestHandler(
		t,
		b.handleHasBlob,
		http.MethodGet,
		routes.HasBlob(":hash"),
		map[string]string{"hash": hash.Base58()},
		nil,
		&res,
	)
	require.EqualValues(t, hashing.HashData(data), res.Hash.HashValue())
}
