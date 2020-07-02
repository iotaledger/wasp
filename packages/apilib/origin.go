package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/table"
)

type CreateOriginParams struct {
	Address              address.Address
	OwnerSignatureScheme signaturescheme.SignatureScheme
	ProgramHash          hashing.HashValue
	Variables            table.MemTable
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
	return origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              par.Address,
		OwnerSignatureScheme: par.OwnerSignatureScheme,
		AllInputs:            allOuts,
		InputColor:           balance.ColorIOTA,
		ProgramHash:          par.ProgramHash,
	})
}

// same as above only gets inputs form local utxodb rather than goshimmer
// deterministic if applied to different owner addresses hardcoded in utxodb
func CreateOriginUtxodb(par CreateOriginParams) (*sctransaction.Transaction, error) {
	ownerAddress := par.OwnerSignatureScheme.Address()
	allOuts := utxodb.GetAddressOutputs(ownerAddress)
	return origin.NewOriginTransaction(origin.NewOriginTransactionParams{
		Address:              par.Address,
		OwnerSignatureScheme: par.OwnerSignatureScheme,
		AllInputs:            allOuts,
		InputColor:           balance.ColorIOTA,
		ProgramHash:          par.ProgramHash,
	})
}
