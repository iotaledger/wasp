package dkgapi

import (
	"fmt"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
	"go.dedis.ch/kyber/v3"
)

//----------------------------------------------------------
// The POST handler implements 'adm/aggregatedks' API
// Parameters(see AggregateDKSRequest struct):
//     tmpId:       temporary numeric id during DKG
//     index:       index of the node in the assembly
//         (node knows it from the previous adm/newdks call, ths parameter is for control only)
//     pri_shares: values P1(j), P2(j), ...., Pn(j) EXCEPT Pj(j), the diagonal
//        where j is index+1 of the called node.
//
// Node does the following:
//  - it sums up all received pri_shares and own diagonal privat share which it kept for itself
//    The result is private share with number j  of the master secret polynomial,
//    which is not know by anybody, only by this node
// - It calculates public share from the private one
//
// Node's response (see AggregateDKSResponse struct)
// - Index is just for control
// - PubShare, calculated from private share
//
// After response from all nodes, dealer has all public information and nodes have all private informations.
// Key set is not saved yet!
// Next: see API call 'adm/commitdks'

func HandlerAggregateDks(c echo.Context) error {
	var req AggregateDKSRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &AggregateDKSResponse{
			Err: err.Error(),
		})
	}
	return misc.OkJson(c, AggregateDKSReq(&req))
}

type AggregateDKSRequest struct {
	TmpId     int      `json:"tmpId"`
	Index     uint16   `json:"index"`      // 0 to N-1
	PriShares []string `json:"pri_shares"` // base58
}

type AggregateDKSResponse struct {
	PubShare string `json:"pub_share"` // base58
	Err      string `json:"err"`
}

func AggregateDKSReq(req *AggregateDKSRequest) *AggregateDKSResponse {
	ks := getFromDkgCache(req.TmpId)
	if ks == nil {
		return &AggregateDKSResponse{Err: fmt.Sprintf("wrong tmpId %d", req.TmpId)}
	}
	if ks.Aggregated {
		return &AggregateDKSResponse{Err: "key set already aggregated"}
	}
	if len(req.PriShares) != int(ks.N) {
		return &AggregateDKSResponse{Err: "wrong number of private shares"}
	}
	if req.Index != ks.Index {
		return &AggregateDKSResponse{Err: "wrong index"}
	}
	// aggregate secret shares
	priShares := make([]kyber.Scalar, ks.N)
	for i, pks := range req.PriShares {
		if uint16(i) == ks.Index {
			continue
		}
		pkb, err := base58.Decode(pks)
		if err != nil {
			return &AggregateDKSResponse{Err: fmt.Sprintf("decode error: %v", err)}
		}
		priShares[i] = ks.Suite.G2().Scalar()
		if err := priShares[i].UnmarshalBinary(pkb); err != nil {
			return &AggregateDKSResponse{Err: fmt.Sprintf("unmarshal error: %v", err)}
		}
	}
	if err := ks.AggregateDKS(priShares); err != nil {
		return &AggregateDKSResponse{Err: fmt.Sprintf("aggregate error 1: %v", err)}
	}
	pkb, err := ks.PubKeyOwn.MarshalBinary()
	if err != nil {
		return &AggregateDKSResponse{Err: fmt.Sprintf("marshal error 2: %v", err)}
	}
	return &AggregateDKSResponse{
		PubShare: base58.Encode(pkb),
	}
}
