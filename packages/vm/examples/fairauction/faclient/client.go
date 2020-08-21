package faclient

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
)

type FairAuctionClient struct {
	nodeClient nodeclient.NodeClient
	waspHost   string
	scAddress  *address.Address
	sigScheme  signaturescheme.SignatureScheme
}

func NewClient(nodeClient nodeclient.NodeClient, waspHost string, scAddress *address.Address, sigScheme signaturescheme.SignatureScheme) *FairAuctionClient {
	return &FairAuctionClient{nodeClient, waspHost, scAddress, sigScheme}
}

type Status struct {
	SCBalance map[balance.Color]int64
	FetchedAt time.Time

	OwnerMarginPromille int64
	AuctionsLen         uint32
	Auctions            map[balance.Color]*fairauction.AuctionInfo
}

func (fc *FairAuctionClient) FetchStatus() (*Status, error) {
	status := &Status{
		FetchedAt: time.Now().UTC(),
	}

	scBalance, err := fc.fetchSCBalance()
	if err != nil {
		return nil, err
	}
	status.SCBalance = scBalance

	query := stateapi.NewQueryRequest(fc.scAddress)
	query.AddInt64(fairauction.VarStateOwnerMarginPromille)
	query.AddDictionary(fairauction.VarStateAuctions, 100)

	results, err := waspapi.QuerySCState(fc.waspHost, query)
	if err != nil {
		return nil, err
	}

	ownerMargin, ok, err := results[fairauction.VarStateOwnerMarginPromille].MustInt64()
	if err != nil {
		return nil, err
	}
	status.OwnerMarginPromille = fairauction.GetOwnerMarginPromille(ownerMargin, ok)

	auctions := results[fairauction.VarStateAuctions].MustDictionaryResult()
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

func (fc *FairAuctionClient) fetchSCBalance() (map[balance.Color]int64, error) {
	outs, err := fc.nodeClient.GetAccountOutputs(fc.scAddress)
	if err != nil {
		return nil, err
	}
	ret, _ := util.OutputBalancesByColor(outs)
	return ret, nil
}

func (frc *FairAuctionClient) postRequest(code sctransaction.RequestCode, transfer map[balance.Color]int64, vars map[string]interface{}) (*sctransaction.Transaction, error) {
	tx, err := waspapi.CreateSimpleRequest(
		frc.nodeClient,
		frc.sigScheme,
		waspapi.CreateSimpleRequestParams{
			SCAddress:   frc.scAddress,
			RequestCode: code,
			Vars:        vars,
			Transfer:    transfer,
		},
	)
	if err != nil {
		return nil, err
	}
	return tx, frc.nodeClient.PostTransaction(tx.Transaction)
}

func (fc *FairAuctionClient) SetOwnerMargin(margin int64) (*sctransaction.Transaction, error) {
	return fc.postRequest(
		fairauction.RequestSetOwnerMargin,
		nil,
		map[string]interface{}{fairauction.VarReqOwnerMargin: margin},
	)
}

func (fc *FairAuctionClient) GetFeeAmount(minimumBid int64) (int64, error) {
	query := stateapi.NewQueryRequest(fc.scAddress)
	query.AddInt64(fairauction.VarStateOwnerMarginPromille)
	results, err := waspapi.QuerySCState(fc.waspHost, query)
	var ownerMarginState int64
	var ok bool
	if err != waspapi.ErrStateNotFound {
		if err != nil {
			return 0, err
		}
		ownerMarginState, ok, err = results[fairauction.VarStateOwnerMarginPromille].MustInt64()
		if err != nil {
			return 0, err
		}
	}
	ownerMargin := fairauction.GetOwnerMarginPromille(ownerMarginState, ok)
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
	return fc.postRequest(
		fairauction.RequestStartAuction,
		map[balance.Color]int64{
			balance.ColorIOTA: fee,
			*color:            tokensForSale,
		},
		map[string]interface{}{
			fairauction.VarReqAuctionColor:                color,
			fairauction.VarReqStartAuctionDescription:     description,
			fairauction.VarReqStartAuctionMinimumBid:      minimumBid,
			fairauction.VarReqStartAuctionDurationMinutes: durationMinutes,
		},
	)
}

func (fc *FairAuctionClient) PlaceBid(color *balance.Color, amountIotas int64) (*sctransaction.Transaction, error) {
	return fc.postRequest(
		fairauction.RequestPlaceBid,
		map[balance.Color]int64{balance.ColorIOTA: amountIotas},
		map[string]interface{}{fairauction.VarReqAuctionColor: color},
	)
}
