package core

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

const (
	VMType = "builtinvm"
)

func GetProcessor(programHash hashing.HashValue) (coretypes.Processor, error) {
	switch programHash {
	case root.Interface.ProgramHash:
		return root.Interface, nil

	case accounts.Interface.ProgramHash:
		return accounts.Interface, nil

	case blob.Interface.ProgramHash:
		return blob.Interface, nil

	case eventlog.Interface.ProgramHash:
		return eventlog.Interface, nil
	}
	return nil, fmt.Errorf("can't find builtin processor with hash %s", programHash.String())
}
