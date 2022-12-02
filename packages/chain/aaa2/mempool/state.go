package mempool

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
)

// TODO these functions need to be re-implemented for the new immutable state

type (
	HasBeenProcessedFunc     func(reqID isc.RequestID) bool
	GetProcessedRequestsFunc func(from, to *isc.AliasOutputWithID) []isc.RequestID
)

func CreateHasBeenProcessedFunc(kvr kv.KVStoreReader) HasBeenProcessedFunc {
	return func(reqID isc.RequestID) (hasBeenProcessed bool) {
		panic("TODO re-implement")
		// // this will need to be refectored for the new immutable state implementation
		// var err error
		// err = optimism.RetryOnStateInvalidated(
		// 	func() error {
		// 		hasBeenProcessed, err = blocklog.IsRequestProcessed(kvr, &reqID)
		// 		return err
		// 	},
		// )
		// if err != nil {
		// 	// its okay to panic here, this will be removed with the new state implementation
		// 	panic(fmt.Sprintf("error while checking hasBeenProcessed, reqID:%s, err:%s", reqID, err.Error()))
		// }
		// return hasBeenProcessed
	}
}

func CreateGetProcessedReqsFunc(kvr kv.KVStoreReader) GetProcessedRequestsFunc {
	return func(from, to *isc.AliasOutputWithID) []isc.RequestID {
		panic("TODO re-implement")
		// this will need to be refectored for the new immutable state implementation
		// ret := []isc.RequestID{}

		// fromIdx := uint32(0)
		// if from != nil {
		// 	fromIdx = from.GetStateIndex()
		// }
		// toIdx := to.GetStateIndex()
		// // from+1 ~ to
		// blockIndexes := make([]uint32, toIdx-fromIdx)
		// for i := range blockIndexes {
		// 	blockIndexes[i] = fromIdx + uint32(i+1)
		// }

		// for _, idx := range blockIndexes {
		// 	stateReadErr := optimism.RetryOnStateInvalidated(
		// 		func() error {
		// 			reqIds, err := blocklog.GetRequestIDsForBlock(kvr, idx)
		// 			ret = append(ret, reqIds...)
		// 			return err
		// 		},
		// 	)
		// 	if stateReadErr != nil {
		// 		// its okay to panic here, this will be removed with the new state implementation
		// 		panic(fmt.Sprintf("error while checking GetRequestIDsForBlock,  err:%s", stateReadErr.Error()))
		// 	}
		// }
		// return ret
	}
}
