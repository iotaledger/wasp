package txstream

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
)

func init() {
	panic("THIS IS A PLACEHOLDER PACKAGE")
}

type (
	Client struct {
		Events
	}

	Event struct {
		Attach func(*events.Closure)
	}
	Events struct {
		TransactionReceived        Event
		InclusionStateReceived     Event
		OutputReceived             Event
		UnspentAliasOutputReceived Event
	}

	MsgTransaction        interface{}
	MsgTxInclusionState   interface{}
	MsgOutput             interface{}
	MsgUnspentAliasOutput interface{}
)

func NewClient() *Client {
	panic("not implemented")
}

// RequestBacklog requests the backlog for a given address.
func (n *Client) RequestBacklog(addr iotago.Address) {
	panic("not implemented")
}

// RequestConfirmedTransaction requests a specific confirmed transaction.
func (n *Client) RequestConfirmedTransaction(addr iotago.Address, txid iotago.TransactionID) {
	panic("not implemented")
}

// RequestTxInclusionState requests the inclusion state of a transaction.
func (n *Client) RequestTxInclusionState(addr iotago.Address, txid iotago.TransactionID) {
	panic("not implemented")
}

// RequestConfirmedOutput requests a specific confirmed output.
func (n *Client) RequestConfirmedOutput(addr iotago.Address, outputID *iotago.OutputID) {
	panic("not implemented")
}

// RequestUnspentAliasOutput requests the unique unspent alias output for the given AliasAddress.
func (n *Client) RequestUnspentAliasOutput(addr *iotago.AliasAddress) {
	panic("not implemented")
}

// PostTransaction posts a transaction to the ledger.
func (n *Client) PostTransaction(tx *iotago.Transaction) {
	panic("not implemented")
}

// Subscribe subscribes to real-time updates for the given address.
func (n *Client) Subscribe(addr iotago.Address) {
	panic("not implemented")
}

// ---------------

type UtxoDBLedger struct{}

func New(log *logger.Logger) *UtxoDBLedger {
	panic("not implemented")
}

// --------
func ServerListen(ledger *UtxoDBLedger, bindAddress string, log *logger.Logger, shutdownSignal <-chan struct{}) error {
	panic("not implemented")
}
