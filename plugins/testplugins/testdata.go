package testplugins

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
)

// BLS addresses
const (
	blsaddr1 = "exZup69X1XwRNHiWWjoYy75aPNgC22YKkPV7sUJSBYA9"
	blsaddr2 = "dV9hfYyHq7uiCKdKYQoLqyiwX6tN448GRm8UgFpUC3Vo"
	blsaddr3 = "eiMbhrJjajqnCLmVJqFXzFsh1ZsbCAnJ9wauU8cP8uxL"
)

var (
	addr1 address.Address
	addr2 address.Address
	addr3 address.Address
	SC1   apilib.NewOriginParams
	SC2   apilib.NewOriginParams
	SC3   apilib.NewOriginParams
)

func init() {
	var err error

	addr1, err = address.FromBase58(blsaddr1)
	if err != nil {
		panic(err)
	}
	addr2, err = address.FromBase58(blsaddr2)
	if err != nil {
		panic(err)
	}
	addr3, err = address.FromBase58(blsaddr3)
	if err != nil {
		panic(err)
	}
	ownerAddress := utxodb.GetAddress(1)
	SC1 = apilib.NewOriginParams{
		Address:      &addr1,
		OwnerAddress: &ownerAddress,
		Description:  "Test smart contract 1 one",
	}
	SC1.ProgramHash = hashing.HashStrings(SC1.Description)

	ownerAddress = utxodb.GetAddress(2)
	SC2 = apilib.NewOriginParams{
		Address:      &addr2,
		OwnerAddress: &ownerAddress,
		Description:  "Test smart contract 2 two",
	}
	SC2.ProgramHash = hashing.HashStrings(SC2.Description)

	ownerAddress = utxodb.GetAddress(3)
	SC3 = apilib.NewOriginParams{
		Address:      &addr3,
		OwnerAddress: &ownerAddress,
		Description:  "Test smart contract 3 three",
	}
	SC3.ProgramHash = hashing.HashStrings(SC3.Description)
}

func CreateOriginData(par apilib.NewOriginParams, nodeLocations []string) (*valuetransaction.Transaction, *registry.SCMetaData) {
	allOuts := utxodb.GetAddressOutputs(*par.OwnerAddress)
	outs := apilib.SelectMinimumOutputs(allOuts, balance.ColorIOTA, 1)
	if len(outs) == 0 {
		panic("inconsistency: not enough outputs for 1 iota!")
	}
	// select first and the only
	var input valuetransaction.OutputID
	var inputBalances []*balance.Balance

	for oid, v := range outs {
		input = oid
		inputBalances = v
		break
	}

	originTx, err := apilib.NewOriginTransaction(apilib.NewOriginTransactionParams{
		NewOriginParams: par,
		Input:           input,
		InputBalances:   inputBalances,
		InputColor:      balance.ColorIOTA,
		OwnerSigScheme:  utxodb.GetSigScheme(*par.OwnerAddress),
	})
	if err != nil {
		panic(err)
	}
	if nodeLocations == nil {
		return originTx, nil
	}
	scdata := &registry.SCMetaData{
		Address:       *par.Address,
		Color:         balance.Color(originTx.ID()),
		OwnerAddress:  *par.OwnerAddress,
		Description:   par.Description,
		ProgramHash:   *par.ProgramHash,
		NodeLocations: nodeLocations,
	}
	return originTx, scdata
}
