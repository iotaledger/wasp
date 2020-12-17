package examples

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"sync"
)

const VMType = "examplevm"

var (
	allExamples      = make(map[hashing.HashValue]vmtypes.Processor)
	allExamplesMutex = &sync.Mutex{}
)

func AddProcessor(progHash hashing.HashValue, proc vmtypes.Processor) {
	allExamplesMutex.Lock()
	defer allExamplesMutex.Unlock()
	allExamples[progHash] = proc
	fmt.Printf("AddProcessor: added example processor with hash %s\n", progHash.String())
}

func GetExampleProcessor(progHash hashing.HashValue) (vmtypes.Processor, bool) {
	ret, ok := allExamples[progHash]
	if !ok {
		return nil, false
	}
	return ret, true
}
