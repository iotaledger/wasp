package txutil

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"
	"sort"
)

func MintedColorFromOutput(out ledgerstate.Output) ledgerstate.Color {
	return blake2b.Sum256(out.ID().Bytes())
}

// NewOutputs returns sorted outputs and permutation to keep track of original indices
// TODO not correct with duplicates
func NewOutputs(outputs ...ledgerstate.Output) (ledgerstate.Outputs, []uint16, error) {
	// using same sorting as in ledgerstate.NewOutputs
	type sortable struct {
		bytes []byte
		index uint16
	}
	arr := make([]sortable, len(outputs))
	for i, out := range outputs {
		arr[i].bytes = out.Bytes()
		arr[i].index = uint16(i)
	}
	sort.Slice(arr, func(i, j int) bool {
		return bytes.Compare(arr[i].bytes, arr[j].bytes) < 0
	})
	retPermutation := make([]uint16, len(outputs))
	for i := range retPermutation {
		retPermutation[i] = arr[i].index
	}
	ret := ledgerstate.NewOutputs(outputs...)
	if len(ret) != len(retPermutation) {
		return nil, nil, xerrors.New("duplicates (identical outputs) not allowed")
	}
	return ret, retPermutation, nil
}
