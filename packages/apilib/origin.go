package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/packages/waspconn/apilib"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/variables"
)

type CreateOriginParams struct {
	Address              address.Address
	OwnerSignatureScheme signaturescheme.SignatureScheme
	ProgramHash          hashing.HashValue
	Variables            variables.Variables
}

// CreateOrigin creates origin transaction. It asks for inputs from goshimmer node
// origin transaction approves origin state.
func CreateOrigin(nodeurl string, par CreateOriginParams) (*sctransaction.Transaction, error) {
	ownerAddress := par.OwnerSignatureScheme.Address()
	// get outputs from goshimmer
	allOuts, err := nodeapi.GetAccountOutputs(nodeurl, &ownerAddress)
	if err != nil {
		return nil, err
	}
	return createOrigin(par, allOuts)
}

// same as above only gets inputs form local utxodb rather than goshimmer
// deterministic if applied to different owner addresses hardcoded in utxodb
func CreateOriginUtxodb(par CreateOriginParams) (*sctransaction.Transaction, error) {
	ownerAddress := par.OwnerSignatureScheme.Address()
	allOuts := utxodb.GetAddressOutputs(ownerAddress)
	return createOrigin(par, allOuts)
}

func createOrigin(par CreateOriginParams, allOutputs map[valuetransaction.OutputID][]*balance.Balance) (*sctransaction.Transaction, error) {

	originTx, err := origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              par.Address,
		OwnerSignatureScheme: par.OwnerSignatureScheme,
		AllInputs:            allOutputs,
		InputColor:           balance.ColorIOTA,
	})
	if err != nil {
		return nil, err
	}

	return originTx, nil
}
