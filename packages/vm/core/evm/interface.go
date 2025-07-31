// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm/evmnames"
)

var Contract = coreutil.NewContract(evmnames.Contract)

var (
	// FuncSendTransaction is the main entry point, called by an
	// evmOffLedgerTxRequest in order to process an Ethereum tx (e.g.
	// eth_sendRawTransaction).
	FuncSendTransaction = coreutil.NewEP1(Contract, evmnames.FuncSendTransaction,
		coreutil.Field[*types.Transaction]("transaction"),
	)

	// FuncCallContract is the entry point called by an evmOffLedgerCallRequest
	// in order to process a view call or gas estimation (e.g. eth_call, eth_estimateGas).
	FuncCallContract = coreutil.NewEP11(Contract, evmnames.FuncCallContract,
		coreutil.Field[ethereum.CallMsg]("callMessage"),
		coreutil.Field[[]byte]("functionResult"),
	)

	FuncRegisterERC20Coin = coreutil.NewEP1(Contract, evmnames.FuncRegisterERC20Coin,
		coreutil.Field[coin.Type]("coinType"),
	)
	FuncNewL1Deposit = coreutil.NewEP3(Contract, evmnames.FuncNewL1Deposit,
		coreutil.Field[isc.AgentID]("l1DepositOriginatorAgentID"),
		coreutil.Field[common.Address]("targetAddress"),
		coreutil.Field[*isc.Assets]("assets"),
	)
	ViewGetChainID = coreutil.NewViewEP01(Contract, evmnames.ViewGetChainID,
		coreutil.Field[uint16]("chainID"),
	)
)

const (
	DefaultChainID = uint16(1074) // IOTA -- get it?
)
