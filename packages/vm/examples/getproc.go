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

//
//var allExamples = map[string]vmtypes.Processor{
//	vmnil.ProgramHash:         vmnil.GetProcessor(),
//	logsc.ProgramHash:         logsc.GetProcessor(),
//	inccounter.ProgramHashStr: inccounter.GetProcessor(),
//	// TODO
//	//fairroulette.ProgramHashStr:  fairroulette.GetProcessor(),
//	//fairauction.ProgramHashStr:   fairauction.GetProcessor(),
//	//tokenregistry.ProgramHashStr: tokenregistry.GetProcessor(),
//	//dwfimpl.ProgramHashStr:       dwfimpl.GetProcessor(),
//}

func GetExampleProcessor(progHash hashing.HashValue) (vmtypes.Processor, bool) {
	ret, ok := allExamples[progHash]
	if !ok {
		return nil, false
	}
	return ret, true
}
