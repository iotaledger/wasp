package tests

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/stretchr/testify/require"
)

var magicContractAddress = common.HexToAddress("1074")

func TestEVMChainHalt(t *testing.T) {
	e := newClusterTestEnv(t, []int{0}, waspClusterOpts{
		nNodes: 1,
	})

	// deposit funds to some evm account
	ethKey, _ := e.newEthereumAccountWithL2Funds()

	someL1Address := tpkg.RandEd25519Address()
	balance := e.cluster.L1BaseTokens(someL1Address)

	iscABI, err := abi.JSON(strings.NewReader(iscmagic.ABI))
	require.NoError(t, err)

	wdTokens := 1 * isc.Million
	callArguments, err := iscABI.Pack("send",
		iscmagic.WrapL1Address(someL1Address),
		iscmagic.WrapISCFungibleTokens(
			*isc.NewFungibleBaseTokens(wdTokens),
		),
		false,
		iscmagic.WrapISCSendMetadata(isc.SendMetadata{}),
		iscmagic.WrapISCSendOptions(isc.SendOptions{}),
	)
	require.NoError(t, err)

	// try to withdraw multiple times via magic contract
	for i := 0; i < 10; i++ {
		tx, err := types.SignTx(
			types.NewTransaction(
				uint64(i),
				magicContractAddress,
				common.Big0,
				1_000_000,
				evm.GasPrice,
				callArguments,
			),
			e.Signer(),
			ethKey,
		)
		require.NoError(t, err)

		_, err = e.SendTransactionAndWait(tx)
		require.NoError(t, err)

		newBalance := e.cluster.L1BaseTokens(someL1Address)
		require.Equal(t, balance+wdTokens, newBalance)
		balance = newBalance
	}
}
