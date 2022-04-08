package tests

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/stretchr/testify/require"
)

func TestWaspCLIEVMDeploy(t *testing.T) {
	w := newWaspCLITest(t)
	w.Run("init")
	w.Run("request-funds")
	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)
	// for off-ledger requests
	w.Run("chain", "deposit", "IOTA:2000")

	faucetKey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	faucetAddress := crypto.PubkeyToAddress(faucetKey.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))

	// test that the EVM chain can be deployed using wasp-cli
	w.Run("chain", "evm", "deploy",
		"--alloc", fmt.Sprintf("%s:%s", faucetAddress.String(), faucetSupply.String()),
		"--evm-flavor", evm.Contract.Name,
	)

	out := w.Run("chain", "list-contracts")
	found := false
	for _, s := range out {
		if strings.Contains(s, evm.Contract.Name) {
			found = true
			break
		}
	}
	require.True(t, found)
}
