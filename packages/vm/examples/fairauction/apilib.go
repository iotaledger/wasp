package fairauction

import (
	"bytes"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/webapi/stateapi"
)

type Status struct {
	SCBalance map[balance.Color]int64
	FetchedAt time.Time

	OwnerMarginPromille int64
	AuctionsLen         uint32
	Auctions            map[balance.Color]*AuctionInfo
}

func FetchStatus(nodeclient nodeclient.NodeClient, waspHost string, scAddress *address.Address) (*Status, error) {
	status := &Status{
		FetchedAt: time.Now().UTC(),
	}

	scBalance, err := fetchSCBalance(nodeclient, scAddress)
	if err != nil {
		return nil, err
	}
	status.SCBalance = scBalance

	query := stateapi.NewQueryRequest(scAddress)
	query.AddInt64(VarStateOwnerMarginPromille)
	query.AddDictionary(VarStateAuctions, 100)

	results, err := waspapi.QuerySCState(waspHost, query)
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

func fetchSCBalance(nodeClient nodeclient.NodeClient, scAddress *address.Address) (map[balance.Color]int64, error) {
	outs, err := nodeClient.GetAccountOutputs(scAddress)
	if err != nil {
		return nil, err
	}
	ret, _ := util.OutputBalancesByColor(outs)
	return ret, nil
}

func StartAuction(nodeClient nodeclient.NodeClient, waspHost string, scAddress *address.Address) error {
	// TODO
	return nil
}
