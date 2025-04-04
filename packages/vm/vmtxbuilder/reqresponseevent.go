package vmtxbuilder

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm"
)

func CreateRequestResponseEvents(results []*vm.RequestResult, resolver func(e *isc.UnresolvedVMError) string) map[iotago.ObjectID]iscmoveclient.RequestResultEvent {
	requestResultEvents := map[iotago.ObjectID]iscmoveclient.RequestResultEvent{}

	for _, v := range results {
		id := v.Request.ID()
		var receiptError string = ""

		if v.Receipt.Error != nil {
			receiptError = resolver(v.Receipt.Error)
		}

		requestResultEvents[iotago.ObjectID(id)] = iscmoveclient.RequestResultEvent{
			RequestID: id.AsIotaObjectID(),
			Error:     receiptError,
		}
	}

	return requestResultEvents
}
