// 'root' a core contract on the chain. It is responsible for:
// - initial setup of the chain during chain deployment
// - maintaining of core parameters of the chain

package rootimpl

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

var Processor = root.Contract.Processor(nil,
	root.ViewFindContract.WithHandler(findContract),
	root.ViewGetContractRecords.WithHandler(getContractRecords),
)

// findContract view finds and returns encoded record of the contract
func findContract(ctx isc.SandboxView, hname isc.Hname) (bool, **root.ContractRecord) {
	state := root.NewStateReaderFromSandbox(ctx)
	rec := state.FindContract(hname)
	if rec == nil {
		return false, nil
	}
	return true, &rec
}

func getContractRecords(ctx isc.SandboxView) []lo.Tuple2[*isc.Hname, *root.ContractRecord] {
	var ret []lo.Tuple2[*isc.Hname, *root.ContractRecord]
	state := root.NewStateReaderFromSandbox(ctx)
	state.GetContractRegistry().Iterate(func(elemKey []byte, value []byte) bool {
		hname := codec.MustDecode[isc.Hname](elemKey)
		rec := lo.Must(root.ContractRecordFromBytes(value))
		ret = append(ret, lo.T2(&hname, rec))
		return true
	})
	return ret
}
