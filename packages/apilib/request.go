package apilib

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/subscribe"
	"strconv"
	"time"
)

type RequestBlockParams struct {
	TargetContractID coretypes.ContractID
	EntryPointCode   coretypes.EntryPointCode
	Timelock         uint32
	Transfer         map[balance.Color]int64 // should not not include request token. It is added automatically
	Vars             map[string]interface{}  ` `
}

type CreateRequestTransactionParams struct {
	NodeClient          nodeclient.NodeClient
	SenderSigScheme     signaturescheme.SignatureScheme
	BlockParams         []RequestBlockParams
	Mint                map[address.Address]int64
	Post                bool
	WaitForConfirmation bool
	WaitForCompletion   bool
	PublisherHosts      []string
	PublisherQuorum     int
	Timeout             time.Duration
}

func CreateRequestTransaction(par CreateRequestTransactionParams) (*sctransaction.Transaction, error) {
	senderAddr := par.SenderSigScheme.Address()
	allOuts, err := par.NodeClient.GetConfirmedAccountOutputs(&senderAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}

	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}

	for targetAddress, amount := range par.Mint {
		// TODO: check that targetAddress is not any target address in request blocks
		err = txb.MintColor(targetAddress, balance.ColorIOTA, amount)
		if err != nil {
			return nil, err
		}
	}

	for _, blockPar := range par.BlockParams {
		reqBlk := sctransaction.NewRequestBlockByWallet(blockPar.TargetContractID, blockPar.EntryPointCode).
			WithTimelock(blockPar.Timelock)

		args := convertArgs(blockPar.Vars)
		if args == nil {
			return nil, errors.New("wrong arguments")
		}
		reqBlk.SetArgs(args)

		err = txb.AddRequestSectionWithTransfer(reqBlk, blockPar.Transfer)
		if err != nil {
			return nil, err
		}
	}

	tx, err := txb.Build(false)

	//dump := txb.Dump()

	if err != nil {
		return nil, err
	}
	tx.Sign(par.SenderSigScheme)

	// semantic check just in case
	if _, err := tx.Properties(); err != nil {
		return nil, err
	}
	//fmt.Printf("$$$$ dumping builder for %s\n%s\n", tx.ID().String(), dump)

	if !par.Post {
		return tx, nil
	}
	if !par.WaitForConfirmation {
		if err = par.NodeClient.PostTransaction(tx.Transaction); err != nil {
			return nil, err
		}
		return tx, nil
	}
	var subs *subscribe.Subscription
	if par.WaitForCompletion {
		// post and wait for completion
		subs, err = subscribe.SubscribeMulti(par.PublisherHosts, "request_out", par.PublisherQuorum)
		if err != nil {
			return nil, err
		}
		defer subs.Close()
	}

	err = par.NodeClient.PostAndWaitForConfirmation(tx.Transaction)
	if err != nil {
		return nil, err
	}
	if par.WaitForCompletion {
		patterns := make([][]string, len(par.BlockParams))
		for i := range patterns {
			patterns[i] = []string{"request_out", par.BlockParams[i].TargetContractID.String(), tx.ID().String(), strconv.Itoa(i)}
		}
		if !subs.WaitForPatterns(patterns, par.Timeout, par.PublisherQuorum) {
			return nil, fmt.Errorf("didn't receive completion message after %v", par.Timeout)
		}
	}

	return tx, nil
}

func convertArgs(vars map[string]interface{}) dict.Dict {
	ret := dict.NewDict()
	args := codec.NewCodec(ret)
	for k, v := range vars {
		key := kv.Key(k)
		switch vt := v.(type) {
		case int:
			args.SetInt64(key, int64(vt))
		case byte:
			args.SetInt64(key, int64(vt))
		case int16:
			args.SetInt64(key, int64(vt))
		case int32:
			args.SetInt64(key, int64(vt))
		case int64:
			args.SetInt64(key, vt)
		case uint16:
			args.SetInt64(key, int64(vt))
		case uint32:
			args.SetInt64(key, int64(vt))
		case uint64:
			args.SetInt64(key, int64(vt))
		case string:
			args.SetString(key, vt)
		case []byte:
			args.Set(key, vt)
		case *hashing.HashValue:
			args.SetHashValue(key, vt)
		case *address.Address:
			args.Set(key, vt.Bytes())
		case *balance.Color:
			args.Set(key, vt.Bytes())
		case *coretypes.ChainID:
			args.SetChainID(key, vt)
		default:
			return nil
		}
	}
	return ret
}
