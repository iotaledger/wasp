package dkg

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/httperrors"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
	"go.dedis.ch/kyber/v3"
)

// The POST handler implements 'adm/commitdks' API
// RequestSectionParams(see CommitDKSRequest struct):
//     tmpId:      tmp num id during DKG process
//     pub_shares: all public chares collected from all nodes
//
// Node does the following:
//    - finalizes all necessary information, which allows to sign data with the private key
//    in the way for it to be verifiable with public keys as per BLS threshold encryption
//    - saves the key into the registry of the nodes in persistent way
//
// After all adm/commitdks calls returns SUCCESS (no error), the dealer is sure, that ditributed
// key set with 'assembly_id' and 'id' was successfully distributed and persistently stored in respective nodes
//
// NOTE: the way the keys are distributed with 'newdks', 'aggregatedks' and 'commitdks' calls
// ensure that nobody, except at least 't' of nodes can create valid BLS signatures.
// Even dealer has not enough information to do that

func addDksCommitEndpoint(adm *echo.Group) {
	adm.POST("/"+client.DKSCommitRoute, handleCommitDks)
}

func handleCommitDks(c echo.Context) error {
	var req client.CommitDKSRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.BadRequest("Invalid request body")
	}

	ks := getFromDkgCache(req.TmpId)
	if ks == nil {
		return httperrors.BadRequest(fmt.Sprintf("wrong tmpId %d", req.TmpId))
	}
	if ks.Committed {
		return httperrors.Conflict("key set is already committed")
	}
	if len(req.PubShares) != int(ks.N) {
		return httperrors.BadRequest("wrong number of private shares")
	}

	pubKeys := make([]kyber.Point, len(req.PubShares))
	for i, s := range req.PubShares {
		b, err := base58.Decode(s)
		if err != nil {
			return httperrors.BadRequest(err.Error())
		}
		p := ks.Suite.G2().Point()
		if err := p.UnmarshalBinary(b); err != nil {
			return httperrors.BadRequest(err.Error())
		}
		pubKeys[i] = p
	}
	err := registry.CommitDKShare(ks, pubKeys)
	if err != nil {
		return err
	}

	// delete from the DKG cache
	_ = putToDkgCache(req.TmpId, nil)

	log.Infow("Created new key share",
		"address", ks.Address.String(),
		"N", ks.N,
		"T", ks.T,
		"Index", ks.Index,
	)

	return c.JSON(http.StatusOK, jsonable.NewAddress(ks.Address))
}
