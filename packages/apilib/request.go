package apilib

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	nodeapi "github.com/iotaledger/goshimmer/packages/waspconn/apilib"
	"github.com/iotaledger/wasp/packages/util"
	"strconv"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type RequestBlockJson struct {
	Address     string            `json:"address"`
	RequestCode uint16            `json:"request_code"`
	Amount      int64             `json:"amount"` // minimum 1i
	Vars        map[string]string `json:"vars"`
}

func CreateRequestTransaction(node string, senderSigScheme signaturescheme.SignatureScheme, reqsJson []*RequestBlockJson) (*sctransaction.Transaction, error) {
	var err error
	if len(reqsJson) == 0 {
		return nil, errors.New("CreateRequestTransaction: must be at least 1 request block")
	}
	totalAmount := int64(0)
	for _, r := range reqsJson {
		a := r.Amount
		if a <= 0 {
			a = 1
		}
		totalAmount += a
	}
	senderAddr := senderSigScheme.Address()
	allOuts, err := nodeapi.GetAccountOutputs(node, &senderAddr)
	if err != nil {
		return nil, fmt.Errorf("can't get outputs from the node: %v", err)
	}

	selectedOutputs := util.SelectOutputsForAmount(allOuts, balance.ColorIOTA, totalAmount)
	if len(selectedOutputs) == 0 {
		return nil, errors.New("not enough funds")
	}
	oids := make([]valuetransaction.OutputID, 0, len(selectedOutputs))
	for oid := range selectedOutputs {
		oids = append(oids, oid)
	}

	txb := sctransaction.NewTransactionBuilder()
	if err := txb.AddInputs(oids...); err != nil {
		return nil, err
	}

	for _, reqBlkJson := range reqsJson {
		reqBlk, err := requestBlockFromJson(reqBlkJson)
		if err != nil {
			return nil, err
		}
		// add request block
		txb.AddRequestBlock(reqBlk)
		// add new request token to outputs
		a := reqBlkJson.Amount
		if a <= 0 {
			a = 1
		}
		txb.AddBalanceToOutput(reqBlk.Address(), balance.New(balance.ColorNew, 1))
		if a > 1 {
			txb.AddBalanceToOutput(reqBlk.Address(), balance.New(balance.ColorIOTA, a-1))
		}
	}
	totalInputs := TotalBalanceOfInputs(selectedOutputs)
	if totalInputs > totalAmount {
		// add reminder
		txb.AddBalanceToOutput(senderAddr, balance.New(balance.ColorIOTA, totalInputs-totalAmount))
	}
	tx, err := txb.Finalize()
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

	for k, v := range reqBlkJson.Vars {
		n, err := strconv.Atoi(v)
		if err != nil {
			ret.Args().SetString(k, v)
		} else {
			ret.Args().SetInt64(k, int64(n))
		}
	}
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
