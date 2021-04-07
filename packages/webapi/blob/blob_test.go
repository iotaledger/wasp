package blob

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestPutBlob(t *testing.T) {
	blobCache := coretypes.NewInMemoryBlobCache()
	b := &blobWebAPI{blobCache}

	data := []byte{1, 3, 3, 7}
	hash := hashing.HashData(data)

	var res model.BlobInfo
	do(
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
	b := &blobWebAPI{blobCache}

	data := []byte{1, 3, 3, 7}

	hash, err := blobCache.PutBlob(data)
	require.NoError(t, err)

	var res model.BlobData
	do(
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
	b := &blobWebAPI{blobCache}

	data := []byte{1, 3, 3, 7}

	hash, err := blobCache.PutBlob(data)
	require.NoError(t, err)

	var res model.BlobInfo
	do(
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

func do(
	t *testing.T,
	handler echo.HandlerFunc,
	method string,
	route string,
	params map[string]string,
	body interface{},
	res interface{},
) {
	e := echo.New()

	var req *http.Request
	if body != nil {
		dataJSON, err := json.Marshal(body)
		require.NoError(t, err)
		req = httptest.NewRequest(method, "/", bytes.NewReader(dataJSON))
	} else {
		req = httptest.NewRequest(method, "/", nil)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath(route)
	for k, v := range params {
		c.SetParamNames(k)
		c.SetParamValues(v)
	}

	err := handler(c)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rec.Code)

	if res != nil {
		err = json.Unmarshal(rec.Body.Bytes(), res)
		require.NoError(t, err)
	}
}
