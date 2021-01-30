package client

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// PutBlob uploads a blob to the registry
func (c *WaspClient) PutBlob(data []byte) (hashing.HashValue, error) {
	req := model.NewBlobData(data)
	res := &model.BlobInfo{}
	err := c.do(http.MethodGet, routes.PutBlob(), req, res)
	return res.Hash.HashValue(), err
}

// GetBlob fetches a blob by its hash
func (c *WaspClient) GetBlob(hash hashing.HashValue) ([]byte, error) {
	res := &model.BlobData{}
	err := c.do(http.MethodGet, routes.GetBlob(hash.String()), nil, res)
	if err != nil {
		return nil, err
	}
	return res.Data.Bytes(), nil
}

// HasBlob returns whether or not a blob exists
func (c *WaspClient) HasBlob(hash hashing.HashValue) (bool, error) {
	res := &model.BlobInfo{}
	err := c.do(http.MethodGet, routes.HasBlob(hash.String()), nil, res)
	return res.Exists, err
}
