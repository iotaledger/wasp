package trclient

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
)

type TokenRegistryClient struct {
	nodeClient nodeclient.NodeClient
	waspHost   string
	scAddress  *address.Address
	sigScheme  signaturescheme.SignatureScheme
}

func NewClient(nodeClient nodeclient.NodeClient, waspHost string, scAddress *address.Address, sigScheme signaturescheme.SignatureScheme) *TokenRegistryClient {
	return &TokenRegistryClient{nodeClient, waspHost, scAddress, sigScheme}
}

type MintAndRegisterParams struct {
	Supply          int64           // number of tokens to mint
	MintTarget      address.Address // where to mint new Supply
	Description     string
	UserDefinedData []byte
}

// MintAndRegister mints new Supply of colored tokens to some address and sends request
// to register it in the TokenRegistry smart contract
func (tc *TokenRegistryClient) MintAndRegister(params MintAndRegisterParams) (*balance.Color, error) {
	ownerAddr := tc.sigScheme.Address()
	outs, err := tc.nodeClient.GetAccountOutputs(&ownerAddr)
	if err != nil {
		return nil, err
	}
	txb, err := txbuilder.NewFromOutputBalances(outs)
	if err != nil {
		return nil, err
	}
	err = txb.MintColor(params.MintTarget, balance.ColorIOTA, params.Supply)
	if err != nil {
		return nil, err
	}
	args := kv.NewMap()
	codec := args.Codec()
	codec.SetString(tokenregistry.VarReqDescription, params.Description)
	if params.UserDefinedData != nil {
		codec.Set(tokenregistry.VarReqUserDefinedMetadata, params.UserDefinedData)
	}

	reqBlk := sctransaction.NewRequestBlock(*tc.scAddress, tokenregistry.RequestMintSupply)
	reqBlk.SetArgs(args)
	err = txb.AddRequestBlock(reqBlk)
	if err != nil {
		return nil, err
	}
	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(tc.sigScheme)

	// TODO wait optionally
	err = tc.nodeClient.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return nil, err
	}
	col := (balance.Color)(tx.ID())
	return &col, nil
}
