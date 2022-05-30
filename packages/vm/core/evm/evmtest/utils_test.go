// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/isccontract"
	"github.com/stretchr/testify/require"
)

var latestBlock = rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber)

type soloChainEnv struct {
	t          testing.TB
	solo       *solo.Solo
	soloChain  *solo.Chain
	evmChainID uint16
	evmChain   *jsonrpc.EVMChain
}

type evmContractInstance struct {
	chain         *soloChainEnv
	defaultSender *ecdsa.PrivateKey
	address       common.Address
	abi           abi.ABI
}

type iscContractInstance struct {
	*evmContractInstance
}

type iscTestContractInstance struct {
	*evmContractInstance
}

type storageContractInstance struct {
	*evmContractInstance
}

type erc20ContractInstance struct {
	*evmContractInstance
}

type loopContractInstance struct {
	*evmContractInstance
}

type fibonacciContractInstance struct {
	*evmContractInstance
}

type iotaCallOptions struct {
	wallet    *cryptolib.KeyPair
	before    func(*solo.CallParams)
	offledger bool
}

type ethCallOptions struct {
	sender    *ecdsa.PrivateKey
	value     *big.Int
	allowance *iscp.Allowance
	gasLimit  uint64
}

func initEVM(t testing.TB, nativeContracts ...*coreutil.ContractProcessor) *soloChainEnv {
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustDustDeposit: true,
		Debug:                 true,
		PrintStackTrace:       true,
	})
	for _, c := range nativeContracts {
		env = env.WithNativeContract(c)
	}
	return initEVMWithSolo(t, env)
}

func initEVMWithSolo(t testing.TB, env *solo.Solo) *soloChainEnv {
	soloChain := env.NewChain(nil, "ch1")
	return &soloChainEnv{
		t:          t,
		solo:       env,
		soloChain:  soloChain,
		evmChainID: evm.DefaultChainID,
		evmChain:   soloChain.EVM(),
	}
}

func (e *soloChainEnv) parseIotaCallOptions(opts []iotaCallOptions) iotaCallOptions {
	if len(opts) == 0 {
		opts = []iotaCallOptions{{}}
	}
	opt := opts[0]
	if opt.wallet == nil {
		opt.wallet = e.soloChain.OriginatorPrivateKey
	}
	return opt
}

func (e *soloChainEnv) postRequest(opts []iotaCallOptions, funName string, params ...interface{}) (dict.Dict, error) {
	opt := e.parseIotaCallOptions(opts)
	req := solo.NewCallParams(evm.Contract.Name, funName, params...)
	if opt.before != nil {
		opt.before(req)
	}
	if req.GasBudget() == 0 {
		gasBudget, gasFee, err := e.soloChain.EstimateGasOnLedger(req, opt.wallet, true)
		if err != nil {
			return nil, fmt.Errorf("could not estimate gas: %w", e.resolveError(err))
		}
		req.WithGasBudget(gasBudget).AddIotas(gasFee)
	}
	if opt.offledger {
		ret, err := e.soloChain.PostRequestOffLedger(req, opt.wallet)
		if err != nil {
			return nil, fmt.Errorf("PostRequestSync failed: %w", e.resolveError(err))
		}
		return ret, nil
	}
	ret, err := e.soloChain.PostRequestSync(req, opt.wallet)
	if err != nil {
		return nil, fmt.Errorf("PostRequestSync failed: %w", e.resolveError(err))
	}
	return ret, nil
}

func (e *soloChainEnv) resolveError(err error) error {
	if err == nil {
		return nil
	}
	if vmError, ok := err.(*iscp.UnresolvedVMError); ok {
		resolvedErr := e.soloChain.ResolveVMError(vmError)
		return resolvedErr.AsGoError()
	}
	return err
}

func (e *soloChainEnv) callView(funName string, params ...interface{}) (dict.Dict, error) {
	ret, err := e.soloChain.CallView(evm.Contract.Name, funName, params...)
	if err != nil {
		return nil, fmt.Errorf("CallView failed: %w", e.resolveError(err))
	}
	return ret, nil
}

func (e *soloChainEnv) getBlockNumber() uint64 {
	n, err := e.evmChain.BlockNumber()
	require.NoError(e.t, err)
	return n.Uint64()
}

func (e *soloChainEnv) getBlockByNumber(n uint64) *types.Block {
	block, err := e.evmChain.BlockByNumber(new(big.Int).SetUint64(n))
	require.NoError(e.t, err)
	return block
}

func (e *soloChainEnv) getCode(addr common.Address) []byte {
	ret, err := e.evmChain.Code(addr, latestBlock)
	require.NoError(e.t, err)
	return ret
}

func (e *soloChainEnv) getGasRatio() util.Ratio32 {
	ret, err := e.callView(evm.FuncGetGasRatio.Name)
	require.NoError(e.t, err)
	ratio, err := codec.DecodeRatio32(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return ratio
}

func (e *soloChainEnv) setGasRatio(newGasRatio util.Ratio32, opts ...iotaCallOptions) error {
	_, err := e.postRequest(opts, evm.FuncSetGasRatio.Name, evm.FieldGasRatio, newGasRatio)
	return err
}

func (e *soloChainEnv) getBalance(addr common.Address) *big.Int {
	bal, err := e.evmChain.Balance(addr, latestBlock)
	require.NoError(e.t, err)
	return bal
}

func (e *soloChainEnv) getNonce(addr common.Address) uint64 {
	ret, err := e.callView(evm.FuncGetNonce.Name, evm.FieldAddress, addr.Bytes())
	require.NoError(e.t, err)
	nonce, err := codec.DecodeUint64(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return nonce
}

func (e *soloChainEnv) ISCContract(defaultSender *ecdsa.PrivateKey) *iscContractInstance {
	iscABI, err := abi.JSON(strings.NewReader(isccontract.ABI))
	require.NoError(e.t, err)
	return &iscContractInstance{
		evmContractInstance: &evmContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       vm.ISCAddress,
			abi:           iscABI,
		},
	}
}

func (e *soloChainEnv) deployISCTestContract(creator *ecdsa.PrivateKey) *iscTestContractInstance {
	return &iscTestContractInstance{e.deployContract(creator, evmtest.ISCTestContractABI, evmtest.ISCTestContractBytecode)}
}

func (e *soloChainEnv) deployStorageContract(creator *ecdsa.PrivateKey, n uint32) *storageContractInstance { // nolint:unparam
	return &storageContractInstance{e.deployContract(creator, evmtest.StorageContractABI, evmtest.StorageContractBytecode, n)}
}

func (e *soloChainEnv) deployERC20Contract(creator *ecdsa.PrivateKey, name, symbol string) *erc20ContractInstance {
	return &erc20ContractInstance{e.deployContract(creator, evmtest.ERC20ContractABI, evmtest.ERC20ContractBytecode, name, symbol)}
}

func (e *soloChainEnv) deployLoopContract(creator *ecdsa.PrivateKey) *loopContractInstance {
	return &loopContractInstance{e.deployContract(creator, evmtest.LoopContractABI, evmtest.LoopContractBytecode)}
}

func (e *soloChainEnv) deployFibonacciContract(creator *ecdsa.PrivateKey) *fibonacciContractInstance {
	return &fibonacciContractInstance{e.deployContract(creator, evmtest.FibonacciContractABI, evmtest.FibonacciContractByteCode)}
}

func (e *soloChainEnv) signer() types.Signer {
	return evmutil.Signer(big.NewInt(int64(e.evmChainID)))
}

func (e *soloChainEnv) deployContract(creator *ecdsa.PrivateKey, abiJSON string, bytecode []byte, args ...interface{}) *evmContractInstance {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := e.getNonce(creatorAddress)

	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	require.NoError(e.t, err)
	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(e.t, err)

	data := []byte{}
	data = append(data, bytecode...)
	data = append(data, constructorArguments...)

	value := big.NewInt(0)

	gasLimit, err := e.evmChain.EstimateGas(ethereum.CallMsg{
		From:     creatorAddress,
		GasPrice: evm.GasPrice,
		Value:    value,
		Data:     data,
	}, nil)
	require.NoError(e.t, err)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, gasLimit, evm.GasPrice, data),
		e.signer(),
		creator,
	)
	require.NoError(e.t, err)

	err = e.evmChain.SendTransaction(tx, nil)
	require.NoError(e.t, err)

	return &evmContractInstance{
		chain:         e,
		defaultSender: creator,
		address:       crypto.CreateAddress(creatorAddress, nonce),
		abi:           contractABI,
	}
}

func (e *evmContractInstance) callMsg(callMsg ethereum.CallMsg) ethereum.CallMsg {
	callMsg.To = &e.address
	return callMsg
}

func (e *evmContractInstance) parseEthCallOptions(opts []ethCallOptions, callData []byte) ethCallOptions {
	var opt ethCallOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.sender == nil {
		opt.sender = e.defaultSender
	}
	if opt.value == nil {
		opt.value = big.NewInt(0)
	}
	if opt.gasLimit == 0 {
		var err error
		senderAddress := crypto.PubkeyToAddress(opt.sender.PublicKey)
		opt.gasLimit, err = e.chain.evmChain.EstimateGas(ethereum.CallMsg{
			From:     senderAddress,
			To:       &e.address,
			GasPrice: evm.GasPrice,
			Value:    opt.value,
			Data:     callData,
		}, opt.allowance)
		require.NoError(e.chain.t, err)
	}
	return opt
}

func (e *evmContractInstance) buildEthTx(opts []ethCallOptions, fnName string, args ...interface{}) (*types.Transaction, *iscp.Allowance) {
	callArguments, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	opt := e.parseEthCallOptions(opts, callArguments)

	senderAddress := crypto.PubkeyToAddress(opt.sender.PublicKey)

	nonce := e.chain.getNonce(senderAddress)

	unsignedTx := types.NewTransaction(nonce, e.address, opt.value, opt.gasLimit, evm.GasPrice, callArguments)

	tx, err := types.SignTx(unsignedTx, e.chain.signer(), opt.sender)
	require.NoError(e.chain.t, err)
	return tx, opt.allowance
}

type callFnResult struct {
	tx          *types.Transaction
	evmReceipt  *types.Receipt
	iscpReceipt *blocklog.RequestReceipt
}

func (e *evmContractInstance) callFn(opts []ethCallOptions, fnName string, args ...interface{}) (res callFnResult, err error) {
	e.chain.t.Logf("callFn: %s %+v", fnName, args)

	var allowance *iscp.Allowance
	res.tx, allowance = e.buildEthTx(opts, fnName, args...)

	err = e.chain.evmChain.SendTransaction(res.tx, allowance)
	if err != nil {
		return
	}

	res.iscpReceipt = e.chain.soloChain.LastReceipt()

	res.evmReceipt, err = e.chain.evmChain.TransactionReceipt(res.tx.Hash())
	require.NoError(e.chain.t, err)

	return
}

func (e *evmContractInstance) callFnExpectError(opts []ethCallOptions, fnName string, args ...interface{}) error {
	_, err := e.callFn(opts, fnName, args...)
	require.Error(e.chain.t, err)
	return err
}

func (e *evmContractInstance) callFnExpectEvent(opts []ethCallOptions, eventName string, v interface{}, fnName string, args ...interface{}) {
	res, err := e.callFn(opts, fnName, args...)
	require.NoError(e.chain.t, err)
	require.Equal(e.chain.t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	require.Len(e.chain.t, res.evmReceipt.Logs, 1)
	if v != nil {
		err = e.abi.UnpackIntoInterface(v, eventName, res.evmReceipt.Logs[0].Data)
	}
	require.NoError(e.chain.t, err)
}

func (e *evmContractInstance) callView(fnName string, args []interface{}, v interface{}) {
	e.chain.t.Logf("callView: %s %+v", fnName, args)
	callArguments, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	senderAddress := crypto.PubkeyToAddress(e.defaultSender.PublicKey)
	callMsg := e.callMsg(ethereum.CallMsg{
		From:     senderAddress,
		Gas:      0,
		GasPrice: evm.GasPrice,
		Data:     callArguments,
	})
	ret, err := e.chain.evmChain.CallContract(callMsg, latestBlock)
	require.NoError(e.chain.t, err)
	if v != nil {
		err = e.abi.UnpackIntoInterface(v, fnName, ret)
		require.NoError(e.chain.t, err)
	}
}

func (i *iscTestContractInstance) getChainID() *iscp.ChainID {
	var v isccontract.ISCChainID
	i.callView("getChainID", nil, &v)
	return v.MustUnwrap()
}

func (i *iscTestContractInstance) triggerEvent(s string) (res callFnResult, err error) {
	return i.callFn(nil, "triggerEvent", s)
}

func (i *iscTestContractInstance) triggerEventFail(s string, opts ...ethCallOptions) (res callFnResult, err error) {
	return i.callFn(opts, "triggerEventFail", s)
}

func (s *storageContractInstance) retrieve() uint32 {
	var v uint32
	s.callView("retrieve", nil, &v)
	return v
}

func (s *storageContractInstance) store(n uint32, opts ...ethCallOptions) (res callFnResult, err error) {
	return s.callFn(opts, "store", n)
}

func (e *erc20ContractInstance) balanceOf(addr common.Address) *big.Int {
	v := new(big.Int)
	e.callView("balanceOf", []interface{}{addr}, &v)
	return v
}

func (e *erc20ContractInstance) totalSupply() *big.Int {
	v := new(big.Int)
	e.callView("totalSupply", nil, &v)
	return v
}

func (e *erc20ContractInstance) transfer(recipientAddress common.Address, amount *big.Int, opts ...ethCallOptions) (res callFnResult, err error) {
	return e.callFn(opts, "transfer", recipientAddress, amount)
}

func (l *loopContractInstance) loop(opts ...ethCallOptions) (res callFnResult, err error) {
	return l.callFn(opts, "loop")
}

func (f *fibonacciContractInstance) fib(n uint32, opts ...ethCallOptions) (res callFnResult, err error) {
	return f.callFn(opts, "fib", n)
}

func generateEthereumKey(t testing.TB) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}
