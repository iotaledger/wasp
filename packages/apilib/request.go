package apilib

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
)

type CreateSimpleRequestParams struct {
	SCAddress   *address.Address
	RequestCode sctransaction.RequestCode
	Timelock    uint32
	Transfer    map[balance.Color]int64 // should not not include request token. It is added automatically
	Vars        map[string]interface{}  ` `
}

func CreateSimpleRequest(node string, sigScheme signaturescheme.SignatureScheme, par CreateSimpleRequestParams) (*sctransaction.Transaction, error) {
	senderAddr := sigScheme.Address()
	allOuts, err := nodeapi.GetAccountOutputs(node, &senderAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}

	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}

	reqBlk := sctransaction.NewRequestBlock(*par.SCAddress, par.RequestCode).WithTimelock(par.Timelock)

	args := convertArgs(par.Vars)
	if args == nil {
		return nil, errors.New("wrong arguments")
	}
	reqBlk.SetArgs(args)

	err = txb.AddRequestBlockWithTransfer(reqBlk, par.SCAddress, par.Transfer)
	if err != nil {
		return nil, err
	}

	tx, err := txb.Build(false)

	dump := txb.Dump()

	if err != nil {
		return nil, err
	}
	tx.Sign(sigScheme)

	fmt.Printf("$$$$ dumping builder for %s\n%s\n", tx.ID().String(), dump)

	return tx, nil
}

func convertArgs(vars map[string]interface{}) kv.Map {
	args := kv.NewMap()
	codec := args.Codec()
	for k, v := range vars {
		key := kv.Key(k)
		switch vt := v.(type) {
		case int:
			codec.SetInt64(key, int64(vt))
		case byte:
			codec.SetInt64(key, int64(vt))
		case int16:
			codec.SetInt64(key, int64(vt))
		case int32:
			codec.SetInt64(key, int64(vt))
		case int64:
			codec.SetInt64(key, vt)
		case uint16:
			codec.SetInt64(key, int64(vt))
		case uint32:
			codec.SetInt64(key, int64(vt))
		case uint64:
			codec.SetInt64(key, int64(vt))
		case string:
			codec.SetString(key, vt)
		case []byte:
			codec.Set(key, vt)
		case *hashing.HashValue:
			args.Codec().SetHashValue(key, vt)
		case *address.Address:
			args.Codec().Set(key, vt.Bytes())
		case *balance.Color:
			args.Codec().Set(key, vt.Bytes())
		default:
			return nil
		}
	}
	return args
}

// Deprecated
type RequestBlockJson struct {
	Address     string                    `json:"address"`
	RequestCode sctransaction.RequestCode `json:"request_code"`
	Timelock    uint32                    `json:"timelock"`
	AmountIotas int64                     `json:"amount_iotas"` // minimum 1 iota will be taken anyway
	Vars        map[string]interface{}    `json:"vars"`
}

// Deprecated
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

	args := convertArgs(reqBlkJson.Vars)
	if args == nil {
		return nil, errors.New("wrong arguments")
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
