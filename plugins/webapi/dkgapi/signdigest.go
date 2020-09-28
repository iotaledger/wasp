package dkgapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

func HandlerSignDigest(c echo.Context) error {
	var req SignDigestRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &SignDigestResponse{
			Err: err.Error(),
		})
	}
	return misc.OkJson(c, SignDigestReq(&req))
}

type SignDigestRequest struct {
	Address    string             `json:"address"`
	DataDigest *hashing.HashValue `json:"data_digest"`
}

type SignDigestResponse struct {
	SigShare string `json:"sig_share"`
	Err      string `json:"err"`
}

func SignDigestReq(req *SignDigestRequest) *SignDigestResponse {
	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return &SignDigestResponse{Err: err.Error()}
	}
	if addr.Version() != address.VersionBLS {
		return &SignDigestResponse{Err: "expected BLS address"}
	}
	ks, ok, err := registry.GetDKShare(&addr)
	if err != nil {
		return &SignDigestResponse{Err: err.Error()}
	}
	if !ok {
		return &SignDigestResponse{Err: "unknown key share"}
	}
	if !ks.Committed {
		return &SignDigestResponse{Err: "uncommitted key set"}
	}
	signature, err := ks.SignShare(req.DataDigest.Bytes())
	if err != nil {
		return &SignDigestResponse{Err: err.Error()}
	}
	return &SignDigestResponse{
		SigShare: base58.Encode(signature),
	}
}
