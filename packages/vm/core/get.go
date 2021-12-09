package core

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/accounts/commonaccount"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/governance/governanceimpl"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/root/rootimpl"
)

var AllCoreContractsByHash = map[hashing.HashValue]*coreutil.ContractProcessor{
	root.Contract.ProgramHash:       rootimpl.Processor,
	accounts.Contract.ProgramHash:   accounts.Processor,
	blob.Contract.ProgramHash:       blob.Processor,
	blocklog.Contract.ProgramHash:   blocklog.Processor,
	governance.Contract.ProgramHash: governanceimpl.Processor,
}

func init() {
	for _, rec := range AllCoreContractsByHash {
		commonaccount.SetCoreHname(rec.Contract.Hname())
	}
}

func GetProcessor(programHash hashing.HashValue) (iscp.VMProcessor, error) {
	ret, ok := AllCoreContractsByHash[programHash]
	if !ok {
		return nil, fmt.Errorf("can't find builtin processor with hash %s", programHash.String())
	}
	return ret, nil
}
