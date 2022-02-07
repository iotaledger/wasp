// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/contracts/native/evm/evmlight/iscptest"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmflavors"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/stretchr/testify/require"
)

func withEVMFlavors(t *testing.T, f func(*testing.T, *coreutil.ContractInfo)) {
	for _, evmFlavor := range evmflavors.All {
		t.Run(evmFlavor.Name, func(t *testing.T) { f(t, evmFlavor) })
	}
}

type evmChainInstance struct {
	t            testing.TB
	evmFlavor    *coreutil.ContractInfo
	solo         *solo.Solo
	soloChain    *solo.Chain
	faucetKey    *ecdsa.PrivateKey
	faucetSupply *big.Int
	chainID      int
}

type evmContractInstance struct {
	chain   *evmChainInstance
	creator *ecdsa.PrivateKey
	address common.Address
	abi     abi.ABI
}

type iscpTestContractInstance struct {
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
	wallet   *cryptolib.KeyPair
	transfer uint64
}

type ethCallOptions struct {
	iota     iotaCallOptions
	sender   *ecdsa.PrivateKey
	value    *big.Int
	gasLimit uint64
}

func initEVMChain(t testing.TB, evmFlavor *coreutil.ContractInfo, nativeContracts ...*coreutil.ContractProcessor) *evmChainInstance {
	env := solo.New(t, &solo.InitOptions{
		AutoAdjustDustDeposit: true,
		Debug:                 true,
		PrintStackTrace:       true,
	}).
		WithNativeContract(evmflavors.Processors[evmFlavor.Name])
	for _, c := range nativeContracts {
		env = env.WithNativeContract(c)
	}
	return initEVMChainWithSolo(t, evmFlavor, env)
}

func initEVMChainWithSolo(t testing.TB, evmFlavor *coreutil.ContractInfo, env *solo.Solo) *evmChainInstance {
	faucetKey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	chainID := evm.DefaultChainID
	e := &evmChainInstance{
		t:            t,
		evmFlavor:    evmFlavor,
		solo:         env,
		soloChain:    env.NewChain(nil, "ch1"),
		faucetKey:    faucetKey,
		faucetSupply: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9)),
		chainID:      chainID,
	}
	err := e.soloChain.DeployContract(nil, evmFlavor.Name, evmFlavor.ProgramHash,
		evm.FieldChainID, codec.EncodeUint16(uint16(chainID)),
		evm.FieldGenesisAlloc, evmtypes.EncodeGenesisAlloc(map[common.Address]core.GenesisAccount{
			e.faucetAddress(): {Balance: e.faucetSupply},
		}),
	)
	require.NoError(e.t, err)
	return e
}

func (e *evmChainInstance) parseIotaCallOptions(opts []iotaCallOptions) iotaCallOptions {
	if len(opts) == 0 {
		opts = []iotaCallOptions{{}}
	}
	opt := opts[0]
	if opt.wallet == nil {
		opt.wallet = &e.soloChain.OriginatorPrivateKey
	}
	if opt.transfer == 0 {
		opt.transfer = 1
	}
	return opt
}

func (e *evmChainInstance) buildSoloRequest(funName string, transfer uint64, params ...interface{}) *solo.CallParams {
	return solo.NewCallParams(e.evmFlavor.Name, funName, params...).AddAssetsIotas(transfer)
}

func (e *evmChainInstance) postRequest(opts []iotaCallOptions, funName string, params ...interface{}) (dict.Dict, error) {
	opt := e.parseIotaCallOptions(opts)
	req := e.buildSoloRequest(funName, opt.transfer, params...)
	gas, gasFee, err := e.soloChain.EstimateGasOnLedger(req, opt.wallet, true)
	if err != nil {
		return nil, err
	}
	return e.soloChain.PostRequestSync(
		req.WithGasBudget(gas).AddAssetsIotas(gasFee),
		opt.wallet,
	)
}

func (e *evmChainInstance) callView(funName string, params ...interface{}) (dict.Dict, error) {
	return e.soloChain.CallView(e.evmFlavor.Name, funName, params...)
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

func (e *evmChainInstance) getGasPerIotas() uint64 {
	ret, err := e.callView(evm.FuncGetGasPerIota.Name)
	require.NoError(e.t, err)
	gasPerIotas, err := codec.DecodeUint64(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return gasPerIotas
}

func (e *evmChainInstance) setGasPerIotas(newGasPerIota uint64, opts ...iotaCallOptions) error {
	_, err := e.postRequest(opts, evm.FuncSetGasPerIota.Name, evm.FieldGasPerIota, newGasPerIota)
	return err
}

func (e *evmChainInstance) claimOwnership(opts ...iotaCallOptions) error {
	_, err := e.postRequest(opts, evm.FuncClaimOwnership.Name)
	return err
}

func (e *evmChainInstance) setNextOwner(nextOwner *iscp.AgentID, opts ...iotaCallOptions) error {
	_, err := e.postRequest(opts, evm.FuncSetNextOwner.Name, evm.FieldNextEVMOwner, nextOwner)
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

func (e *evmChainInstance) getOwner() *iscp.AgentID {
	ret, err := e.callView(evm.FuncGetOwner.Name)
	require.NoError(e.t, err)
	owner, err := codec.DecodeAgentID(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return owner
}

func (e *evmChainInstance) deployISCPTestContract(creator *ecdsa.PrivateKey) *iscpTestContractInstance {
	return &iscpTestContractInstance{e.deployContract(creator, iscptest.ISCPTestContractABI, iscptest.ISCPTestContractBytecode)}
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

	// estimate EVM gas
	evmGas := e.estimateGas(ethereum.CallMsg{
		From:     creatorAddress,
		GasPrice: evm.GasPrice,
		Value:    value,
		Data:     data,
	})

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, evmGas, evm.GasPrice, data),
		e.signer(),
		e.faucetKey,
	)
	require.NoError(e.t, err)

	txdata, err := tx.MarshalBinary()
	require.NoError(e.t, err)

	// estimate ISCP gas
	req := solo.NewCallParamsFromDic(e.evmFlavor.Name, evm.FuncSendTransaction.Name, dict.Dict{
		evm.FieldTransactionData: txdata,
	})
	iscpGas, iscpGasFee, err := e.soloChain.EstimateGasOnLedger(req, nil, true)

	// send EVM tx
	_, err = e.soloChain.PostRequestSync(
		solo.NewCallParamsFromDic(e.evmFlavor.Name, evm.FuncSendTransaction.Name, dict.Dict{
			evm.FieldTransactionData: txdata,
		}).
			WithGasBudget(iscpGas).
			AddAssetsIotas(evmGas/e.getGasPerIotas()+iscpGasFee),
		nil,
	)
	require.NoError(e.t, err)

	return &evmContractInstance{
		chain:   e,
		creator: creator,
		address: crypto.CreateAddress(creatorAddress, nonce),
		abi:     contractABI,
	}
}

func (e *evmChainInstance) estimateGas(callMsg ethereum.CallMsg) uint64 {
	ret, err := e.callView(evm.FuncEstimateGas.Name, evm.FieldCallMsg, evmtypes.EncodeCallMsg(callMsg))
	if err != nil {
		e.t.Logf("%v", err)
		return evm.GasLimitDefault - 1
	}
	gas, err := codec.DecodeUint64(ret.MustGet(evm.FieldResult))
	require.NoError(e.t, err)
	return gas
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
		opt.sender = e.creator
	}
	if opt.iota.wallet == nil {
		opt.iota.wallet = &e.chain.soloChain.OriginatorPrivateKey
	}
	if opt.gasLimit == 0 {
		opt.gasLimit = e.chain.estimateGas(e.callMsg(ethereum.CallMsg{
			From: crypto.PubkeyToAddress(opt.sender.PublicKey),
			Data: callData,
		}))
	}
	if opt.iota.transfer == 0 {
		opt.iota.transfer = opt.gasLimit / e.chain.getGasPerIotas()
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

	unsignedTx := types.NewTransaction(nonce, e.address, opt.value, opt.gasLimit, evm.GasPrice, callArguments)

	tx, err := types.SignTx(unsignedTx, e.chain.signer(), opt.sender)
	require.NoError(e.chain.t, err)

	txdata, err := tx.MarshalBinary()
	require.NoError(e.chain.t, err)
	return txdata, tx, opt
}

type callFnResult struct {
	tx                 *types.Transaction
	evmReceipt         *types.Receipt
	evmChargedFeeIotas uint64
	iscpReceipt        *blocklog.RequestReceipt
}

func (e *evmContractInstance) callFn(opts []ethCallOptions, fnName string, args ...interface{}) (res callFnResult, err error) {
	e.chain.t.Logf("callFn: %s %+v", fnName, args)

	txdata, tx, opt := e.buildEthTxData(opts, fnName, args...)
	res.tx = tx

	result, err := e.chain.postRequest([]iotaCallOptions{opt.iota}, evm.FuncSendTransaction.Name, evm.FieldTransactionData, txdata)
	if err != nil {
		return
	}

	res.iscpReceipt = e.chain.soloChain.LastReceipt()

	res.evmChargedFeeIotas, err = codec.DecodeUint64(result.MustGet(evm.FieldGasFee))
	require.NoError(e.chain.t, err)

	gasUsed, err := codec.DecodeUint64(result.MustGet(evm.FieldGasUsed))
	require.NoError(e.chain.t, err)

	receiptResult, err := e.chain.callView(evm.FuncGetReceipt.Name, evm.FieldTransactionHash, tx.Hash().Bytes())
	require.NoError(e.chain.t, err)

	res.evmReceipt, err = evmtypes.DecodeReceiptFull(receiptResult.MustGet(evm.FieldResult))
	require.NoError(e.chain.t, err)

	require.EqualValues(e.chain.t, res.evmReceipt.GasUsed, gasUsed)
	return
}

func (e *evmContractInstance) callView(opts []ethCallOptions, fnName string, args []interface{}, v interface{}) {
	e.chain.t.Logf("callView: %s %+v", fnName, args)
	callArguments, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	opt := e.parseEthCallOptions(opts, callArguments)
	senderAddress := crypto.PubkeyToAddress(opt.sender.PublicKey)
	callMsg := e.callMsg(ethereum.CallMsg{
		From:     senderAddress,
		Gas:      opt.gasLimit,
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

func (i *iscpTestContractInstance) getChainID() *iscp.ChainID {
	var v [iscp.ChainIDLength]byte
	i.callView(nil, "getChainID", nil, &v)
	chainID, err := iscp.ChainIDFromBytes(v[:])
	require.NoError(i.chain.t, err)
	return chainID
}

func (i *iscpTestContractInstance) triggerEvent(s string) (res callFnResult, err error) {
	return i.callFn(nil, "triggerEvent", s)
}

func (i *iscpTestContractInstance) emitEntropy() (res callFnResult, err error) {
	return i.callFn(nil, "emitEntropy")
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
