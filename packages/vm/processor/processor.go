// package present processor interface. It must be implemented by VM
package processor

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
)

type Processor interface {
	// returns true if processor can process specific request code
	// valid only for not reserved codes
	// to return true for reserved codes is ignored
	// the best way to implement is with meta-data next to the Wasm binary
	GetEntryPoint(code sctransaction.RequestCode) (EntryPoint, bool)
}

type EntryPoint interface {
	Run(ctx sandbox.Sandbox)
}
