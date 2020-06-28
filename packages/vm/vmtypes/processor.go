// package present processor interface. It must be implemented by VM
package vmtypes

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	flag "github.com/spf13/pflag"
)

const (
	CfgVMBinaryDir   = "vm.binaries"
	CfgDefaultVmType = "vm.defaultvm"
)

func init() {
	flag.String(CfgVMBinaryDir, "wasm", "path where Wasm binaries are located (using file:// schema")
	flag.String(CfgDefaultVmType, "dummmy", "default VM type")
}

type Processor interface {
	// returns true if processor can process specific request code
	// valid only for not reserved codes
	// to return true for reserved codes is ignored
	GetEntryPoint(code sctransaction.RequestCode) (EntryPoint, bool)
}

type EntryPoint interface {
	Run(ctx Sandbox)
}
