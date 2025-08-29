// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package emulator

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"math"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/evm/evmtest"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/dict"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestDecodeHeader(t *testing.T) {
	h1 := &header{
		Hash:        common.BytesToHash(hashing.PseudoRandomHash(nil).Bytes()),
		GasLimit:    123456789,
		GasUsed:     54321,
		Time:        uint64(time.Now().UnixNano()),
		TxHash:      common.BytesToHash(hashing.PseudoRandomHash(nil).Bytes()),
		ReceiptHash: common.BytesToHash(hashing.PseudoRandomHash(nil).Bytes()),
		Bloom:       [256]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	b1 := h1.Bytes()
	h2 := mustHeaderFromBytes(b1)
	require.EqualValues(t, h1, h2)
	b2 := h2.Bytes()
	require.EqualValues(t, len(b1), len(b2))
	require.EqualValues(t, b1, b2)
	h3 := mustHeaderFromBytes(b2)
	require.EqualValues(t, h2, h3)

	// old format, gob encoded
	b3, err := hex.DecodeString("6bff8103010109686561646572476f6201ff8200010701044861736801ff840001084761734c696d6974010600010747617355736564010600010454696d65010600010654784861736801ff8400010b526563656970744861736801ff84000105426c6f6f6d01ff8600000014ff83010101044861736801ff840001060140000017ff8501010105426c6f6f6d01ff8600010601fe02000000fe018bff8201204dffddff8bff8f51ffbbffa218427effd2110a64ffb9ff96ffd1171bffe37f5923ffe53d26ffd216ffbd5a26ff8d01fc3b9aca0001fea56401fc64243bc8012000000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000000001fe01000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	h4 := mustHeaderFromBytes(b3)
	b4 := h4.Bytes()
	require.NotEqualValues(t, len(b3), len(b4))
	require.NotEqualValues(t, b3, b4)
	h5 := mustHeaderFromBytes(b4)
	require.EqualValues(t, h4, h5)
}

var gasLimits = GasLimits{
	Block: gas.EVMBlockGasLimit(gas.LimitsDefault, &util.Ratio32{A: 1, B: 1}),
	Call:  gas.EVMCallGasLimit(gas.LimitsDefault, &util.Ratio32{A: 1, B: 1}),
}

var gasPrice = big.NewInt(0) // ignored in the emulator

func estimateGas(callMsg ethereum.CallMsg, e *EVMEmulator) (uint64, error) {
	lo := params.TxGas
	hi := gasLimits.Call
	lastOk := uint64(0)
	var lastErr error
	for hi >= lo {
		callMsg.Gas = (lo + hi) / 2
		res, err := e.CallContract(callMsg, true)
		if err != nil {
			return 0, fmt.Errorf("CallContract failed: %w", err)
		}
		if res.Err != nil {
			lastErr = res.Err
			lo = callMsg.Gas + 1
		} else {
			lastOk = callMsg.Gas
			hi = callMsg.Gas - 1
		}
	}
	if lastOk == 0 {
		if lastErr != nil {
			return 0, fmt.Errorf("estimateGas failed: %w", lastErr)
		}
		return 0, errors.New("estimateGas failed")
	}
	return lastOk, nil
}

func sendTransaction(
	t testing.TB,
	emu *EVMEmulator,
	sender *ecdsa.PrivateKey,
	receiverAddress common.Address,
	amount *big.Int,
	data []byte,
	gasLimit uint64,
	mintBlock bool,
) *types.Receipt {
	senderAddress := crypto.PubkeyToAddress(sender.PublicKey)

	if gasLimit == 0 {
		var err error
		gasLimit, err = estimateGas(ethereum.CallMsg{
			From:  senderAddress,
			To:    &receiverAddress,
			Value: amount,
			Data:  data,
		}, emu)
		require.NoError(t, err)
	}

	nonce := emu.StateDB().GetNonce(senderAddress)
	tx, err := types.SignTx(
		types.NewTransaction(nonce, receiverAddress, amount, gasLimit, gasPrice, data),
		emu.Signer(),
		sender,
	)
	require.NoError(t, err)

	receipt, res, err := emu.SendTransaction(tx, nil)
	require.NoError(t, err)
	if res.Err != nil {
		t.Logf("Execution failed: %v", res.Err)
	}
	if mintBlock {
		emu.MintBlock()
	}

	return receipt
}

type context struct {
	state     dict.Dict
	bal       map[common.Address]*big.Int
	snapshots []*context
	timestamp uint64
}

var _ Context = &context{}

func newContext(supply map[common.Address]*big.Int) *context {
	return &context{
		state: dict.Dict{},
		bal:   supply,
	}
}

func (*context) BlockKeepAmount() int32 {
	return BlockKeepAll
}

func (*context) GasBurnEnable(bool) {
	panic("unimplemented")
}

func (*context) GasBurnEnabled() bool {
	panic("unimplemented")
}

func (*context) WithoutGasBurn(f func()) {
	f()
}

func (*context) GasLimits() GasLimits {
	return gasLimits
}

func (*context) MagicContracts() map[common.Address]vm.ISCMagicContract {
	return nil
}

func (ctx *context) State() kv.KVStore {
	return ctx.state
}

func (ctx *context) Timestamp() uint64 {
	return ctx.timestamp
}

func (ctx *context) GetBaseTokensBalance(addr common.Address) *big.Int {
	bal := ctx.bal[addr]
	if bal == nil {
		return big.NewInt(0)
	}
	return bal
}

func (ctx *context) AddBaseTokensBalance(addr common.Address, amount *big.Int) {
	ctx.bal[addr] = new(big.Int).Add(ctx.bal[addr], ctx.bal[addr])
}

func (ctx *context) SubBaseTokensBalance(addr common.Address, amount *big.Int) {
	ctx.bal[addr] = new(big.Int).Sub(ctx.bal[addr], amount)
}

func (*context) BaseTokensDecimals() uint8 {
	return 18 // same as ether decimals
}

func (ctx *context) RevertToSnapshot(i int) {
	ctx.state = ctx.snapshots[i].state
	ctx.bal = ctx.snapshots[i].bal
}

func (ctx *context) TakeSnapshot() int {
	ctx.snapshots = append(ctx.snapshots, &context{
		state: ctx.state.Clone(),
		bal:   maps.Clone(ctx.bal),
	})
	return len(ctx.snapshots) - 1
}

func TestBlockchain(t *testing.T) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).SetUint64(math.MaxUint64)

	genesisAlloc := map[common.Address]types.Account{}
	ctx := newContext(map[common.Address]*big.Int{
		faucetAddress: faucetSupply,
	})

	Init(ctx.State(), evm.DefaultChainID, ctx.GasLimits(), ctx.Timestamp(), genesisAlloc)
	ctx.timestamp++
	emu := NewEVMEmulator(ctx)

	// some assertions
	{
		require.EqualValues(t, evm.DefaultChainID, emu.BlockchainDB().GetChainID())

		genesis := emu.BlockchainDB().GetBlockByNumber(0)
		require.NotNil(t, genesis)

		// the genesis block has block number 0
		require.EqualValues(t, 0, genesis.NumberU64())
		// the genesis block has 0 transactions
		require.EqualValues(t, 0, genesis.Transactions().Len())

		{
			// assert that current block is genesis
			block := emu.BlockchainDB().GetCurrentBlock()
			require.NotNil(t, block)
			require.EqualValues(t, gasLimits.Block, block.Header().GasLimit)
		}

		{
			// same, getting the block by hash
			genesis2 := emu.BlockchainDB().GetBlockByHash(genesis.Hash())
			require.NotNil(t, genesis2)
			require.EqualValues(t, genesis.Hash(), genesis2.Hash())
		}

		{
			state := emu.StateDB()
			// check the balance of the faucet address
			require.EqualValues(t, faucetSupply, state.GetBalance(faucetAddress).ToBig())
		}
	}

	// check the balances
	require.EqualValues(t, faucetSupply, emu.StateDB().GetBalance(faucetAddress).ToBig())

	// deploy a contract
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	contractAddress, _ := deployEVMContract(
		t,
		emu,
		faucet,
		contractABI,
		evmtest.StorageContractBytecode,
		uint32(42),
	)
	ctx.timestamp++

	require.EqualValues(t, 1, emu.BlockchainDB().GetNumber())
	block := emu.BlockchainDB().GetCurrentBlock()
	require.EqualValues(t, 1, block.Header().Number.Uint64())
	receipts := emu.BlockchainDB().GetReceiptsByBlockNumber(1)
	require.Len(t, receipts, 1)
	receipt := receipts[0]
	require.EqualValues(t, receipt.Bloom, block.Bloom())
	require.EqualValues(t, receipt.GasUsed, block.GasUsed())
	require.EqualValues(t, emu.BlockchainDB().GetBlockByNumber(0).Hash(), block.ParentHash())
	require.NotEmpty(t, emu.StateDB().GetCode(contractAddress))
}

func TestBlockchainPersistence(t *testing.T) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)

	genesisAlloc := map[common.Address]types.Account{}
	ctx := newContext(map[common.Address]*big.Int{})

	Init(ctx.State(), evm.DefaultChainID, ctx.GasLimits(), ctx.Timestamp(), genesisAlloc)
	ctx.timestamp++

	// deploy a contract using one instance of EVMEmulator
	var contractAddress common.Address
	func() {
		emu := NewEVMEmulator(ctx)
		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)

		contractAddress, _ = deployEVMContract(
			t,
			emu,
			faucet,
			contractABI,
			evmtest.StorageContractBytecode,
			uint32(42),
		)
		ctx.timestamp++
	}()

	// initialize a new EVMEmulator using the same DB and check the state
	{
		emu := NewEVMEmulator(ctx)
		// check the contract address
		require.NotEmpty(t, emu.StateDB().GetCode(contractAddress))
	}
}

type contractFnCaller func(
	mintBlock bool,
	sender *ecdsa.PrivateKey,
	gasLimit uint64,
	name string,
	args ...interface{},
) *types.Receipt

func deployEVMContract(t testing.TB, emu *EVMEmulator, creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (common.Address, contractFnCaller) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := emu.StateDB().GetNonce(creatorAddress)

	txValue := big.NewInt(0)

	// initialize number as 42
	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(t, err)
	require.NotEmpty(t, constructorArguments)

	data := []byte{}
	data = append(data, contractBytecode...)
	data = append(data, constructorArguments...)

	gasLimit, err := estimateGas(ethereum.CallMsg{
		From:  creatorAddress,
		Value: txValue,
		Data:  data,
	}, emu)
	require.NoError(t, err)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, txValue, gasLimit, gasPrice, data),
		emu.Signer(),
		creator,
	)
	require.NoError(t, err)

	receipt, res, err := emu.SendTransaction(tx, nil)
	require.NoError(t, err)
	require.NoError(t, res.Err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	emu.MintBlock()

	contractAddress := crypto.CreateAddress(creatorAddress, nonce)

	// assertions
	{
		require.EqualValues(t, 1, emu.BlockchainDB().GetNumber())

		// verify contract address
		{
			require.EqualValues(t, contractAddress, receipt.ContractAddress)
			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
		}

		// verify contract code
		{
			code := emu.StateDB().GetCode(contractAddress)
			require.NotEmpty(t, code)
		}
	}

	callFn := func(
		mintBlock bool,
		sender *ecdsa.PrivateKey,
		gasLimit uint64,
		name string,
		args ...interface{},
	) *types.Receipt {
		callArguments, err := contractABI.Pack(name, args...)
		require.NoError(t, err)
		receipt := sendTransaction(t, emu, sender, contractAddress, big.NewInt(0), callArguments, gasLimit, mintBlock)
		t.Logf("callFn %s Status: %d", name, receipt.Status)
		t.Log("Logs:")
		for _, log := range receipt.Logs {
			t.Logf("  - %s %x", log.Address, log.Data)
			ev, err := contractABI.EventByID(log.Topics[0])
			if err != nil {
				t.Logf("    - log is not an event: %+v", log.Topics)
			} else {
				t.Logf("    - %s %+v", ev, log.Topics[1:])
			}
		}
		return receipt
	}

	return contractAddress, callFn
}

func TestStorageContract(t *testing.T) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(t, err)

	genesisAlloc := map[common.Address]types.Account{}
	ctx := newContext(map[common.Address]*big.Int{})

	Init(ctx.State(), evm.DefaultChainID, ctx.GasLimits(), ctx.Timestamp(), genesisAlloc)
	ctx.timestamp++
	emu := NewEVMEmulator(ctx)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)

	contractAddress, callFn := deployEVMContract(
		t,
		emu,
		faucet,
		contractABI,
		evmtest.StorageContractBytecode,
		uint32(42),
	)
	ctx.timestamp++

	// call `retrieve` view, get 42
	{
		callArguments, err := contractABI.Pack("retrieve")
		require.NoError(t, err)
		require.NotEmpty(t, callArguments)

		res, err := emu.CallContract(ethereum.CallMsg{To: &contractAddress, Data: callArguments}, false)
		require.NoError(t, err)
		require.NotEmpty(t, res)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", res.Return())
		require.NoError(t, err)
		require.EqualValues(t, 42, v)

		// no state change
		require.EqualValues(t, 1, emu.BlockchainDB().GetNumber())
	}

	// send tx that calls `store(43)`
	{
		callFn(true, faucet, 0, "store", uint32(43))
		ctx.timestamp++
		require.EqualValues(t, 2, emu.BlockchainDB().GetNumber())
	}

	// call `retrieve` view again, get 43
	{
		callArguments, err := contractABI.Pack("retrieve")
		require.NoError(t, err)

		res, err := emu.CallContract(ethereum.CallMsg{
			To:   &contractAddress,
			Data: callArguments,
		}, false)
		require.NoError(t, err)
		require.NotEmpty(t, res)

		var v uint32
		err = contractABI.UnpackIntoInterface(&v, "retrieve", res.Return())
		require.NoError(t, err)
		require.EqualValues(t, 43, v)

		// no state change
		require.EqualValues(t, 2, emu.BlockchainDB().GetNumber())
	}
}

func TestERC20Contract(t *testing.T) {
	genesisAlloc := map[common.Address]types.Account{}
	ctx := newContext(map[common.Address]*big.Int{})

	Init(ctx.State(), evm.DefaultChainID, ctx.GasLimits(), ctx.Timestamp(), genesisAlloc)
	ctx.timestamp++
	emu := NewEVMEmulator(ctx)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(t, err)

	erc20Owner, err := crypto.GenerateKey()
	require.NoError(t, err)
	erc20OwnerAddress := crypto.PubkeyToAddress(erc20Owner.PublicKey)

	contractAddress, callFn := deployEVMContract(
		t,
		emu,
		erc20Owner,
		contractABI,
		evmtest.ERC20ContractBytecode,
		"TestCoin",
		"TEST",
	)
	ctx.timestamp++

	callIntViewFn := func(name string, args ...interface{}) *big.Int {
		callArguments, err2 := contractABI.Pack(name, args...)
		require.NoError(t, err2)

		res, err2 := emu.CallContract(ethereum.CallMsg{To: &contractAddress, Data: callArguments}, false)
		require.NoError(t, err2)

		v := new(big.Int)
		err2 = contractABI.UnpackIntoInterface(&v, name, res.Return())
		require.NoError(t, err2)
		return v
	}

	// call `totalSupply` view
	{
		v := callIntViewFn("totalSupply")
		// 100 * 10^18
		expected := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		require.Zero(t, v.Cmp(expected))
	}

	recipient, err := crypto.GenerateKey()
	require.NoError(t, err)
	recipientAddress := crypto.PubkeyToAddress(recipient.PublicKey)

	transferAmount := big.NewInt(1337)

	// call `transfer` => send 1337 TestCoin to recipientAddress
	receipt := callFn(true, erc20Owner, 0, "transfer", recipientAddress, transferAmount)
	ctx.timestamp++
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))

	// call `transferFrom` as recipient without allowance => get error
	receipt = callFn(true, recipient, gasLimits.Call, "transferFrom", erc20OwnerAddress, recipientAddress, transferAmount)
	ctx.timestamp++
	require.Equal(t, types.ReceiptStatusFailed, receipt.Status)
	require.Equal(t, 0, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(transferAmount))

	// call `approve` as erc20Owner
	receipt = callFn(true, erc20Owner, 0, "approve", recipientAddress, transferAmount)
	ctx.timestamp++
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `transferFrom` as recipient with allowance => ok
	receipt = callFn(true, recipient, 0, "transferFrom", erc20OwnerAddress, recipientAddress, transferAmount)
	ctx.timestamp++
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	require.Equal(t, 1, len(receipt.Logs))

	// call `balanceOf` view => check balance of recipient = 2 * 1337 TestCoin
	require.Zero(t, callIntViewFn("balanceOf", recipientAddress).Cmp(new(big.Int).Mul(transferAmount, big.NewInt(2))))

	// call `transfer` 10 times on the same block, check receipt and logs derived fields
	{
		for i := 0; i < 10; i++ {
			receipt := callFn(false, erc20Owner, 0, "transfer", recipientAddress, transferAmount)
			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
			require.Equal(t, 1, len(receipt.Logs))
		}
		emu.MintBlock()
		ctx.timestamp++
	}
	{
		blockNumber := emu.BlockchainDB().GetNumber()
		block := emu.BlockchainDB().GetBlockByNumber(blockNumber)
		receipts := emu.BlockchainDB().GetReceiptsByBlockNumber(blockNumber)
		require.Len(t, receipts, 10)
		gas := uint64(0)
		for i, r := range receipts {
			tx := emu.BlockchainDB().GetTransactionByBlockNumberAndIndex(blockNumber, uint32(i))
			require.Equal(t, tx.Hash(), r.TxHash)
			require.NotZero(t, r.GasUsed)
			gas += r.GasUsed
			require.Equal(t, block.Hash(), r.BlockHash)
			require.Equal(t, blockNumber, r.BlockNumber.Uint64())
			require.EqualValues(t, i, r.TransactionIndex)

			require.Len(t, r.Logs, 1)
			log := r.Logs[0]
			require.Equal(t, blockNumber, log.BlockNumber)
			require.Equal(t, tx.Hash(), log.TxHash)
			require.EqualValues(t, i, log.TxIndex)
			require.Equal(t, block.Hash(), log.BlockHash)
			require.EqualValues(t, i, log.Index)
		}
		require.Equal(t, gas, receipts[len(receipts)-1].CumulativeGasUsed, gas)
	}
}

func initBenchmark(b *testing.B) (*EVMEmulator, []*types.Transaction, *context) {
	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	require.NoError(b, err)

	genesisAlloc := map[common.Address]types.Account{}
	ctx := newContext(map[common.Address]*big.Int{})

	Init(ctx.State(), evm.DefaultChainID, ctx.GasLimits(), ctx.Timestamp(), genesisAlloc)
	ctx.timestamp++
	emu := NewEVMEmulator(ctx)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(b, err)

	contractAddress, _ := deployEVMContract(
		b,
		emu,
		faucet,
		contractABI,
		evmtest.StorageContractBytecode,
		uint32(42),
	)
	ctx.timestamp++

	txs := make([]*types.Transaction, b.N)
	for i := 0; i < b.N; i++ {
		sender, err := crypto.GenerateKey() // send from a new address so that nonce is always 0
		require.NoError(b, err)

		amount := big.NewInt(0)
		nonce := uint64(0)

		callArguments, err := contractABI.Pack("store", uint32(i))
		require.NoError(b, err)

		gasLimit := uint64(100000)

		txs[i], err = types.SignTx(
			types.NewTransaction(nonce, contractAddress, amount, gasLimit, gasPrice, callArguments),
			emu.Signer(),
			sender,
		)
		require.NoError(b, err)
	}

	return emu, txs, ctx
}

// benchmarkEVMEmulator is a benchmark for the EVMEmulator that sends N EVM transactions
// calling `storage.store()`, committing a block every k transactions.
//
// run with: go test -benchmem -cpu=1 -run=' ' -bench='Bench.*'
//
// To generate mem and cpu profiles, add -cpuprofile=cpu.out -memprofile=mem.out
// Then: go tool pprof -http :8080 {cpu,mem}.out
func benchmarkEVMEmulator(b *testing.B, k int) {
	// setup: deploy the storage contract and prepare N transactions to send
	emu, txs, ctx := initBenchmark(b)

	var chunks [][]*types.Transaction
	var chunk []*types.Transaction
	for _, tx := range txs {
		chunk = append(chunk, tx)
		if len(chunk) == k {
			chunks = append(chunks, chunk)
			chunk = nil
		}
	}
	if len(chunk) > 0 {
		chunks = append(chunks, chunk)
	}

	b.ResetTimer()
	for _, chunk := range chunks {
		for _, tx := range chunk {
			receipt, res, err := emu.SendTransaction(tx, nil)
			ctx.timestamp++
			require.NoError(b, err)
			require.NoError(b, res.Err)
			require.Equal(b, types.ReceiptStatusSuccessful, receipt.Status)
		}
		emu.MintBlock()
	}

	b.ReportMetric(dbSize(ctx.State())/float64(b.N), "db:bytes/op")
}

func BenchmarkEVMEmulator1(b *testing.B)   { benchmarkEVMEmulator(b, 1) }
func BenchmarkEVMEmulator10(b *testing.B)  { benchmarkEVMEmulator(b, 10) }
func BenchmarkEVMEmulator50(b *testing.B)  { benchmarkEVMEmulator(b, 50) }
func BenchmarkEVMEmulator100(b *testing.B) { benchmarkEVMEmulator(b, 100) }

func dbSize(db kv.KVStore) float64 {
	r := float64(0)
	db.Iterate("", func(key kv.Key, value []byte) bool {
		r += float64(len(key) + len(value))
		return true
	})
	return r
}
