package dkgapi

import (
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

//----------------------------------------------------------
// The POST handler implements 'adm/newdks' API
// Parameters (see NewDKSRequest struct):
//     tmpId:       int value, tmp id for the key set. Should be unique during DKG session
//     n:           size of the assembly
//     t:           required quorum: normally t=floor( 2*n/3)+1
//	   index:       index of the node in the quorum
//
// This API must be called to n nodes with respective indices from 0 to n-1.
// Each node does the following:
// - generates random private key a0
// - generates random keys a1, ..., at
// - creates polynomial Pj(x) out with coefficients a0, a1, ... at of degree t,
//   where j is index of the node
//
// Response (see NewDKSResponse):
// - echoed values like id, assembly_id, index (probably not needed)
// - create - timestamp in unix milliseconds when created (probably not needed)
// - private shares PriShares: values of Pj(1), Pj(2).... Pj(n) EXCEPT diagonal value Pj(j)
//   where j == 'index'+1 of the current node
//
// After this call:
// - n nodes keeps in memory generated random polynomials (not saved yet)
// - caller nxn matrix of Private Shares (except the diagonal values).
//   j-th row of the matrix corresponds to private shares sent by the node j to the dealer (caller)
//
// In the next call 'adm/aggregatedks' dealer will be sending COLUMNS of the matrix to the same nodes.
// Note, that diagonal values never appear in public, so dealer is not able to reconstruct secret polynomials
//
// Next: see '/adm/aggregatedks' API

func HandlerNewDks(c echo.Context) error {
	var req NewDKSRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &NewDKSResponse{
			Err: err.Error(),
		})
	}
	return misc.OkJson(c, NewDKSetReq(&req))
}

type NewDKSRequest struct {
	TmpId int    `json:"tmpId"`
	N     uint16 `json:"n"`
	T     uint16 `json:"t"`
	Index uint16 `json:"index"` // 0 to N-1
}

type NewDKSResponse struct {
	PriShares []string `json:"pri_shares"` // base58
	Err       string   `json:"err"`
}

func NewDKSetReq(req *NewDKSRequest) *NewDKSResponse {
	if err := tcrypto.ValidateDKSParams(req.T, req.N, req.Index); err != nil {
		return &NewDKSResponse{Err: err.Error()}
	}
	ks, err := tcrypto.NewRndDKShare(req.T, req.N, req.Index)
	if err != nil {
		return &NewDKSResponse{Err: err.Error()}
	}
	err = putToDkgCache(req.TmpId, ks)
	if err != nil {
		return &NewDKSResponse{Err: err.Error()}
	}
	resp := NewDKSResponse{
		PriShares: make([]string, ks.N),
	}
	for i, s := range ks.PriShares {
		if uint16(i) != ks.Index {
			data, err := s.V.MarshalBinary()
			if err != nil {
				return &NewDKSResponse{Err: err.Error()}
			}
			resp.PriShares[i] = base58.Encode(data)
		}
	}
	return &resp
}
