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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/evm/jsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
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

type iscCallOptions struct {
	wallet *cryptolib.KeyPair
}

type ethCallOptions struct {
	sender   *ecdsa.PrivateKey
	value    *big.Int
	gasLimit uint64
	gasPrice *big.Int
}

func initEVM(t testing.TB, nativeContracts ...*coreutil.ContractProcessor) *soloChainEnv {
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustStorageDeposit: true,
		Debug:                    true,
		PrintStackTrace:          true,
	})
	for _, c := range nativeContracts {
		env = env.WithNativeContract(c)
	}
	return initEVMWithSolo(t, env)
}

func initEVMWithSolo(t testing.TB, env *solo.Solo) *soloChainEnv {
	soloChain := env.NewChain()
	return &soloChainEnv{
		t:          t,
		solo:       env,
		soloChain:  soloChain,
		evmChainID: evm.DefaultChainID,
		evmChain:   soloChain.EVM(),
	}
}

func (e *soloChainEnv) parseISCCallOptions(opts []iscCallOptions) iscCallOptions {
	if len(opts) == 0 {
		opts = []iscCallOptions{{}}
	}
	opt := opts[0]
	if opt.wallet == nil {
		opt.wallet = e.soloChain.OriginatorPrivateKey
	}
	return opt
}

func (e *soloChainEnv) resolveError(err error) error {
	if err == nil {
		return nil
	}
	if vmError, ok := err.(*isc.UnresolvedVMError); ok {
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

func (e *soloChainEnv) getCode(addr common.Address) []byte {
	ret, err := e.evmChain.Code(addr, latestBlock)
	require.NoError(e.t, err)
	return ret
}

func (e *soloChainEnv) getGasRatio() util.Ratio32 {
	ret, err := e.soloChain.CallView(governance.Contract.Name, governance.ViewGetEVMGasRatio.Name)
	require.NoError(e.t, err)
	ratio, err := codec.DecodeRatio32(ret.MustGet(governance.ParamEVMGasRatio))
	require.NoError(e.t, err)
	return ratio
}

func (e *soloChainEnv) setGasRatio(newGasRatio util.Ratio32, opts ...iscCallOptions) error {
	opt := e.parseISCCallOptions(opts)
	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetEVMGasRatio.Name, governance.ParamEVMGasRatio, newGasRatio.Bytes())
	_, err := e.soloChain.PostRequestSync(req, opt.wallet)
	return err
}

func (e *soloChainEnv) setFeePolicy(p gas.FeePolicy, opts ...iscCallOptions) error {
	opt := e.parseISCCallOptions(opts)
	req := solo.NewCallParams(
		governance.Contract.Name, governance.FuncSetFeePolicy.Name,
		governance.ParamFeePolicyBytes,
		p.Bytes(),
	)
	_, err := e.soloChain.PostRequestSync(req, opt.wallet)
	return err
}

func (e *soloChainEnv) getNonce(addr common.Address) uint64 {
	ret, err := e.callView(evm.FuncGetNonce.Name, evm.FieldAddress, addr.Bytes())
	require.NoError(e.t, err)
	nonce, err := codec.DecodeUint64(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return nonce
}

func (e *soloChainEnv) contractFromABI(address common.Address, abiJSON string, defaultSender *ecdsa.PrivateKey) *iscContractInstance {
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	require.NoError(e.t, err)
	return &iscContractInstance{
		evmContractInstance: &evmContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       address,
			abi:           parsedABI,
		},
	}
}

func (e *soloChainEnv) ISCMagicSandbox(defaultSender *ecdsa.PrivateKey) *iscContractInstance {
	return e.contractFromABI(iscmagic.Address, iscmagic.SandboxABI, defaultSender)
}

func (e *soloChainEnv) ISCMagicUtil(defaultSender *ecdsa.PrivateKey) *iscContractInstance {
	return e.contractFromABI(iscmagic.Address, iscmagic.UtilABI, defaultSender)
}

func (e *soloChainEnv) ISCMagicAccounts(defaultSender *ecdsa.PrivateKey) *iscContractInstance {
	return e.contractFromABI(iscmagic.Address, iscmagic.AccountsABI, defaultSender)
}

func (e *soloChainEnv) ISCMagicPrivileged(defaultSender *ecdsa.PrivateKey) *iscContractInstance {
	return e.contractFromABI(iscmagic.Address, iscmagic.PrivilegedABI, defaultSender)
}

func (e *soloChainEnv) ERC20BaseTokens(defaultSender *ecdsa.PrivateKey) *iscContractInstance {
	erc20BaseABI, err := abi.JSON(strings.NewReader(iscmagic.ERC20BaseTokensABI))
	require.NoError(e.t, err)
	return &iscContractInstance{
		evmContractInstance: &evmContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC20BaseTokensAddress,
			abi:           erc20BaseABI,
		},
	}
}

func (e *soloChainEnv) ERC20NativeTokens(defaultSender *ecdsa.PrivateKey, foundrySN uint32) *iscContractInstance {
	erc20BaseABI, err := abi.JSON(strings.NewReader(iscmagic.ERC20NativeTokensABI))
	require.NoError(e.t, err)
	return &iscContractInstance{
		evmContractInstance: &evmContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC20NativeTokensAddress(foundrySN),
			abi:           erc20BaseABI,
		},
	}
}

func (e *soloChainEnv) ERC20ExternalNativeTokens(defaultSender *ecdsa.PrivateKey, addr common.Address) *iscContractInstance {
	erc20BaseABI, err := abi.JSON(strings.NewReader(iscmagic.ERC20ExternalNativeTokensABI))
	require.NoError(e.t, err)
	return &iscContractInstance{
		evmContractInstance: &evmContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       addr,
			abi:           erc20BaseABI,
		},
	}
}

func (e *soloChainEnv) ERC721NFTs(defaultSender *ecdsa.PrivateKey) *iscContractInstance {
	erc721ABI, err := abi.JSON(strings.NewReader(iscmagic.ERC721NFTsABI))
	require.NoError(e.t, err)
	return &iscContractInstance{
		evmContractInstance: &evmContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC721NFTsAddress,
			abi:           erc721ABI,
		},
	}
}

func (e *soloChainEnv) ERC721NFTCollection(defaultSender *ecdsa.PrivateKey, collectionID iotago.NFTID) *iscContractInstance {
	erc721NFTCollectionABI, err := abi.JSON(strings.NewReader(iscmagic.ERC721NFTCollectionABI))
	require.NoError(e.t, err)
	return &iscContractInstance{
		evmContractInstance: &evmContractInstance{
			chain:         e,
			defaultSender: defaultSender,
			address:       iscmagic.ERC721NFTCollectionAddress(collectionID),
			abi:           erc721NFTCollectionABI,
		},
	}
}

func (e *soloChainEnv) deployISCTestContract(creator *ecdsa.PrivateKey) *iscTestContractInstance {
	return &iscTestContractInstance{e.deployContract(creator, evmtest.ISCTestContractABI, evmtest.ISCTestContractBytecode)}
}

func (e *soloChainEnv) deployStorageContract(creator *ecdsa.PrivateKey) *storageContractInstance {
	return &storageContractInstance{e.deployContract(creator, evmtest.StorageContractABI, evmtest.StorageContractBytecode, uint32(42))}
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
	})
	require.NoError(e.t, err)

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, gasLimit, evm.GasPrice, data),
		e.signer(),
		creator,
	)
	require.NoError(e.t, err)

	err = e.evmChain.SendTransaction(tx)
	require.NoError(e.t, err)

	return &evmContractInstance{
		chain:         e,
		defaultSender: creator,
		address:       crypto.CreateAddress(creatorAddress, nonce),
		abi:           contractABI,
	}
}

func (e *soloChainEnv) registerERC20NativeToken(
	foundryOwner *cryptolib.KeyPair,
	foundrySN uint32,
	tokenName, tokenTickerSymbol string,
	tokenDecimals uint8,
) error {
	_, err := e.soloChain.PostRequestOffLedger(solo.NewCallParams(evm.Contract.Name, evm.FuncRegisterERC20NativeToken.Name, dict.Dict{
		evm.FieldFoundrySN:         codec.EncodeUint32(foundrySN),
		evm.FieldTokenName:         codec.EncodeString(tokenName),
		evm.FieldTokenTickerSymbol: codec.EncodeString(tokenTickerSymbol),
		evm.FieldTokenDecimals:     codec.EncodeUint8(tokenDecimals),
	}).WithMaxAffordableGasBudget(), foundryOwner)
	return err
}

func (e *soloChainEnv) registerERC20ExternalNativeToken(
	fromChain *solo.Chain,
	foundrySN uint32,
	tokenName, tokenTickerSymbol string,
	tokenDecimals uint8,
) (ret common.Address, err error) {
	_, err = fromChain.PostRequestOffLedger(solo.NewCallParams(evm.Contract.Name, evm.FuncRegisterERC20NativeTokenOnRemoteChain.Name, dict.Dict{
		evm.FieldFoundrySN:         codec.EncodeUint32(foundrySN),
		evm.FieldTokenName:         codec.EncodeString(tokenName),
		evm.FieldTokenTickerSymbol: codec.EncodeString(tokenTickerSymbol),
		evm.FieldTokenDecimals:     codec.EncodeUint8(tokenDecimals),
		evm.FieldTargetAddress:     codec.EncodeAddress(e.soloChain.ChainID.AsAddress()),
	}).WithMaxAffordableGasBudget().WithAllowance(isc.NewAssetsBaseTokens(1*isc.Million)), fromChain.OriginatorPrivateKey)
	if err != nil {
		return ret, err
	}

	foundryOutput, err := fromChain.GetFoundryOutput(foundrySN)
	require.NoError(e.t, err)
	nativeTokenID, err := foundryOutput.ID()
	require.NoError(e.t, err)

	if !e.soloChain.WaitUntil(func() bool {
		res, err2 := e.soloChain.CallView(evm.Contract.Name, evm.FuncGetERC20ExternalNativeTokenAddress.Name,
			evm.FieldNativeTokenID, nativeTokenID[:],
		)
		require.NoError(e.t, err2)
		if len(res[evm.FieldResult]) == 0 {
			return false
		}
		copy(ret[:], res[evm.FieldResult])
		return true
	}) {
		require.FailNow(e.t, "could not get ERC20 address on target chain")
	}
	return ret, err
}

func (e *soloChainEnv) registerERC721NFTCollection(collectionOwner *cryptolib.KeyPair, collectionID iotago.NFTID) error {
	_, err := e.soloChain.PostRequestOffLedger(solo.NewCallParams(evm.Contract.Name, evm.FuncRegisterERC721NFTCollection.Name, dict.Dict{
		evm.FieldNFTCollectionID: codec.EncodeNFTID(collectionID),
	}).WithMaxAffordableGasBudget(), collectionOwner)
	return err
}

func (e *evmContractInstance) callMsg(callMsg ethereum.CallMsg) ethereum.CallMsg {
	callMsg.To = &e.address
	return callMsg
}

func (e *evmContractInstance) parseEthCallOptions(opts []ethCallOptions, callData []byte) (ethCallOptions, error) {
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
	if opt.gasPrice == nil {
		opt.gasPrice = evm.GasPrice
	}
	if opt.gasLimit == 0 {
		var err error
		senderAddress := crypto.PubkeyToAddress(opt.sender.PublicKey)
		opt.gasLimit, err = e.chain.evmChain.EstimateGas(ethereum.CallMsg{
			From:     senderAddress,
			To:       &e.address,
			GasPrice: opt.gasPrice,
			Value:    opt.value,
			Data:     callData,
		})
		if err != nil {
			return opt, fmt.Errorf("error estimating gas limit: %w", e.chain.resolveError(err))
		}
	}
	return opt, nil
}

func (e *evmContractInstance) buildEthTx(opts []ethCallOptions, fnName string, args ...interface{}) (*types.Transaction, error) {
	callData, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	opt, err := e.parseEthCallOptions(opts, callData)
	if err != nil {
		return nil, err
	}

	senderAddress := crypto.PubkeyToAddress(opt.sender.PublicKey)

	nonce := e.chain.getNonce(senderAddress)

	unsignedTx := types.NewTransaction(nonce, e.address, opt.value, opt.gasLimit, opt.gasPrice, callData)

	return types.SignTx(unsignedTx, e.chain.signer(), opt.sender)
}

type callFnResult struct {
	tx         *types.Transaction
	evmReceipt *types.Receipt
	iscReceipt *isc.Receipt
}

func (e *evmContractInstance) estimateGas(opts []ethCallOptions, fnName string, args ...interface{}) (uint64, error) {
	tx, err := e.buildEthTx(opts, fnName, args...)
	if err != nil {
		return 0, err
	}
	return tx.Gas(), nil
}

func (e *evmContractInstance) callFn(opts []ethCallOptions, fnName string, args ...interface{}) (callFnResult, error) {
	e.chain.t.Logf("callFn: %s %+v", fnName, args)

	tx, err := e.buildEthTx(opts, fnName, args...)
	if err != nil {
		return callFnResult{}, err
	}
	res := callFnResult{tx: tx}

	sendTxErr := e.chain.evmChain.SendTransaction(res.tx)

	res.iscReceipt = e.chain.soloChain.LastReceipt()

	res.evmReceipt, err = e.chain.evmChain.TransactionReceipt(res.tx.Hash())
	require.NoError(e.chain.t, err)

	return res, sendTxErr
}

func (e *evmContractInstance) callFnExpectEvent(opts []ethCallOptions, eventName string, v interface{}, fnName string, args ...interface{}) callFnResult {
	res, err := e.callFn(opts, fnName, args...)
	require.NoError(e.chain.t, err)
	require.Equal(e.chain.t, types.ReceiptStatusSuccessful, res.evmReceipt.Status)
	require.Len(e.chain.t, res.evmReceipt.Logs, 1)
	if v != nil {
		err = e.abi.UnpackIntoInterface(v, eventName, res.evmReceipt.Logs[0].Data)
	}
	require.NoError(e.chain.t, err)
	return res
}

func (e *evmContractInstance) callView(fnName string, args []interface{}, v interface{}) error {
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
	if err != nil {
		return err
	}
	if v != nil {
		return e.abi.UnpackIntoInterface(v, fnName, ret)
	}
	return nil
}

func (i *iscTestContractInstance) getChainID() isc.ChainID {
	var v iscmagic.ISCChainID
	require.NoError(i.chain.t, i.callView("getChainID", nil, &v))
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
	require.NoError(s.chain.t, s.callView("retrieve", nil, &v))
	return v
}

func (s *storageContractInstance) store(n uint32, opts ...ethCallOptions) (res callFnResult, err error) {
	return s.callFn(opts, "store", n)
}

func (e *erc20ContractInstance) balanceOf(addr common.Address) *big.Int {
	v := new(big.Int)
	require.NoError(e.chain.t, e.callView("balanceOf", []interface{}{addr}, &v))
	return v
}

func (e *erc20ContractInstance) totalSupply() *big.Int {
	v := new(big.Int)
	require.NoError(e.chain.t, e.callView("totalSupply", nil, &v))
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
