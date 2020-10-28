package apilib

import (
	"errors"
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/subscribe"
	"strconv"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
)

type CreateSimpleRequestParamsOld struct {
	TargetContract coretypes.ContractID
	RequestCode    coretypes.EntryPointCode
	Timelock       uint32
	Transfer       map[balance.Color]int64 // should not not include request token. It is added automatically
	Vars           map[string]interface{}  ` `
}

func CreateSimpleRequestOld(client nodeclient.NodeClient, sigScheme signaturescheme.SignatureScheme, par CreateSimpleRequestParamsOld) (*sctransaction.Transaction, error) {
	senderAddr := sigScheme.Address()
	allOuts, err := client.GetConfirmedAccountOutputs(&senderAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}

	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}

	reqBlk := sctransaction.NewRequestBlockByWallet(par.TargetContract, par.RequestCode).WithTimelock(par.Timelock)

	args := convertArgsOld(par.Vars)
	if args == nil {
		return nil, errors.New("wrong arguments")
	}
	reqBlk.SetArgs(args)

	err = txb.AddRequestBlockWithTransfer(reqBlk, par.Transfer)
	if err != nil {
		return nil, err
	}

	tx, err := txb.Build(false)

	//dump := txb.Dump()

	if err != nil {
		return nil, err
	}
	tx.Sign(sigScheme)

	// check semantic just in case
	if _, err := tx.Properties(); err != nil {
		return nil, err
	}
	//fmt.Printf("$$$$ dumping builder for %s\n%s\n", tx.ID().String(), dump)

	return tx, nil
}

func CreateSimpleRequestMultiOld(client nodeclient.NodeClient, sigScheme signaturescheme.SignatureScheme, pars []CreateSimpleRequestParamsOld) (*sctransaction.Transaction, error) {
	senderAddr := sigScheme.Address()
	allOuts, err := client.GetConfirmedAccountOutputs(&senderAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}

	txb, err := txbuilder.NewFromOutputBalances(allOuts)
	if err != nil {
		return nil, err
	}

	for _, par := range pars {
		reqBlk := sctransaction.NewRequestBlockByWallet(par.TargetContract, par.RequestCode).WithTimelock(par.Timelock)

		args := convertArgsOld(par.Vars)
		if args == nil {
			return nil, errors.New("wrong arguments")
		}
		reqBlk.SetArgs(args)

		err = txb.AddRequestBlockWithTransfer(reqBlk, par.Transfer)
		if err != nil {
			return nil, err
		}
	}

	tx, err := txb.Build(false)

	if err != nil {
		return nil, err
	}
	tx.Sign(sigScheme)

	// check semantic just in case
	if _, err := tx.Properties(); err != nil {
		return nil, err
	}

	return tx, nil
}

func convertArgsOld(vars map[string]interface{}) kv.Map {
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
	Address     string                   `json:"address"`
	RequestCode coretypes.EntryPointCode `json:"request_code"`
	Timelock    uint32                   `json:"timelock"`
	AmountIotas int64                    `json:"amount_iotas"` // minimum 1 iota will be taken anyway
	Vars        map[string]interface{}   `json:"vars"`
}

// Deprecated
func CreateRequestTransactionOld(client nodeclient.NodeClient, senderSigScheme signaturescheme.SignatureScheme, reqsJson []*RequestBlockJson) (*sctransaction.Transaction, error) {
	var err error
	if len(reqsJson) == 0 {
		return nil, errors.New("CreateRequestTransactionOld: must be at least 1 request block")
	}
	senderAddr := senderSigScheme.Address()
	allOuts, err := client.GetConfirmedAccountOutputs(&senderAddr)
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
			if err = txb.MoveTokensToAddress(reqBlk.Address(), balance.ColorIOTA, reqBlkJson.AmountIotas-1); err != nil {
				return nil, err
			}
		}
	}
	tx, err := txb.Build(false)
	if err != nil {
		return nil, err
	}
	tx.Sign(senderSigScheme)

	// check semantic just in case
	if _, err := tx.Properties(); err != nil {
		return nil, err
	}

	MustNotNullInputs(tx.Transaction)

	return tx, nil
}

func requestBlockFromJson(reqBlkJson *RequestBlockJson) (*sctransaction.RequestBlock, error) {
	var err error
	addr, err := address.FromBase58(reqBlkJson.Address)
	if err != nil {
		return nil, err
	}
	ret := sctransaction.NewRequestBlock(addr, reqBlkJson.RequestCode)
	ret.WithTimelock(reqBlkJson.Timelock)

	if reqBlkJson.Vars == nil {
		// no args
		return ret, nil
	}

	args := convertArgsOld(reqBlkJson.Vars)
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

func WaitForRequestProcessedMulti(hosts []string, scAddr *address.Address, txid *valuetransaction.ID, reqIndex uint16, timeout time.Duration) error {
	var err error
	pattern := []string{"request_out", scAddr.String(), txid.String(), strconv.Itoa(int(reqIndex))}
	var wg sync.WaitGroup
	wg.Add(1)
	subscribe.ListenForPatternMulti(hosts, pattern, func(ok bool) {
		if !ok {
			err = fmt.Errorf("request [%d]%s wasn't processed in %v", reqIndex, txid.String(), timeout)
		}
		wg.Done()
	}, timeout)
	return err
}

func RunAndWaitForRequestProcessedMulti(hosts []string, scAddr *address.Address, reqIndex uint16, timeout time.Duration, f func() (*sctransaction.Transaction, error)) (*sctransaction.Transaction, error) {
	var tx *sctransaction.Transaction
	err := subscribe.SubscribeRunAndWaitForPattern(hosts, "request_out", timeout, func() ([]string, error) {
		var err error
		tx, err = f()
		if err != nil {
			return nil, fmt.Errorf("failed to create tx: %v", err)
		}
		return []string{"request_out", scAddr.String(), tx.ID().String(), strconv.Itoa(int(reqIndex))}, nil
	})
	if err != nil {
		if tx != nil {
			return nil, fmt.Errorf("request [%d]%s failed: %v", reqIndex, tx.ID(), err)
		}
		return nil, err
	}
	return tx, nil
}
