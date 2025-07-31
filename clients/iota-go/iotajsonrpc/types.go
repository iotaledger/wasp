package iotajsonrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

var IotaCoinType CoinType = CoinType(iotago.MustNewResourceType("0x2::iota::IOTA").String())

// ShortString Returns the address with leading zeros trimmed, e.g. 0x2

type InputObjectKind map[string]interface{}

type TransactionBytes struct {
	// the gas object to be used
	Gas []iotago.ObjectRef `json:"gas"`

	// objects to be used in this transaction
	InputObjects []InputObjectKind `json:"inputObjects"`

	// transaction data bytes
	TxBytes iotago.Base64Data `json:"txBytes"`
}

type TransferObject struct {
	Recipient iotago.Address   `json:"recipient"`
	ObjectRef iotago.ObjectRef `json:"object_ref"`
}
type ModulePublish struct {
	Modules [][]byte `json:"modules"`
}
type MoveCall struct {
	Package  iotago.ObjectID `json:"package"`
	Module   string          `json:"module"`
	Function string          `json:"function"`
	TypeArgs []interface{}   `json:"typeArguments"`
	Args     []interface{}   `json:"arguments"`
}
type TransferIota struct {
	Recipient iotago.Address `json:"recipient"`
	Amount    uint64         `json:"amount"`
}
type Pay struct {
	Coins      []iotago.ObjectRef `json:"coins"`
	Recipients []iotago.Address   `json:"recipients"`
	Amounts    []uint64           `json:"amounts"`
}
type PayIota struct {
	Coins      []iotago.ObjectRef `json:"coins"`
	Recipients []iotago.Address   `json:"recipients"`
	Amounts    []uint64           `json:"amounts"`
}
type PayAllIota struct {
	Coins     []iotago.ObjectRef `json:"coins"`
	Recipient iotago.Address     `json:"recipient"`
}
type ChangeEpoch struct {
	Epoch             interface{} `json:"epoch"`
	StorageCharge     uint64      `json:"storage_charge"`
	ComputationCharge uint64      `json:"computation_charge"`
}

type SingleTransactionKind struct {
	TransferObject *TransferObject `json:"TransferObject,omitempty"`
	Publish        *ModulePublish  `json:"Publish,omitempty"`
	Call           *MoveCall       `json:"Call,omitempty"`
	TransferIota   *TransferIota   `json:"TransferIota,omitempty"`
	ChangeEpoch    *ChangeEpoch    `json:"ChangeEpoch,omitempty"`
	PayIota        *PayIota        `json:"PayIota,omitempty"`
	Pay            *Pay            `json:"Pay,omitempty"`
	PayAllIota     *PayAllIota     `json:"PayAllIota,omitempty"`
}

type SenderSignedData struct {
	Transactions []SingleTransactionKind `json:"transactions,omitempty"`

	Sender     *iotago.Address   `json:"sender"`
	GasPayment *iotago.ObjectRef `json:"gasPayment"`
	GasBudget  uint64            `json:"gasBudget"`
	// GasPrice     uint64      `json:"gasPrice"`
}

type TimeRange struct {
	StartTime uint64 `json:"startTime"` // left endpoint of time interval, milliseconds since epoch, inclusive
	EndTime   uint64 `json:"endTime"`   // right endpoint of time interval, milliseconds since epoch, exclusive
}

type MoveModule struct {
	Package iotago.ObjectID   `json:"package"`
	Module  iotago.Identifier `json:"module"`
}

func (o ObjectOwner) MarshalJSON() ([]byte, error) {
	if o.string != nil {
		data, err := json.Marshal(o.string)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	if o.ObjectOwnerInternal != nil {
		data, err := json.Marshal(o.ObjectOwnerInternal)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, errors.New("nil value")
}

func (o *ObjectOwner) UnmarshalJSON(data []byte) error {
	if bytes.HasPrefix(data, []byte("\"")) {
		stringData := string(data[1 : len(data)-1])
		o.string = &stringData
		return nil
	}
	if bytes.HasPrefix(data, []byte("{")) {
		oOI := ObjectOwnerInternal{}
		err := json.Unmarshal(data, &oOI)
		if err != nil {
			return err
		}
		o.ObjectOwnerInternal = &oOI
		return nil
	}
	return errors.New("value not json")
}

func IsSameAddressString(addr1, addr2 string) bool {
	addr1 = strings.TrimPrefix(addr1, "0x")
	addr2 = strings.TrimPrefix(addr2, "0x")
	return strings.TrimLeft(addr1, "0") == strings.TrimLeft(addr2, "0")
}
