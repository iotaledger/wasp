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
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/isccontract"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

type evmChainInstance struct {
	t            testing.TB
	solo         *solo.Solo
	soloChain    *solo.Chain
	faucetKey    *ecdsa.PrivateKey
	faucetSupply *big.Int
	chainID      int
}

type evmContractInstance struct {
	chain         *evmChainInstance
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

type iotaCallOptions struct {
	wallet    *cryptolib.KeyPair
	before    func(*solo.CallParams)
	offledger bool
}

type ethCallOptions struct {
	iota   iotaCallOptions
	sender *ecdsa.PrivateKey
	value  *big.Int
}

func initEVM(t testing.TB, nativeContracts ...*coreutil.ContractProcessor) *evmChainInstance {
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

func initEVMWithSolo(t testing.TB, env *solo.Solo) *evmChainInstance {
	faucetKey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
	chainID := evm.DefaultChainID
	return &evmChainInstance{
		t:    t,
		solo: env,
		soloChain: env.NewChain(nil, "ch1", dict.Dict{
			root.ParamEVM(evm.FieldChainID): codec.EncodeUint16(uint16(chainID)),
			root.ParamEVM(evm.FieldGenesisAlloc): evmtypes.EncodeGenesisAlloc(map[common.Address]core.GenesisAccount{
				crypto.PubkeyToAddress(faucetKey.PublicKey): {Balance: faucetSupply},
			}),
		}),
		faucetKey:    faucetKey,
		faucetSupply: faucetSupply,
		chainID:      chainID,
	}
}

func (e *evmChainInstance) parseIotaCallOptions(opts []iotaCallOptions) iotaCallOptions {
	if len(opts) == 0 {
		opts = []iotaCallOptions{{}}
	}
	opt := opts[0]
	if opt.wallet == nil {
		opt.wallet = e.soloChain.OriginatorPrivateKey
	}
	return opt
}

func (e *evmChainInstance) buildSoloRequest(funName string, params ...interface{}) *solo.CallParams {
	return solo.NewCallParams(evm.Contract.Name, funName, params...)
}

func (e *evmChainInstance) postRequest(opts []iotaCallOptions, funName string, params ...interface{}) (dict.Dict, error) {
	opt := e.parseIotaCallOptions(opts)
	req := e.buildSoloRequest(funName, params...)
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

func (e *evmChainInstance) resolveError(err error) error {
	if err == nil {
		return nil
	}
	if vmError, ok := err.(*iscp.UnresolvedVMError); ok {
		resolvedErr := e.soloChain.ResolveVMError(vmError)
		return resolvedErr.AsGoError()
	}
	return err
}

func (e *evmChainInstance) callView(funName string, params ...interface{}) (dict.Dict, error) {
	ret, err := e.soloChain.CallView(evm.Contract.Name, funName, params...)
	if err != nil {
		return nil, fmt.Errorf("CallView failed: %w", e.resolveError(err))
	}
	return ret, nil
}

func (e *evmChainInstance) setBlockTime(t uint32) {
	_, err := e.postRequest(nil, evm.FuncSetBlockTime.Name, evm.FieldBlockTime, codec.EncodeUint32(t))
	require.NoError(e.t, err)
}

func (e *evmChainInstance) getBlockNumber() uint64 {
	ret, err := e.callView(evm.FuncGetBlockNumber.Name)
	require.NoError(e.t, err)
	return new(big.Int).SetBytes(ret.MustGet(evm.FieldResult)).Uint64()
}

func (e *evmChainInstance) getBlockByNumber(n uint64) *types.Block {
	ret, err := e.callView(evm.FuncGetBlockByNumber.Name, evm.FieldBlockNumber, new(big.Int).SetUint64(n).Bytes())
	require.NoError(e.t, err)
	block, err := evmtypes.DecodeBlock(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return block
}

func (e *evmChainInstance) getCode(addr common.Address) []byte {
	ret, err := e.callView(evm.FuncGetCode.Name, evm.FieldAddress, addr.Bytes())
	require.NoError(e.t, err)
	return ret.MustGet(evm.FieldResult)
}

func (e *evmChainInstance) getGasRatio() util.Ratio32 {
	ret, err := e.callView(evm.FuncGetGasRatio.Name)
	require.NoError(e.t, err)
	ratio, err := codec.DecodeRatio32(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return ratio
}

func (e *evmChainInstance) setGasRatio(newGasRatio util.Ratio32, opts ...iotaCallOptions) error {
	_, err := e.postRequest(opts, evm.FuncSetGasRatio.Name, evm.FieldGasRatio, newGasRatio)
	return err
}

func (e *evmChainInstance) faucetAddress() common.Address {
	return crypto.PubkeyToAddress(e.faucetKey.PublicKey)
}

func (e *evmChainInstance) getBalance(addr common.Address) *big.Int {
	ret, err := e.callView(evm.FuncGetBalance.Name, evm.FieldAddress, addr.Bytes())
	require.NoError(e.t, err)
	bal := big.NewInt(0)
	bal.SetBytes(ret.MustGet(evm.FieldResult))
	return bal
}

func (e *evmChainInstance) getNonce(addr common.Address) uint64 {
	ret, err := e.callView(evm.FuncGetNonce.Name, evm.FieldAddress, addr.Bytes())
	require.NoError(e.t, err)
	nonce, err := codec.DecodeUint64(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return nonce
}

func (e *evmChainInstance) ISCContract(defaultSender *ecdsa.PrivateKey) *iscContractInstance {
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

func (e *evmChainInstance) deployISCTestContract(creator *ecdsa.PrivateKey) *iscTestContractInstance {
	return &iscTestContractInstance{e.deployContract(creator, evmtest.ISCTestContractABI, evmtest.ISCTestContractBytecode)}
}

func (e *evmChainInstance) deployStorageContract(creator *ecdsa.PrivateKey, n uint32) *storageContractInstance { // nolint:unparam
	return &storageContractInstance{e.deployContract(creator, evmtest.StorageContractABI, evmtest.StorageContractBytecode, n)}
}

func (e *evmChainInstance) deployERC20Contract(creator *ecdsa.PrivateKey, name, symbol string) *erc20ContractInstance {
	return &erc20ContractInstance{e.deployContract(creator, evmtest.ERC20ContractABI, evmtest.ERC20ContractBytecode, name, symbol)}
}

func (e *evmChainInstance) deployLoopContract(creator *ecdsa.PrivateKey) *loopContractInstance {
	return &loopContractInstance{e.deployContract(creator, evmtest.LoopContractABI, evmtest.LoopContractBytecode)}
}

func (e *evmChainInstance) signer() types.Signer {
	return evmtypes.Signer(big.NewInt(int64(e.chainID)))
}

func (e *evmChainInstance) deployContract(creator *ecdsa.PrivateKey, abiJSON string, bytecode []byte, args ...interface{}) *evmContractInstance {
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

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, 0, evm.GasPrice, data),
		e.signer(),
		e.faucetKey,
	)
	require.NoError(e.t, err)

	txdata, err := tx.MarshalBinary()
	require.NoError(e.t, err)

	req := solo.NewCallParamsFromDict(evm.Contract.Name, evm.FuncSendTransaction.Name, dict.Dict{
		evm.FieldTransactionData: txdata,
	})

	iscpGas, iscpGasFee, err := e.soloChain.EstimateGasOnLedger(req, nil, true)
	require.NoError(e.t, e.resolveError(err))
	req.WithGasBudget(iscpGas)

	// deposit gas fee
	depositGasFeeReq := solo.NewCallParams(accounts.Contract.Name, accounts.FuncDeposit.Name)
	_, fee2, err := e.soloChain.EstimateGasOnLedger(depositGasFeeReq, nil, true)
	require.NoError(e.t, e.resolveError(err))
	_, err = e.soloChain.PostRequestSync(depositGasFeeReq.AddIotas(iscpGasFee+fee2), nil)

	// send EVM tx
	_, err = e.soloChain.PostRequestOffLedger(req, nil)

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
	if opt.iota.wallet == nil {
		opt.iota.wallet = e.chain.soloChain.OriginatorPrivateKey
	}
	if opt.value == nil {
		opt.value = big.NewInt(0)
	}
	return opt
}

func (e *evmContractInstance) buildEthTxData(opts []ethCallOptions, fnName string, args ...interface{}) ([]byte, *types.Transaction, ethCallOptions) {
	callArguments, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	opt := e.parseEthCallOptions(opts, callArguments)

	senderAddress := crypto.PubkeyToAddress(opt.sender.PublicKey)

	nonce := e.chain.getNonce(senderAddress)

	unsignedTx := types.NewTransaction(nonce, e.address, opt.value, 0, evm.GasPrice, callArguments)

	tx, err := types.SignTx(unsignedTx, e.chain.signer(), opt.sender)
	require.NoError(e.chain.t, err)

	txdata, err := tx.MarshalBinary()
	require.NoError(e.chain.t, err)
	return txdata, tx, opt
}

type callFnResult struct {
	tx          *types.Transaction
	evmReceipt  *types.Receipt
	iscpReceipt *blocklog.RequestReceipt
}

func (e *evmContractInstance) callFn(opts []ethCallOptions, fnName string, args ...interface{}) (res callFnResult, err error) {
	e.chain.t.Logf("callFn: %s %+v", fnName, args)

	txdata, tx, opt := e.buildEthTxData(opts, fnName, args...)
	res.tx = tx

	_, err = e.chain.postRequest([]iotaCallOptions{opt.iota}, evm.FuncSendTransaction.Name, evm.FieldTransactionData, txdata)
	if err != nil {
		return
	}

	res.iscpReceipt = e.chain.soloChain.LastReceipt()

	receiptResult, err := e.chain.callView(evm.FuncGetReceipt.Name, evm.FieldTransactionHash, tx.Hash().Bytes())
	require.NoError(e.chain.t, err)

	res.evmReceipt, err = evmtypes.DecodeReceiptFull(receiptResult.MustGet(evm.FieldResult))
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

func (e *evmContractInstance) callView(opts []ethCallOptions, fnName string, args []interface{}, v interface{}) {
	e.chain.t.Logf("callView: %s %+v", fnName, args)
	callArguments, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	opt := e.parseEthCallOptions(opts, callArguments)
	senderAddress := crypto.PubkeyToAddress(opt.sender.PublicKey)
	callMsg := e.callMsg(ethereum.CallMsg{
		From:     senderAddress,
		Gas:      0,
		GasPrice: evm.GasPrice,
		Value:    opt.value,
		Data:     callArguments,
	})
	ret, err := e.chain.callView(evm.FuncCallContract.Name, evm.FieldCallMsg, evmtypes.EncodeCallMsg(callMsg))
	require.NoError(e.chain.t, err)
	if v != nil {
		err = e.abi.UnpackIntoInterface(v, fnName, ret.MustGet(evm.FieldResult))
		require.NoError(e.chain.t, err)
	}
}

func (i *iscTestContractInstance) getChainID() *iscp.ChainID {
	var v isccontract.ISCChainID
	i.callView(nil, "getChainID", nil, &v)
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
	s.callView(nil, "retrieve", nil, &v)
	return v
}

func (s *storageContractInstance) store(n uint32, opts ...ethCallOptions) (res callFnResult, err error) {
	return s.callFn(opts, "store", n)
}

func (e *erc20ContractInstance) balanceOf(addr common.Address, opts ...ethCallOptions) *big.Int {
	v := new(big.Int)
	e.callView(opts, "balanceOf", []interface{}{addr}, &v)
	return v
}

func (e *erc20ContractInstance) totalSupply(opts ...ethCallOptions) *big.Int {
	v := new(big.Int)
	e.callView(opts, "totalSupply", nil, &v)
	return v
}

func (e *erc20ContractInstance) transfer(recipientAddress common.Address, amount *big.Int, opts ...ethCallOptions) (res callFnResult, err error) {
	return e.callFn(opts, "transfer", recipientAddress, amount)
}

func (l *loopContractInstance) loop(opts ...ethCallOptions) (res callFnResult, err error) {
	return l.callFn(opts, "loop")
}

func generateEthereumKey(t testing.TB) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}
