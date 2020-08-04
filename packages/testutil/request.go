package testutil

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
)

const RequestFundsAmount = utxodb.RequestFundsAmount

func RequestFunds(goshimmerHost string, targetAddress address.Address) error {
	// TODO: allow using the Faucet API to request funds
	return nodeapi.RequestFunds(goshimmerHost, &targetAddress)
}
