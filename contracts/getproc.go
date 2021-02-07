package contracts

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"sync"
)

const VMType = "examplevm"

var (
	allExamples      = make(map[hashing.HashValue]coretypes.Processor)
	allExamplesMutex = &sync.Mutex{}
)

// AddProcessor adds new processor to the runtime registry of example processors.
// The 'proc' represents executable of the specific smart contract.
// It must implement vmtypes.Processor
func AddProcessor(c *contract.ContractInterface) {
	allExamplesMutex.Lock()
	defer allExamplesMutex.Unlock()
	allExamples[c.ProgramHash] = c
	fmt.Printf("added example processor name: '%s', program hash: %s, dscr: '%s'\n",
		c.Name, c.ProgramHash.String(), c.Description)
}

// GetExampleProcessor retrieves smart contract processor (VM) by the hash (whith existence flag)
func GetExampleProcessor(progHash hashing.HashValue) (coretypes.Processor, bool) {
	ret, ok := allExamples[progHash]
	if !ok {
		return nil, false
	}
	return ret, true
}
