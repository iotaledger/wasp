package sbtests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestTypesFull(t *testing.T) { run2(t, testTypesFull) }
func testTypesFull(t *testing.T) {
	_, ch := setupChain(t, nil)
	cID := setupTestSandboxSC(t, ch, nil)

	ch.MustDepositBaseTokensToL2(10_000, nil)

	req := solo.NewCallParamsEx(ScName, sbtestsc.FuncPassTypesFull.Name, isc.NewCallArguments(
		codec.Encode("string"),
		codec.Encode(42),
		codec.Encode(0),
		codec.Encode(hashing.HashStrings("Hash")),
		codec.Encode(isc.Hn("Hname")),
		codec.Encode(isc.Hname(0)),
		codec.Encode(cID),
		codec.Encode(ch.ChainID),
		codec.Encode(ch.ChainID.AsAddress()),
		codec.Encode(ch.OriginatorAgentID),
	)).WithGasBudget(150_000)
	_, err := ch.PostRequestSync(req, nil)
	require.NoError(t, err)
}

func TestTypesView(t *testing.T) { run2(t, testTypesView) }
func testTypesView(t *testing.T) {
	_, chain := setupChain(t, nil)
	cID := setupTestSandboxSC(t, chain, nil)

	_, err := chain.CallViewEx(ScName, sbtestsc.FuncPassTypesView.Name, isc.NewCallArguments(
		codec.Encode("string"),
		codec.Encode(42),
		codec.Encode(0),
		codec.Encode(hashing.HashStrings("Hash")),
		codec.Encode(isc.Hn("Hname")),
		codec.Encode(isc.Hname(0)),
		codec.Encode(cID),
		codec.Encode(chain.ChainID),
		codec.Encode(chain.ChainID.AsAddress()),
		codec.Encode(chain.OriginatorAgentID),
	))

	require.NoError(t, err)
}
