package apilib

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"strconv"
)

type RequestBlockJson struct {
	Address     string                    `json:"address"`
	RequestCode sctransaction.RequestCode `json:"request_code"`
	Timelock    uint32                    `json:"timelock"`
	AmountIotas int64                     `json:"amount_iotas"` // minimum 1 iota will be taken anyway
	Vars        map[string]interface{}    `json:"vars"`
}

func CreateRequestTransaction(node string, senderSigScheme signaturescheme.SignatureScheme, reqsJson []*RequestBlockJson) (*sctransaction.Transaction, error) {
	var err error
	if len(reqsJson) == 0 {
		return nil, errors.New("CreateRequestTransaction: must be at least 1 request block")
	}
	senderAddr := senderSigScheme.Address()
	allOuts, err := nodeapi.GetAccountOutputs(node, &senderAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}

	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}

	for _, reqBlkJson := range reqsJson {
		reqBlk, err := requestBlockFromJson(reqBlkJson)
		if err != nil {
			return nil, err
		}
		// add request block. It also move 1 iota as ColorNew for request token
		if err = txb.AddRequestBlock(reqBlk); err != nil {
			return nil, err
		}
		if reqBlkJson.AmountIotas > 1 {
			if err = txb.MoveToAddress(reqBlk.Address(), balance.ColorIOTA, reqBlkJson.AmountIotas-1); err != nil {
				return nil, err
			}
		}
	}
	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(senderSigScheme)

	MustNotNullInputs(tx.Transaction)

	return tx, nil
}

func requestBlockFromJson(reqBlkJson *RequestBlockJson) (*sctransaction.RequestBlock, error) {
	var err error
	addr, err := address.FromBase58(reqBlkJson.Address)
	if err != nil {
		return nil, err
	}
	ret := sctransaction.NewRequestBlock(addr, sctransaction.RequestCode(reqBlkJson.RequestCode))
	ret.WithTimelock(reqBlkJson.Timelock)

	if reqBlkJson.Vars == nil {
		// no args
		return ret, nil
	}

	args := kv.NewMap()
	for k, v := range reqBlkJson.Vars {
		switch vt := v.(type) {
		case int:
			args.Codec().SetInt64(kv.Key(k), int64(vt))
		case byte:
			args.Codec().SetInt64(kv.Key(k), int64(vt))
		case int16:
			args.Codec().SetInt64(kv.Key(k), int64(vt))
		case int32:
			args.Codec().SetInt64(kv.Key(k), int64(vt))
		case int64:
			args.Codec().SetInt64(kv.Key(k), vt)
		case uint16:
			args.Codec().SetInt64(kv.Key(k), int64(vt))
		case uint32:
			args.Codec().SetInt64(kv.Key(k), int64(vt))
		case uint64:
			args.Codec().SetInt64(kv.Key(k), int64(vt))
		case string:
			if n, err := strconv.Atoi(vt); err == nil {
				args.Codec().SetInt64(kv.Key(k), int64(n))
			} else {
				args.Codec().SetString(kv.Key(k), vt)
			}
		}
	}
	ret.SetArgs(args)

	return ret, nil
}

func MustNotNullInputs(tx *valuetransaction.Transaction) {
	var nilid valuetransaction.ID
	tx.Inputs().ForEach(func(outputId valuetransaction.OutputID) bool {
		if outputId.TransactionID() == nilid {
			panic(fmt.Sprintf("nil input in txid %s", tx.ID().String()))
		}
		return true
	})
}
