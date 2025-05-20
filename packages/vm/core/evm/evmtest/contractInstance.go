package evmtest

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/evm/evmutil"
	"github.com/iotaledger/wasp/packages/isc"
)

type EVMContractInstance struct {
	chain         *SoloChainEnv
	defaultSender *ecdsa.PrivateKey
	address       common.Address
	abi           abi.ABI
}

type CallFnResult struct {
	tx         *types.Transaction
	EVMReceipt *types.Receipt
	ISCReceipt *isc.Receipt
}

type IscContractInstance struct {
	*EVMContractInstance
}

type iscTestContractInstance struct {
	*EVMContractInstance
}

type storageContractInstance struct {
	*EVMContractInstance
}

type erc20ContractInstance struct {
	*EVMContractInstance
}

type loopContractInstance struct {
	*EVMContractInstance
}

type fibonacciContractInstance struct {
	*EVMContractInstance
}

func (e *EVMContractInstance) callMsg(callMsg ethereum.CallMsg) ethereum.CallMsg {
	callMsg.To = &e.address
	return callMsg
}

func (e *EVMContractInstance) parseEthCallOptions(opts []ethCallOptions, callData []byte) (ethCallOptions, error) {
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

func (e *EVMContractInstance) buildEthTx(opts []ethCallOptions, fnName string, args ...interface{}) (*types.Transaction, error) {
	callData, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	opt, err := e.parseEthCallOptions(opts, callData)
	if err != nil {
		return nil, err
	}

	senderAddress := crypto.PubkeyToAddress(opt.sender.PublicKey)

	nonce := e.chain.getNonce(senderAddress)

	unsignedTx := types.NewTx(
		&types.LegacyTx{
			Nonce:    nonce,
			To:       &e.address,
			Value:    opt.value,
			Gas:      opt.gasLimit,
			GasPrice: opt.gasPrice,
			Data:     callData,
		},
	)

	return types.SignTx(unsignedTx, evmutil.Signer(big.NewInt(int64(e.chain.evmChain.ChainID()))), opt.sender)
}

func (e *EVMContractInstance) estimateGas(opts []ethCallOptions, fnName string, args ...interface{}) (uint64, error) {
	tx, err := e.buildEthTx(opts, fnName, args...)
	if err != nil {
		return 0, err
	}
	return tx.Gas(), nil
}

func (e *EVMContractInstance) MakeCallFn(opts []ethCallOptions, fnName string, args ...interface{}) (*types.Transaction, error) {
	return e.buildEthTx(opts, fnName, args...)
}

func (e *EVMContractInstance) CallFn(opts []ethCallOptions, fnName string, args ...interface{}) (CallFnResult, error) {
	e.chain.t.Logf("callFn: %s %+v", fnName, args)

	tx, err := e.buildEthTx(opts, fnName, args...)
	if err != nil {
		return CallFnResult{}, err
	}
	res := CallFnResult{tx: tx}

	sendTxErr := e.chain.evmChain.SendTransaction(res.tx)
	res.ISCReceipt = e.chain.Chain.LastReceipt()
	res.EVMReceipt = e.chain.evmChain.TransactionReceipt(res.tx.Hash())

	return res, sendTxErr
}

func (e *EVMContractInstance) CallFnExpectEvent(opts []ethCallOptions, eventName string, v interface{}, fnName string, args ...interface{}) CallFnResult {
	res, err := e.CallFn(opts, fnName, args...)
	require.NoError(e.chain.t, err)
	require.Equal(e.chain.t, types.ReceiptStatusSuccessful, res.EVMReceipt.Status)
	require.Len(e.chain.t, res.EVMReceipt.Logs, 1)
	if v != nil {
		err = e.abi.UnpackIntoInterface(v, eventName, res.EVMReceipt.Logs[0].Data)
	}
	require.NoError(e.chain.t, err)
	return res
}

func (e *EVMContractInstance) callView(fnName string, args []interface{}, v interface{}, blockNumberOrHash ...rpc.BlockNumberOrHash) error {
	e.chain.t.Logf("callView: %s %+v", fnName, args)
	callArguments, err := e.abi.Pack(fnName, args...)
	require.NoError(e.chain.t, err)
	var senderAddress common.Address
	if e.defaultSender != nil {
		senderAddress = crypto.PubkeyToAddress(e.defaultSender.PublicKey)
	}
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
