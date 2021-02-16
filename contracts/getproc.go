package contracts

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"sync"
)

const VMType = "examplevm"

var (
	allExamples      = make(map[hashing.HashValue]coretypes.Processor)
	allExamplesMutex = &sync.Mutex{}
)

// AddExampleProcessor adds new processor to the runtime registry of example processors.
// The 'proc' represents executable of the specific smart contract.
// It must implement coretypes.Processor
func AddExampleProcessor(c *coreutil.ContractInterface) {
	allExamplesMutex.Lock()
	defer allExamplesMutex.Unlock()
	allExamples[c.ProgramHash] = c
	fmt.Printf("----- AddExampleProcessor: name: '%s', program hash: %s, description: '%s'\n",
		c.Name, c.ProgramHash.String(), c.Description)
}

// GetExampleProcessor retrieves smart contract processor (VM) by the hash (with existence flag)
func GetExampleProcessor(progHash hashing.HashValue) (coretypes.Processor, bool) {
	ret, ok := allExamples[progHash]
	if !ok {
		return nil, false
	}
	return ret, true
}
