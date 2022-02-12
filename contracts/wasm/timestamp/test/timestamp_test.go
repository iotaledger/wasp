package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/timestamp/go/timestamp"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, timestamp.ScName, timestamp.OnLoad)
	require.NoError(t, ctx.ContractExists(timestamp.ScName))
}
