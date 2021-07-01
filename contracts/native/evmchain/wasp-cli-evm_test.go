package evmchain

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
)

func TestWaspCLIEVMContract(t *testing.T) {
	w := testutil.NewWaspCLITest(t)
	w.Run("init")
	w.Run("request-funds")
	committee, quorum := w.CommitteeConfig()
	w.Run("chain", "deploy", "--chain=chain1", committee, quorum)

	vmtype := vmtypes.Native
	name := Interface.Name

	faucetKey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	faucetAddress := crypto.PubkeyToAddress(faucetKey.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	// test that the evmchain contract can be deployed using wasp-cli
	w.Run("chain", "deploy-contract", vmtype, name, Interface.Description, Interface.ProgramHash.Base58(),
		"string", FieldGenesisAlloc, "bytes", base58.Encode(EncodeGenesisAlloc(genesisAlloc)),
	)

	out := w.Run("chain", "list-contracts")
	found := false
	for _, s := range out {
		if strings.Contains(s, name) {
			found = true
			break
		}
	}
	require.True(t, found)

	// deploy the EVM contract `storage`
	{
		abiJSON := evmtest.StorageContractABI
		bytecode := evmtest.StorageContractBytecode

		args := []interface{}{uint32(42)}

		nonce := uint64(0)

		contractABI, err := abi.JSON(strings.NewReader(abiJSON))
		require.NoError(t, err)
		constructorArguments, err := contractABI.Pack("", args...)
		require.NoError(t, err)
		data := []byte{}
		data = append(data, bytecode...)
		data = append(data, constructorArguments...)

		value := big.NewInt(0)

		gas := uint64(1000)

		tx, err := types.SignTx(
			types.NewContractCreation(nonce, value, gas, evm.GasPrice, data),
			evm.Signer(),
			faucetKey,
		)
		require.NoError(t, err)

		txdata, err := tx.MarshalBinary()
		require.NoError(t, err)

		w.Run("chain", "post-request", name, "increment", "--transfer=IOTA:1000", "--off-ledger",
			FieldTransactionData, base58.Encode(txdata),
		)

		checkCounter := func(n int) {
			// test chain call-view command
			out = w.Run("chain", "call-view", name, "getCounter")
			out = w.Pipe(out, "decode", "string", "counter", "int")
			require.Regexp(t, fmt.Sprintf("(?m)counter:[[:space:]]+%d$", n), out[0])
		}
		checkCounter(43)
	}
}
