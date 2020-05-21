package apilib

import (
	"encoding/json"
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"strconv"
)

type RequestBlockJson struct {
	Address string            `json:"address"`
	Vars    map[string]string `json:"vars"`
}

type RequestTransactionJson struct {
	RequestorIndex int                `json:"requestor_index"` // from 1-10, corresponds to one of hardcodded addresses. 0 means 1
	Requests       []RequestBlockJson `json:"requests"`        // can't be empty
}

func RequestBlockFromJson(reqBlkJson *RequestBlockJson) (*sctransaction.RequestBlock, error) {
	var err error
	addr, err := address.FromBase58(reqBlkJson.Address)
	if err != nil {
		return nil, err
	}
	ret := sctransaction.NewRequestBlock(addr)

	for k, v := range reqBlkJson.Vars {
		n, err := strconv.Atoi(v)
		if err != nil {
			ret.Variables().Set(k, v)
		} else {
			ret.Variables().Set(k, uint16(n))
		}
	}
	return ret, nil
}

func TransactionFromJson(data []byte) (*sctransaction.Transaction, error) {
	var err error
	jsonObj := RequestTransactionJson{}
	if err = json.Unmarshal(data, &jsonObj); err != nil {
		return nil, err
	}
	if jsonObj.RequestorIndex <= 0 {
		jsonObj.RequestorIndex = 1
	}
	if jsonObj.RequestorIndex > 10 {
		return nil, errors.New("wrong input params")
	}
	if len(jsonObj.Requests) == 0 {
		return nil, errors.New("wrong input params")
	}
	inputAddr := utxodb.GetAddress(jsonObj.RequestorIndex)
	allOuts := utxodb.GetAddressOutputs(inputAddr)
	selectedOutputs := util.SelectMinimumOutputs(allOuts, balance.ColorIOTA, int64(len(jsonObj.Requests)))
	if len(selectedOutputs) == 0 {
		return nil, errors.New("not enough funds")
	}
	oids := make([]valuetransaction.OutputID, 0, len(selectedOutputs))
	for oid := range selectedOutputs {
		oids = append(oids, oid)
	}

	txb := sctransaction.NewTransactionBuilder()
	txb.AddInputs(oids...)

	for _, reqBlkJson := range jsonObj.Requests {
		reqBlk, err := RequestBlockFromJson(&reqBlkJson)
		if err != nil {
			return nil, err
		}
		// add request block
		txb.AddRequestBlock(reqBlk)
		// add new request token to outputs
		txb.AddBalanceToOutput(reqBlk.Address(), balance.New(balance.ColorNew, 1))
	}
	total := TotalBalanceOfInputs(selectedOutputs)
	if total > int64(len(jsonObj.Requests)) {
		// add reminder
		txb.AddBalanceToOutput(inputAddr, balance.New(balance.ColorIOTA, total-int64(len(jsonObj.Requests))))
	}
	tx, err := txb.Finalize()
	if err != nil {
		return nil, err
	}
	tx.Sign(utxodb.GetSigScheme(inputAddr))
	return tx, nil
}
