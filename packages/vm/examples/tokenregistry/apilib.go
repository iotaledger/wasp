package tokenregistry

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
)

type MintAndRegisterParams struct {
	SenderSigScheme signaturescheme.SignatureScheme // source sig scheme
	Supply          int64                           // number of tokens to mint
	MintTarget      address.Address                 // where to mint new Supply
	RegistryAddr    address.Address                 // smart contract address
	Description     string
	UserDefinedData []byte
}

// MintAndRegister mints new Supply of colored tokens to some address and sends request
// to register it in the TokenRegistry smart contract
func MintAndRegister(node nodeclient.NodeClient, params MintAndRegisterParams) (*balance.Color, error) {
	ownerAddr := params.SenderSigScheme.Address()
	outs, err := node.GetAccountOutputs(&ownerAddr)
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
	codec.SetString(VarReqDescription, params.Description)
	if params.UserDefinedData != nil {
		codec.Set(VarReqUserDefinedMetadata, params.UserDefinedData)
	}

	reqBlk := sctransaction.NewRequestBlock(params.RegistryAddr, RequestMintSupply)
	reqBlk.SetArgs(args)
	err = txb.AddRequestBlock(reqBlk)
	if err != nil {
		return nil, err
	}
	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(params.SenderSigScheme)

	// TODO wait optionally
	err = node.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return nil, err
	}
	col := (balance.Color)(tx.ID())
	return &col, nil
}
