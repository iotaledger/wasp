// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmutil"
)

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        *common.Hash    `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex *hexutil.Uint64 `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}

// RPCMarshalHeader converts the given header to the RPC output .
func RPCMarshalHeader(head *types.Header) map[string]any {
	return map[string]any{
		"number":           (*hexutil.Big)(head.Number),
		"hash":             head.Hash(),
		"parentHash":       head.ParentHash,
		"nonce":            head.Nonce,
		"mixHash":          head.MixDigest,
		"sha3Uncles":       head.UncleHash,
		"logsBloom":        head.Bloom,
		"stateRoot":        head.Root,
		"miner":            head.Coinbase,
		"difficulty":       (*hexutil.Big)(head.Difficulty),
		"extraData":        hexutil.Bytes(head.Extra),
		"size":             hexutil.Uint64(head.Size()),
		"gasLimit":         hexutil.Uint64(head.GasLimit),
		"gasUsed":          hexutil.Uint64(head.GasUsed),
		"timestamp":        hexutil.Uint64(head.Time),
		"transactionsRoot": head.TxHash,
		"receiptsRoot":     head.ReceiptHash,
	}
}

// RPCMarshalBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func RPCMarshalBlock(block *types.Block, inclTx, fullTx bool) (map[string]any, error) {
	fields := RPCMarshalHeader(block.Header())
	fields["size"] = hexutil.Uint64(block.Size())
	fields["totalDifficulty"] = hexutil.Uint64(0)

	if inclTx {
		formatTx := func(tx *types.Transaction) (any, error) {
			return tx.Hash(), nil
		}
		if fullTx {
			formatTx = func(tx *types.Transaction) (any, error) {
				return newRPCTransactionFromBlockHash(block, tx.Hash()), nil
			}
		}
		txs := block.Transactions()
		transactions := make([]any, len(txs))
		var err error
		for i, tx := range txs {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}
	uncles := block.Uncles()
	uncleHashes := make([]common.Hash, len(uncles))
	for i, uncle := range uncles {
		uncleHashes[i] = uncle.Hash()
	}
	fields["uncles"] = uncleHashes

	return fields, nil
}

// newRPCTransactionFromBlockHash returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockHash(b *types.Block, hash common.Hash) *RPCTransaction {
	for idx, tx := range b.Transactions() {
		if tx.Hash() == hash {
			return newRPCTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}

// newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockIndex(b *types.Block, index uint64) *RPCTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), index)
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *types.Transaction, blockHash common.Hash, blockNumber, index uint64) *RPCTransaction {
	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	result := &RPCTransaction{
		From:     from,
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = &blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = (*hexutil.Uint64)(&index)
	}
	return result
}

func parseBlockNumber(bn rpc.BlockNumber) *big.Int {
	n := bn.Int64()
	if n < 0 {
		// for our uses there is no "pending" block, so always assume "latest"
		return nil
	}
	return big.NewInt(n)
}

func RPCMarshalReceipt(r *types.Receipt, tx *types.Transaction, effectiveGasPrice *big.Int) map[string]any {
	// fix for an already fixed bug where some old failed receipts contain non-empty logs
	if r.Status != types.ReceiptStatusSuccessful {
		r.Logs = []*types.Log{}
		r.Bloom = types.CreateBloom(r)
	}

	result := map[string]any{
		"transactionHash":   r.TxHash,
		"transactionIndex":  hexutil.Uint64(r.TransactionIndex),
		"blockHash":         r.BlockHash,
		"blockNumber":       (*hexutil.Big)(r.BlockNumber),
		"from":              evmutil.MustGetSenderIfTxSigned(tx),
		"to":                tx.To(),
		"cumulativeGasUsed": hexutil.Uint64(r.CumulativeGasUsed),
		"gasUsed":           hexutil.Uint64(r.GasUsed),
		"effectiveGasPrice": hexutil.EncodeBig(effectiveGasPrice),
		"logs":              rpcMarshalLogs(r),
		"logsBloom":         r.Bloom,
		"status":            hexutil.Uint64(r.Status),
		"type":              hexutil.Uint64(types.LegacyTxType),
	}

	// Eth compatibility. Return "null" instead of "0x00000000000000000000000..."
	if slices.Equal(r.ContractAddress.Bytes(), common.Address{}.Bytes()) {
		result["contractAddress"] = nil
	} else {
		result["contractAddress"] = r.ContractAddress
	}

	return result
}

func rpcMarshalLogs(r *types.Receipt) []any {
	ret := make([]any, len(r.Logs))
	for i, log := range r.Logs {
		ret[i] = map[string]any{
			"logIndex":         hexutil.Uint(log.Index),
			"blockNumber":      hexutil.Uint(log.BlockNumber),
			"blockHash":        log.BlockHash,
			"transactionHash":  log.TxHash,
			"transactionIndex": hexutil.Uint(log.TxIndex),
			"address":          log.Address,
			"data":             hexutil.Bytes(log.Data),
			"topics":           log.Topics,
			"removed":          log.Removed,
		}
	}
	return ret
}

type RPCCallArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred by clients.
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`
}

func (c *RPCCallArgs) parse() (ret ethereum.CallMsg) {
	ret.From = c.From
	ret.To = c.To
	if c.Gas != nil {
		ret.Gas = uint64(*c.Gas)
	}
	ret.GasPrice = (*big.Int)(c.GasPrice)
	ret.Value = (*big.Int)(c.Value)
	if c.Data != nil {
		ret.Data = *c.Data
	}
	if c.Input != nil {
		ret.Data = *c.Input
	}
	return
}

// SendTxArgs represents the arguments to submit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Nonce    *hexutil.Uint64 `json:"nonce"`
	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred by clients.
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`
}

// setDefaults is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(e *EthService) error {
	if args.GasPrice == nil {
		args.GasPrice = (*hexutil.Big)(e.evmChain.GasPrice())
	}
	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Nonce == nil {
		nonce, err := e.evmChain.TransactionCount(args.From, nil)
		if err != nil {
			return err
		}
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}
	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return errors.New(`both "data" and "input" are set and not equal. Please use "input" to pass transaction call data`)
	}
	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}
		if len(input) == 0 {
			return errors.New(`contract creation without any data provided`)
		}
	}
	return nil
}

func (args *SendTxArgs) toTransaction() *types.Transaction {
	var input []byte
	if args.Input != nil {
		input = *args.Input
	} else if args.Data != nil {
		input = *args.Data
	}
	if args.To == nil {
		return types.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
	}
	return types.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
}

type RPCFilterQuery ethereum.FilterQuery

// UnmarshalJSON sets *args fields with given data.
//

//nolint:gocyclo
func (q *RPCFilterQuery) UnmarshalJSON(data []byte) error {
	type input struct {
		BlockHash *common.Hash     `json:"blockHash"`
		FromBlock *rpc.BlockNumber `json:"fromBlock"`
		ToBlock   *rpc.BlockNumber `json:"toBlock"`
		Addresses any              `json:"address"`
		Topics    []any            `json:"topics"`
	}

	var raw input
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.BlockHash != nil {
		if raw.FromBlock != nil || raw.ToBlock != nil {
			// BlockHash is mutually exclusive with FromBlock/ToBlock criteria
			return fmt.Errorf("cannot specify both BlockHash and FromBlock/ToBlock, choose one or the other")
		}
		q.BlockHash = raw.BlockHash
	} else {
		if raw.FromBlock != nil {
			q.FromBlock = big.NewInt(raw.FromBlock.Int64())
		}

		if raw.ToBlock != nil {
			q.ToBlock = big.NewInt(raw.ToBlock.Int64())
		}
	}

	q.Addresses = []common.Address{}

	if raw.Addresses != nil {
		// raw.Address can contain a single address or an array of addresses
		switch rawAddr := raw.Addresses.(type) {
		case []any:
			for i, addr := range rawAddr {
				if strAddr, ok := addr.(string); ok {
					addr, err := decodeAddress(strAddr)
					if err != nil {
						return fmt.Errorf("invalid address at index %d: %w", i, err)
					}
					q.Addresses = append(q.Addresses, addr)
				} else {
					return fmt.Errorf("non-string address at index %d", i)
				}
			}
		case string:
			addr, err := decodeAddress(rawAddr)
			if err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}
			q.Addresses = []common.Address{addr}
		default:
			return errors.New("invalid addresses in query")
		}
	}

	// topics is an array consisting of strings and/or arrays of strings.
	// JSON null values are converted to common.Hash{} and ignored by the filter manager.
	if len(raw.Topics) > 0 {
		q.Topics = make([][]common.Hash, len(raw.Topics))
		for i, t := range raw.Topics {
			switch topic := t.(type) {
			case nil:
				// ignore topic when matching logs

			case string:
				// match specific topic
				top, err := decodeTopic(topic)
				if err != nil {
					return err
				}
				q.Topics[i] = []common.Hash{top}

			case []any:
				// or case e.g. [null, "topic0", "topic1"]
				for _, rawTopic := range topic {
					if rawTopic == nil {
						// null component, match all
						q.Topics[i] = nil
						break
					}
					if topic, ok := rawTopic.(string); ok {
						parsed, err := decodeTopic(topic)
						if err != nil {
							return err
						}
						q.Topics[i] = append(q.Topics[i], parsed)
					} else {
						return errors.New("invalid topic(s)")
					}
				}
			default:
				return errors.New("invalid topic(s)")
			}
		}
	}

	return nil
}

func decodeAddress(s string) (common.Address, error) {
	b, err := cryptolib.DecodeHex(s)
	if err == nil && len(b) != common.AddressLength {
		err = fmt.Errorf("hex has invalid length %d after decoding; expected %d for address", len(b), common.AddressLength)
	}
	return common.BytesToAddress(b), err
}

func decodeTopic(s string) (common.Hash, error) {
	b, err := cryptolib.DecodeHex(s)
	if err == nil && len(b) != common.HashLength {
		err = fmt.Errorf("hex has invalid length %d after decoding; expected %d for topic", len(b), common.HashLength)
	}
	return common.BytesToHash(b), err
}

type revertError struct {
	error
	reason string // revert reason hex encoded
}

// ErrorCode returns the JSON error code for a revertal.
// See: https://github.com/ethereum/wiki/wiki/JSON-RPC-Error-Codes-Improvement-Proposal
func (e *revertError) ErrorCode() int {
	return 3
}

// ErrorData returns the hex encoded revert reason.
func (e *revertError) ErrorData() any {
	return e.reason
}

func newRevertError(revertData []byte) *revertError {
	reason, errUnpack := abi.UnpackRevert(revertData)
	err := errors.New("execution reverted")
	if errUnpack == nil {
		err = fmt.Errorf("execution reverted: %v", reason)
	}
	return &revertError{
		error:  err,
		reason: hexutil.Encode(revertData),
	}
}
