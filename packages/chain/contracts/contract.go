package contracts

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type Contract struct {
	index uint16
	vm    vmtypes.Processor
}

type Contracts []*Contract

func NewContracts(n uint16) Contracts {
	return make(Contracts, n)
}

func (cs Contracts) Size() uint16 {
	return (uint16)(len(cs))
}

func (cs Contracts) LoadContract(binaryCode []byte, vmtype string, index uint16) error {
	if index >= cs.Size() {
		return fmt.Errorf("Contracts.LoadContract: wrong contract index %d", index)
	}
	if cs[index] != nil {
		return fmt.Errorf("Contracts.LoadContract: contract with index %d already loaded", index)
	}
	proc, err := processors.FromBinaryCode(vmtype, binaryCode)
	if err != nil {
		return err
	}
	cs[index].vm = proc
	return nil
}
