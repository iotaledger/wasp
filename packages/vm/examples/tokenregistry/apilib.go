package tokenregistry

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/subscribe"
	"sync"
	"time"
)

type MintAndRegisterParams struct {
	SenderSigScheme   signaturescheme.SignatureScheme // source sig scheme
	Supply            int64                           // number of tokens to mint
	MintTarget        address.Address                 // where to mint new Supply
	RegistryAddr      address.Address                 // smart contract address
	Description       string
	UserDefinedData   []byte
	WaitToBeProcessed bool
	PublisherHosts    []string
	Timeout           time.Duration
}

// MintAndRegister mints new Supply of colored tokens to some address and sends request
// to register it in the TokenRegistry smart contract
func MintAndRegister(node nodeclient.NodeClient, par MintAndRegisterParams) (*balance.Color, error) {
	ownerAddr := par.SenderSigScheme.Address()
	outs, err := node.GetAccountOutputs(&ownerAddr)
	if err != nil {
		return nil, err
	}
	txb, err := txbuilder.NewFromOutputBalances(outs)
	if err != nil {
		return nil, err
	}
	err = txb.MintColor(par.MintTarget, balance.ColorIOTA, par.Supply)
	if err != nil {
		return nil, err
	}
	args := kv.NewMap()
	codec := args.Codec()
	codec.SetString(VarReqDescription, par.Description)
	if par.UserDefinedData != nil {
		codec.Set(VarReqUserDefinedMetadata, par.UserDefinedData)
	}

	reqBlk := sctransaction.NewRequestBlock(par.RegistryAddr, RequestMintSupply)
	reqBlk.SetArgs(args)
	err = txb.AddRequestBlock(reqBlk)
	if err != nil {
		return nil, err
	}
	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(par.SenderSigScheme)

	err = node.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return nil, err
	}
	col := (balance.Color)(tx.ID())
	if !par.WaitToBeProcessed {
		return &col, nil
	}
	pattern := []string{"request_out", par.RegistryAddr.String(), tx.ID().String(), "0"}
	var wg sync.WaitGroup
	wg.Add(1)
	subscribe.ListenForPatternMulti(par.PublisherHosts, pattern, func(ok bool) {
		if !ok {
			err = fmt.Errorf("smart contract wasn't deployed correctly in %v", par.Timeout)
		}
		wg.Done()
	}, par.Timeout)
	if err != nil {
		return nil, nil
	}
	return &col, nil
}
