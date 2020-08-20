package builtins

import "github.com/iotaledger/wasp/packages/vm/examples"

func IsBuiltinProgramHash(progHashStr string) bool {
	_, ok := examples.GetProcessor(progHashStr)
	return ok
}
