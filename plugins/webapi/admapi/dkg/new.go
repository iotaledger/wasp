package dkg

import (
	"net/http"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
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
// Next: see 'aggregate' API call

func addDksNewEndpoint(adm *echo.Group) {
	adm.POST("/"+client.DKSNewRoute, handleNewDks)
}

func handleNewDks(c echo.Context) error {
	log.Debugw("HandlerNewDks")
	var req client.NewDKSRequest

	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	if err := tcrypto.ValidateDKSParams(req.T, req.N, req.Index); err != nil {
		return httperrors.BadRequest(err.Error())
	}
	ks, err := tcrypto.NewRndDKShare(req.T, req.N, req.Index)
	if err != nil {
		return err
	}
	err = putToDkgCache(req.TmpId, ks)
	if err != nil {
		return err
	}
	priShares := make([]string, ks.N)
	for i, s := range ks.PriShares {
		if uint16(i) != ks.Index {
			data, err := s.V.MarshalBinary()
			if err != nil {
				return err
			}
			priShares[i] = base58.Encode(data)
		}
	}
	return c.JSON(http.StatusOK, priShares)
}
