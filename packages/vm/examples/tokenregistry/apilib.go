package tokenregistry

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
)

type MintAndRegisterParams struct {
	senderSigScheme signaturescheme.SignatureScheme // source sig scheme
	supply          int64                           // number of tokens to mint
	tokenTarget     address.Address                 // where to mint new supply
	registryAddr    address.Address                 // smart contract address
	description     string
	userDefinedData []byte
}

//
//// MintAndRegister mints new supply of colored tokens to some address and sends request
//// to register it in the TokenRegistry smart contract
//func MintAndRegister(node nodeclient.NodeClient, params MintAndRegisterParams) error{
//	ownerAddr := params.senderSigScheme.Address()
//	outs, err := node.GetAccountOutputs(&ownerAddr)
//	if err != nil{
//		return err
//	}
//	txb, err := txbuilder.NewFromOutputBalances(outs)
//	if err != nil{
//		return err
//	}
//	err = txb.MintColor(params.tokenTarget, balance.ColorIOTA, params.supply)
//	if err != nil{
//		return err
//	}
//	args := kv.NewMap()
//	codec := args.Codec()
//	codec.SetString(VarReqDescription, params.description)
//	codec.Set(VarReqUserDefinedMetadata, params.userDefinedData)
//
//	reqBlk := sctransaction.NewRequestBlock(params.registryAddr, RequestMintSupply)
//	reqBlk.SetArgs(args)
//	err = txb.AddRequestBlock(reqBlk)
//	if err != nil{
//		return err
//	}
//	tx, err := txb.Build(false)
//	if err != nil{
//		return err
//	}
//	// TODO wait optionally
//	err = node.PostAndWaitForConfirmation(tx.Transaction)
//	if err != nil{
//		return err
//	}
//	return nil
//}
