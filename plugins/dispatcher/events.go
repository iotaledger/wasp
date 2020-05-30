package dispatcher

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type dispatcherEvents struct {
	TransactionArrivedFromNode   *events.Event
	OutputsArrivedFromNode       *events.Event
	AddressUpdateArrivedFromNode *events.Event
}

var Events = dispatcherEvents{
	TransactionArrivedFromNode:   events.NewEvent(scTransactionCaller),
	OutputsArrivedFromNode:       events.NewEvent(addressOutputsCaller),
	AddressUpdateArrivedFromNode: events.NewEvent(addressUpdateCaller),
}

func scTransactionCaller(handler interface{}, params ...interface{}) {
	handler.(func(transaction *sctransaction.Transaction))(params[0].(*sctransaction.Transaction))
}

func addressOutputsCaller(handler interface{}, params ...interface{}) {
	handler.(func(addr address.Address, balances map[valuetransaction.ID][]*balance.Balance))(
		params[0].(address.Address),
		params[1].(map[valuetransaction.ID][]*balance.Balance),
	)
}

func addressUpdateCaller(handler interface{}, params ...interface{}) {
	handler.(func(addr address.Address, balances map[valuetransaction.ID][]*balance.Balance, tx *sctransaction.Transaction))(
		params[0].(address.Address),
		params[1].(map[valuetransaction.ID][]*balance.Balance),
		params[2].(*sctransaction.Transaction),
	)
}
