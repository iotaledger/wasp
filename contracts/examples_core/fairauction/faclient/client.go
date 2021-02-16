// +build ignore

package faclient

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/model/statequery"
)

type FairAuctionClient struct {
	*chainclient.Client
	contractHname coretypes.Hname
}

func NewClient(scClient *chainclient.Client, contractHname coretypes.Hname) *FairAuctionClient {
	return &FairAuctionClient{
		Client:        scClient,
		contractHname: contractHname,
	}
}

type Status struct {
	*chainclient.SCStatus

	OwnerMarginPromille int64
	AuctionsLen         uint32
	Auctions            map[balance.Color]*fairauction.AuctionInfo
}

func (fc *FairAuctionClient) FetchStatus() (*Status, error) {
	scStatus, results, err := fc.FetchSCStatus(func(query *statequery.Request) {
		query.AddScalar(fairauction.VarStateOwnerMarginPromille)
		query.AddMap(fairauction.VarStateAuctions, 100)
	})
	if err != nil {
		return nil, err
	}

	status := &Status{SCStatus: scStatus}

	ownerMargin, ok := results.Get(fairauction.VarStateOwnerMarginPromille).MustInt64()
	status.OwnerMarginPromille = fairauction.GetOwnerMarginPromille(ownerMargin, ok, nil)

	auctions := results.Get(fairauction.VarStateAuctions).MustMapResult()
	status.AuctionsLen = auctions.Len
	status.Auctions = make(map[balance.Color]*fairauction.AuctionInfo)
	for _, entry := range auctions.Entries {
		ai := &fairauction.AuctionInfo{}
		if err := ai.Read(bytes.NewReader(entry.Value)); err != nil {
			return nil, err
		}
		status.Auctions[ai.Color] = ai
	}

	return status, nil
}

func (fc *FairAuctionClient) SetOwnerMargin(margin int64) (*sctransaction.Transaction, error) {
	return fc.PostRequest(
		fc.contractHname,
		fairauction.RequestSetOwnerMargin,
		chainclient.PostRequestParams{
			ArgsRaw: codec.MakeDict(map[string]interface{}{fairauction.VarReqOwnerMargin: margin}),
		},
	)
}

func (fc *FairAuctionClient) GetFeeAmount(minimumBid int64) (int64, error) {
	query := statequery.NewRequest()
	query.AddScalar(fairauction.VarStateOwnerMarginPromille)
	res, err := fc.StateQuery(query)
	var ownerMarginState int64
	var ok bool
	if model.IsHTTPNotFound(err) {
		if err != nil {
			return 0, err
		}
		ownerMarginState, ok = res.Get(fairauction.VarStateOwnerMarginPromille).MustInt64()
	}
	ownerMargin := fairauction.GetOwnerMarginPromille(ownerMarginState, ok, nil)
	fee := fairauction.GetExpectedDeposit(minimumBid, ownerMargin)
	return fee, nil
}

func (fc *FairAuctionClient) StartAuction(
	description string,
	color *balance.Color,
	tokensForSale int64,
	minimumBid int64,
	durationMinutes int64,
) (*sctransaction.Transaction, error) {
	fee, err := fc.GetFeeAmount(minimumBid)
	if err != nil {
		return nil, fmt.Errorf("GetFeeAmount failed: %v", err)
	}
	return fc.PostRequest(
		fc.contractHname,
		fairauction.RequestStartAuction,
		chainclient.PostRequestParams{
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: fee,
				*color:            tokensForSale,
			},
			ArgsRaw: codec.MakeDict(map[string]interface{}{
				fairauction.VarReqAuctionColor:                color.String(),
				fairauction.VarReqStartAuctionDescription:     description,
				fairauction.VarReqStartAuctionMinimumBid:      minimumBid,
				fairauction.VarReqStartAuctionDurationMinutes: durationMinutes,
			}),
		},
	)
}

func (fc *FairAuctionClient) PlaceBid(color *balance.Color, amountIotas int64) (*sctransaction.Transaction, error) {
	return fc.PostRequest(
		fc.contractHname,
		fairauction.RequestPlaceBid,
		chainclient.PostRequestParams{
			Transfer: map[balance.Color]int64{balance.ColorIOTA: amountIotas},
			ArgsRaw:  codec.MakeDict(map[string]interface{}{fairauction.VarReqAuctionColor: color.String()}),
		},
	)
}
