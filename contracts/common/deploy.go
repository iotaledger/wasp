package common

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/stretchr/testify/require"
)

const (
	Debug      = false
	StackTrace = false
	TraceHost  = false
)

func DeployWasmContractByName(chain *solo.Chain, scName string, params ...interface{}) error {
	// wasmproc.GoWasmVM = NewWasmTimeJavaVM()
	// wasmproc.GoWasmVM = NewWartVM()
	wasmFile := scName + "_bg.wasm"
	exists, _ := util.ExistsFilePath("../pkg/" + wasmFile)
	if exists {
		wasmFile = "../pkg/" + wasmFile
	}
	return chain.DeployWasmContract(nil, scName, wasmFile, params...)
}

func StartChain(t *testing.T, chainName string) *solo.Chain {
	wasmhost.HostTracing = TraceHost
	env := solo.New(t, Debug, StackTrace)
	return env.NewChain(nil, chainName)
}

func StartChainAndDeployWasmContractByName(t *testing.T, scName string, params ...interface{}) *solo.Chain {
	chain := StartChain(t, "chain1")
	err := DeployWasmContractByName(chain, scName, params...)
	require.NoError(t, err)
	return chain
}
