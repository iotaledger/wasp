package testapilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"net/http"
	"strconv"
)

type RequestBlockJson struct {
	Address string            `json:"address"`
	Amount  int64             `json:"amount"`
	Vars    map[string]string `json:"vars"`
}

type RequestTransactionJson struct {
	RequestorIndex int                `json:"requestor_index"` // from 1-10, corresponds to one of hardcodded addresses. 0 means 1
	Requests       []RequestBlockJson `json:"requests"`        // can't be empty
}

func requestBlockFromJson(reqBlkJson *RequestBlockJson) (*sctransaction.RequestBlock, error) {
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

func TransactionFromJsonTesting(txJson *RequestTransactionJson) (*sctransaction.Transaction, error) {
	var err error
	requestorIndex := txJson.RequestorIndex
	if requestorIndex <= 0 {
		requestorIndex = 1
	}
	if requestorIndex > 10 {
		return nil, errors.New("TransactionFromJsonTesting: wrong input params")
	}
	if len(txJson.Requests) == 0 {
		return nil, errors.New("TransactionFromJsonTesting: wrong input params")
	}
	totalAmount := int64(0)
	for _, r := range txJson.Requests {
		a := r.Amount
		if a <= 0 {
			a = 1
		}
		totalAmount += a
	}
	inputAddr := utxodb.GetAddress(requestorIndex)
	balances, err := GetBalancesFromNodeSync(inputAddr)
	if err != nil {
		return nil, err
	}
	allOuts := waspconn.BalancesToOutputs(inputAddr, balances)

	selectedOutputs := util.SelectMinimumOutputs(allOuts, balance.ColorIOTA, totalAmount)
	if len(selectedOutputs) == 0 {
		return nil, errors.New("not enough funds")
	}
	oids := make([]valuetransaction.OutputID, 0, len(selectedOutputs))
	for oid := range selectedOutputs {
		oids = append(oids, oid)
	}

	txb := sctransaction.NewTransactionBuilder()
	txb.AddInputs(oids...)

	for _, reqBlkJson := range txJson.Requests {
		reqBlk, err := requestBlockFromJson(&reqBlkJson)
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
	totalInputs := apilib.TotalBalanceOfInputs(selectedOutputs)
	if totalInputs > totalAmount {
		// add reminder
		txb.AddBalanceToOutput(inputAddr, balance.New(balance.ColorIOTA, totalInputs-totalAmount))
	}
	tx, err := txb.Finalize()
	if err != nil {
		return nil, err
	}
	tx.Sign(utxodb.GetSigScheme(inputAddr))

	MustNotNullInputs(tx.Transaction)

	// imposible, because imputs are on the server
	//if err = utxodb.CheckInputsOutputs(tx.Transaction); err != nil{
	//	return nil, err
	//}

	return tx, nil
}

func SendTestRequest(netLoc string, txJson *RequestTransactionJson) error {
	data, err := json.Marshal(txJson)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/client/testreq", netLoc)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	result := &misc.SimpleResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Error != "" {
		return errors.New(result.Error)
	}
	return nil
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
