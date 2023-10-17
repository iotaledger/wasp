package evmtest

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/stretchr/testify/require"
	"math/big"
)

type evmContractInstance struct {
	chain         *SoloChainEnv
	defaultSender *ecdsa.PrivateKey
	address       common.Address
	abi           abi.ABI
}

type callFnResult struct {
	tx         *types.Transaction
	evmReceipt *types.Receipt
	iscReceipt *isc.Receipt
}

type IscContractInstance struct {
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
		opt.gasPrice = e.chain.evmChain.GasPrice()
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
		}, nil)
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
	res.evmReceipt = e.chain.evmChain.TransactionReceipt(res.tx.Hash())

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

func (e *evmContractInstance) callView(fnName string, args []interface{}, v interface{}, blockNumberOrHash ...rpc.BlockNumberOrHash) error {
	e.chain.t.Logf("callView: %s %+v", fnName, args)
	callArguments, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	senderAddress := crypto.PubkeyToAddress(e.defaultSender.PublicKey)
	callMsg := e.callMsg(ethereum.CallMsg{
		From: senderAddress,
		Data: callArguments,
	})
	var bn *rpc.BlockNumberOrHash
	if len(blockNumberOrHash) > 0 {
		bn = &blockNumberOrHash[0]
	}
	ret, err := e.chain.evmChain.CallContract(callMsg, bn)
	if err != nil {
		return err
	}
	if v != nil {
		return e.abi.UnpackIntoInterface(v, fnName, ret)
	}
	return nil
}
