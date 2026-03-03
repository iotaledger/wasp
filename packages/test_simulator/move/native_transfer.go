package move

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/test_simulator/l1"
)

var TransferHandlers = map[string]l1.MoveCallFunc{
	"public_transfer": transferTransfer,
	"transfer":        transferTransfer,
	"public_receive":  transferReceive,
	"receive":         transferReceive,
}

// transfer::transfer(obj, recipient)
func transferTransfer(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("transfer::transfer requires 2 arguments")
	}
	objID := args[0].ObjectID
	if objID == nil {
		return nil, fmt.Errorf("transfer::transfer: first argument has no object ID")
	}

	recipientAddr, ok := args[1].Raw.(*iotago.Address)
	if !ok {
		if addr, ok2 := args[1].Raw.(iotago.Address); ok2 {
			recipientAddr = &addr
		} else {
			return nil, fmt.Errorf("transfer::transfer: second argument is not an address")
		}
	}

	obj, exists := ctx.Store.Get(*objID)
	if !exists {
		return nil, fmt.Errorf("transfer::transfer: object %s not found", objID.String())
	}
	obj.Owner = l1.SimOwner{AddressOwner: recipientAddr}
	ctx.Store.Put(obj)

	return nil, nil
}

// transfer::receive(parent_uid, receiving) -> object
func transferReceive(ctx *l1.CallContext, _ *iotago.ProgrammableMoveCall, args []l1.Value) ([]l1.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("transfer::receive requires 2 arguments")
	}

	receivingID := args[1].ObjectID
	if receivingID == nil {
		return nil, fmt.Errorf("transfer::receive: receiving argument has no object ID")
	}

	obj, exists := ctx.Store.Get(*receivingID)
	if !exists {
		return nil, fmt.Errorf("transfer::receive: receiving object %s not found", receivingID.String())
	}

	return []l1.Value{{ObjectID: receivingID, Raw: obj, Type: obj.Type}}, nil
}
