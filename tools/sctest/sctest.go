package sctest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/util"
)

func NewSCTestBed() *SCTestBed {
	return &SCTestBed{
		utxodb: utxodb.New(),
	}
}

type addressData struct {
	name      string
	sigScheme signaturescheme.SignatureScheme
}

type SCTestBed struct {
	utxodb    *utxodb.UtxoDB
	addresses map[address.Address]*addressData
}

func (tb *SCTestBed) RequestFunds(targetAddress address.Address) error {
	_, err := tb.utxodb.RequestFunds(targetAddress)
	return err
}

func (tb *SCTestBed) GetAccountOutputs(addr *address.Address) (map[valuetransaction.OutputID][]*balance.Balance, error) {
	return tb.utxodb.GetAddressOutputs(*addr), nil
}

func (tb *SCTestBed) PostAndWaitForConfirmation(tx *valuetransaction.Transaction) error {
	return tb.utxodb.AddTransaction(tx)
}

// new EdDSA address
func (tb *SCTestBed) NewAddress(name string, requestFunds bool) (*address.Address, error) {
	sigScheme := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	addr := sigScheme.Address()
	tb.addresses[addr] = &addressData{
		name:      name,
		sigScheme: sigScheme,
	}
	if !requestFunds {
		return &addr, nil
	}
	if err := tb.RequestFunds(addr); err != nil {
		return nil, err
	}
	return &addr, nil
}

func (tb *SCTestBed) SigScheme(addr *address.Address) (signaturescheme.SignatureScheme, bool) {
	ret, ok := tb.addresses[*addr]
	return ret.sigScheme, ok
}

type validationResult struct {
	expected int64
	actual   int64
	valid    string // ok, fail, -
}

func (tb *SCTestBed) ValidateBalances(addr *address.Address, expected map[balance.Color]int64) (map[balance.Color]*validationResult, bool, error) {
	allOuts, err := tb.GetAccountOutputs(addr)
	if err != nil {
		return nil, false, err
	}
	ret := make(map[balance.Color]*validationResult)

	for col, exp := range expected {
		ret[col] = &validationResult{
			expected: exp,
			valid:    "fail",
		}
	}
	byColor, _ := util.OutputBalancesByColor(allOuts)
	allOk := true
	for col, act := range byColor {
		vr, ok := ret[col]
		if ok {
			vr.actual = act
			if vr.actual == vr.expected {
				vr.valid = "ok"
			} else {
				allOk = false
			}
		} else {
			ret[col] = &validationResult{
				expected: 0,
				actual:   act,
				valid:    "-",
			}
		}
	}
	return ret, allOk, nil
}
