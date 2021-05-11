package common

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	Debug      = false
	StackTrace = false
	TraceHost  = false
)

func StartChainAndDeployWasmContractByName(t *testing.T, scName string) *solo.Chain {
	wasmhost.HostTracing = TraceHost
	env := solo.New(t, Debug, StackTrace)
	chain := env.NewChain(nil, "chain1")
	wasmFile := scName + "_bg.wasm"
	exists, _ := util.ExistsFilePath("../pkg/" + wasmFile)
	if exists {
		wasmFile = "../pkg/" + wasmFile
	}
	err := chain.DeployWasmContract(nil, scName, wasmFile)
	require.NoError(t, err)
	return chain
}
