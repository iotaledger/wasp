package client

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client/jsonable"
)

const DKSNewRoute = "dks/new"
const DKSAggregateRoute = "dks/aggregate"
const DKSCommitRoute = "dks/commit"

type NewDKSRequest struct {
	TmpId int    `json:"tmpId"`
	N     uint16 `json:"n"`
	T     uint16 `json:"t"`
	Index uint16 `json:"index"` // 0 to N-1
}

func (c *WaspClient) NewDKShare(params NewDKSRequest) ([]string, error) {
	var res []string
	err := c.do(http.MethodPost, AdminRoutePrefix+"/"+DKSNewRoute, &params, &res)
	return res, err
}

type AggregateDKSRequest struct {
	TmpId     int      `json:"tmpId"`
	Index     uint16   `json:"index"`      // 0 to N-1
	PriShares []string `json:"pri_shares"` // base58
}

func (c *WaspClient) AggregateDKShare(params AggregateDKSRequest) (string, error) {
	var res string
	err := c.do(http.MethodPost, AdminRoutePrefix+"/"+DKSAggregateRoute, &params, &res)
	return res, err
}

type CommitDKSRequest struct {
	TmpId     int      `json:"tmpId"`
	PubShares []string `json:"pub_shares"` // base58
}

func (c *WaspClient) CommitDKShare(params CommitDKSRequest) (*address.Address, error) {
	var res jsonable.Address
	if err := c.do(http.MethodPost, AdminRoutePrefix+"/"+DKSCommitRoute, &params, &res); err != nil {
		return nil, err
	}
	return res.Address(), nil
}
