package dkgapi

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
	"go.dedis.ch/kyber/v3"
)

// The POST handler implements 'adm/commitdks' API
// Parameters(see CommitDKSRequest struct):
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

func HandlerCommitDks(c echo.Context) error {
	var req CommitDKSRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &CommitDKSResponse{
			Err: err.Error(),
		})
	}
	return misc.OkJson(c, CommitDKSReq(&req))
}

type CommitDKSRequest struct {
	TmpId     int      `json:"tmpId"`
	PubShares []string `json:"pub_shares"` // base58
}

type CommitDKSResponse struct {
	Address string `json:"address"` //base58
	Err     string `json:"err"`
}

func CommitDKSReq(req *CommitDKSRequest) *CommitDKSResponse {
	ks := getFromDkgCache(req.TmpId)
	if ks == nil {
		return &CommitDKSResponse{Err: fmt.Sprintf("wrong tmpId %d", req.TmpId)}
	}
	if ks.Committed {
		return &CommitDKSResponse{Err: "key set is already committed"}
	}
	if len(req.PubShares) != int(ks.N) {
		return &CommitDKSResponse{Err: "wrong number of private shares"}
	}

	pubKeys := make([]kyber.Point, len(req.PubShares))
	for i, s := range req.PubShares {
		b, err := base58.Decode(s)
		if err != nil {
			return &CommitDKSResponse{Err: err.Error()}
		}
		p := ks.Suite.G2().Point()
		if err := p.UnmarshalBinary(b); err != nil {
			return &CommitDKSResponse{Err: err.Error()}
		}
		pubKeys[i] = p
	}
	err := registry.CommitDKShare(ks, pubKeys)
	if err != nil {
		return &CommitDKSResponse{Err: err.Error()}
	}
	// delete from the DKG cache
	_ = putToDkgCache(req.TmpId, nil)

	log.Infow("Created new key share",
		"address", ks.Address.String(),
		"N", ks.N,
		"T", ks.T,
		"Index", ks.Index,
	)
	return &CommitDKSResponse{
		Address: ks.Address.String(),
	}
}
