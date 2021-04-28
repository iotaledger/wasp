package core

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core/_default"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

const (
	VMType = "builtinvm"
)

var AllCoreContracts = map[hashing.HashValue]*coreutil.ContractInterface{
	_default.Interface.ProgramHash: _default.Interface,
	root.Interface.ProgramHash:     root.Interface,
	accounts.Interface.ProgramHash: accounts.Interface,
	blob.Interface.ProgramHash:     blob.Interface,
	eventlog.Interface.ProgramHash: eventlog.Interface,
}

func GetProcessor(programHash hashing.HashValue) (coretypes.VMProcessor, error) {
	ret, ok := AllCoreContracts[programHash]
	if !ok {
		return nil, fmt.Errorf("can't find builtin processor with hash %s", programHash.String())
	}
	return ret, nil
}
