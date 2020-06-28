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

// Processor is a abstract interface to the VM processor instance. It can be called via exported entry points
// Each entry point is uniquely identified by the request code (uint16). The request code contains information
// if it requires authentication to run (protected) and also if it represents built in processor or
// user-defined processor.
type Processor interface {
	// returns true if processor can process specific request code. Valid only for not reserved codes
	// to return true for reserved codes is ignored
	GetEntryPoint(code sctransaction.RequestCode) (EntryPoint, bool)
}

// EntryPoint is an abstract interface by which VM is run by passing the Sandbox interface to it
// VM is expected to be fully deterministic and it result is 100% reflected
// as a side effect on the Sandbox interface
type EntryPoint interface {
	WithGasLimit(int) EntryPoint
	WithTimeLimitMillis(int) EntryPoint
	Run(ctx Sandbox)
}
