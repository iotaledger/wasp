package fairauction

import (
	"bytes"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
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
	Auctions            map[balance.Color]*AuctionInfo
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
	query.AddInt64(VarStateOwnerMarginPromille)
	query.AddDictionary(VarStateAuctions, 100)

	results, err := waspapi.QuerySCState(fc.waspHost, query)
	if err != nil {
		return nil, err
	}

	status.OwnerMarginPromille = getOwnerMarginPromille(results[VarStateOwnerMarginPromille].Int64())

	auctions := results[VarStateAuctions].MustDictionaryResult()
	status.AuctionsLen = auctions.Len
	status.Auctions = make(map[balance.Color]*AuctionInfo)
	for _, entry := range auctions.Entries {
		ai := &AuctionInfo{}
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

func (frc *FairAuctionClient) postRequest(code sctransaction.RequestCode, transfer map[balance.Color]int64, vars map[string]interface{}) error {
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
		return err
	}
	return frc.nodeClient.PostTransaction(tx.Transaction)
}

func (fc *FairAuctionClient) SetOwnerMargin(margin int64) error {
	return fc.postRequest(
		RequestStartAuction,
		nil,
		map[string]interface{}{VarReqOwnerMargin: margin},
	)
}

func (fc *FairAuctionClient) StartAuction(color *balance.Color, description string, minimumBid int64, durationMinutes int64, transfer map[balance.Color]int64) error {
	return fc.postRequest(
		RequestStartAuction,
		transfer,
		map[string]interface{}{
			VarReqAuctionColor:                color,
			VarReqStartAuctionDescription:     description,
			VarReqStartAuctionMinimumBid:      minimumBid,
			VarReqStartAuctionDurationMinutes: durationMinutes,
		},
	)
}

func (fc *FairAuctionClient) PlaceBid(color *balance.Color, amountIotas int64) error {
	return fc.postRequest(
		RequestStartAuction,
		map[balance.Color]int64{balance.ColorIOTA: amountIotas},
		map[string]interface{}{VarReqAuctionColor: color},
	)
}
