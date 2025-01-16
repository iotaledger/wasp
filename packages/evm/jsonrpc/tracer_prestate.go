// Code on this file adapted from
// https://github.com/ethereum/go-ethereum/blob/master/eth/tracers/native/prestate.go

package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/log"
)

func init() {
	registerTracer("prestateTracer", newPrestateTracer)
}

type PrestateAccountMap = map[common.Address]*PrestateAccount

type PrestateAccount struct {
	Balance *hexutil.Big                `json:"balance,omitempty"`
	Code    hexutil.Bytes               `json:"code,omitempty"`
	Nonce   uint64                      `json:"nonce,omitempty"`
	Storage map[common.Hash]common.Hash `json:"storage,omitempty"`
	empty   bool
}

type PrestateDiffResult struct {
	Post PrestateAccountMap `json:"post"`
	Pre  PrestateAccountMap `json:"pre"`
}

func (a *PrestateAccount) exists() bool {
	return a.Nonce > 0 || len(a.Code) > 0 || len(a.Storage) > 0 || (a.Balance != nil && a.Balance.ToInt().Sign() != 0)
}

type PrestateTxValue struct {
	Pre     PrestateAccountMap `json:"pre"`
	Post    PrestateAccountMap `json:"post"`
	created map[common.Address]bool
	deleted map[common.Address]bool
	to      common.Address
}

type prestateTracer struct {
	env           *tracing.VMContext
	currentTxHash common.Hash
	states        map[common.Hash]*PrestateTxValue // key is the tx hash, value is the state diff
	config        prestateTracerConfig
	interrupt     atomic.Bool // Atomic flag to signal execution interruption
	reason        error       // Textual reason for the interruption
	traceBlock    bool
	blockTxs      types.Transactions
}

type prestateTracerConfig struct {
	DiffMode bool `json:"diffMode"` // If true, this tracer will return state modifications
}

func newPrestateTracer(ctx *tracers.Context, cfg json.RawMessage, traceBlock bool, initValue any) (*Tracer, error) {
	var blockTxs types.Transactions

	if initValue == nil && traceBlock {
		return nil, fmt.Errorf("initValue with block transactions is required for block tracing")
	}

	if initValue != nil {
		var ok bool
		blockTxs, ok = initValue.(types.Transactions)
		if !ok {
			return nil, fmt.Errorf("invalid init value type for prestateTracer: %T", initValue)
		}
	}
	var config prestateTracerConfig
	if err := json.Unmarshal(cfg, &config); err != nil {
		return nil, err
	}
	t := &prestateTracer{
		config:     config,
		traceBlock: traceBlock,
		states:     make(map[common.Hash]*PrestateTxValue),
		blockTxs:   blockTxs,
	}
	return &Tracer{
		Tracer: &tracers.Tracer{
			Hooks: &tracing.Hooks{
				OnTxStart: t.OnTxStart,
				OnTxEnd:   t.OnTxEnd,
				OnOpcode:  t.OnOpcode,
			},
			GetResult: t.GetResult,
			Stop:      t.Stop,
		},
		TraceFakeTx: t.TraceFakeTx,
	}, nil
}

// OnOpcode implements the EVMLogger interface to trace a single step of VM execution.
//
//nolint:gocyclo
func (t *prestateTracer) OnOpcode(pc uint64, opcode byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
	if err != nil {
		return
	}
	// Skip if tracing was interrupted
	if t.interrupt.Load() {
		return
	}
	op := vm.OpCode(opcode)
	stackData := scope.StackData()
	stackLen := len(stackData)
	caller := scope.Address()
	switch {
	case stackLen >= 1 && (op == vm.SLOAD || op == vm.SSTORE):
		slot := common.Hash(stackData[stackLen-1].Bytes32())
		t.lookupStorage(t.currentTxHash, caller, slot)
	case stackLen >= 1 && (op == vm.EXTCODECOPY || op == vm.EXTCODEHASH || op == vm.EXTCODESIZE || op == vm.BALANCE || op == vm.SELFDESTRUCT):
		addr := common.Address(stackData[stackLen-1].Bytes20())
		t.lookupAccount(t.currentTxHash, addr)
		if op == vm.SELFDESTRUCT {
			t.states[t.currentTxHash].deleted[caller] = true
		}
	case stackLen >= 5 && (op == vm.DELEGATECALL || op == vm.CALL || op == vm.STATICCALL || op == vm.CALLCODE):
		addr := common.Address(stackData[stackLen-2].Bytes20())
		t.lookupAccount(t.currentTxHash, addr)
	case op == vm.CREATE:
		nonce := t.env.StateDB.GetNonce(caller)
		addr := crypto.CreateAddress(caller, nonce)
		t.lookupAccount(t.currentTxHash, addr)
		t.states[t.currentTxHash].created[addr] = true
	case stackLen >= 4 && op == vm.CREATE2:
		offset := stackData[stackLen-2]
		size := stackData[stackLen-3]
		init, err := getMemoryCopyPadded(scope.MemoryData(), int64(offset.Uint64()), int64(size.Uint64()))
		if err != nil {
			log.Warn("failed to copy CREATE2 input", "err", err, "tracer", "prestateTracer", "offset", offset, "size", size)
			return
		}
		inithash := crypto.Keccak256(init)
		salt := stackData[stackLen-4]
		addr := crypto.CreateAddress2(caller, salt.Bytes32(), inithash)
		t.lookupAccount(t.currentTxHash, addr)
		t.states[t.currentTxHash].created[addr] = true
	}
}

func (t *prestateTracer) OnTxStart(env *tracing.VMContext, tx *types.Transaction, from common.Address) {
	t.env = env
	t.currentTxHash = tx.Hash()

	t.states[tx.Hash()] = &PrestateTxValue{
		Pre:     make(PrestateAccountMap),
		Post:    make(PrestateAccountMap),
		created: make(map[common.Address]bool),
		deleted: make(map[common.Address]bool),
	}

	txState := t.states[tx.Hash()]

	if tx.To() == nil {
		createdAddr := crypto.CreateAddress(from, env.StateDB.GetNonce(from))
		txState.to = createdAddr
		txState.created[createdAddr] = true
	} else {
		txState.to = *tx.To()
	}

	t.lookupAccount(tx.Hash(), from)
	t.lookupAccount(tx.Hash(), txState.to)
	t.lookupAccount(tx.Hash(), env.Coinbase)
}

func (t *prestateTracer) OnTxEnd(receipt *types.Receipt, err error) {
	if err != nil {
		return
	}
	if t.config.DiffMode {
		t.processDiffState()
	}
	// the new created contracts' prestate were empty, so delete them
	for a := range t.states[t.currentTxHash].created {
		// the created contract maybe exists in statedb before the creating tx
		if s := t.states[t.currentTxHash].Pre[a]; s != nil && s.empty {
			delete(t.states[t.currentTxHash].Pre, a)
		}
	}
}

// GetResult returns the json-encoded nested list of call traces, and any
// error arising from the encoding or forceful termination (via `Stop`).
func (t *prestateTracer) GetResult() (json.RawMessage, error) {
	return GetTraceResults(
		t.blockTxs,
		t.traceBlock,
		t.TraceFakeTx,
		func(tx *types.Transaction) (json.RawMessage, error) {
			txState := t.states[tx.Hash()]
			if t.config.DiffMode {
				return json.Marshal(PrestateDiffResult{txState.Post, txState.Pre})
			}
			return json.Marshal(txState.Pre)
		}, func() (json.RawMessage, error) {
			if t.config.DiffMode {
				return json.Marshal(PrestateDiffResult{t.states[t.currentTxHash].Post, t.states[t.currentTxHash].Pre})
			}
			return json.Marshal(t.states[t.currentTxHash].Pre)
		}, t.reason)
}

// Stop terminates execution of the tracer at the first opportune moment.
func (t *prestateTracer) Stop(err error) {
	t.reason = err
	t.interrupt.Store(true)
}

func (t *prestateTracer) processDiffState() {
	txState := t.states[t.currentTxHash]
	for addr, state := range txState.Pre {
		// The deleted account's state is pruned from `post` but kept in `pre`
		if _, ok := txState.deleted[addr]; ok {
			continue
		}
		modified := false
		postAccount := &PrestateAccount{Storage: make(map[common.Hash]common.Hash)}
		newBalance := t.env.StateDB.GetBalance(addr).ToBig()
		newNonce := t.env.StateDB.GetNonce(addr)
		newCode := t.env.StateDB.GetCode(addr)

		if newBalance.Cmp(txState.Pre[addr].Balance.ToInt()) != 0 {
			modified = true
			postAccount.Balance = (*hexutil.Big)(newBalance)
		}
		if newNonce != txState.Pre[addr].Nonce {
			modified = true
			postAccount.Nonce = newNonce
		}
		if !bytes.Equal(newCode, txState.Pre[addr].Code) {
			modified = true
			postAccount.Code = newCode
		}

		for key, val := range state.Storage {
			// don't include the empty slot
			if val == (common.Hash{}) {
				delete(txState.Pre[addr].Storage, key)
			}

			newVal := t.env.StateDB.GetState(addr, key)
			if val == newVal {
				// Omit unchanged slots
				delete(txState.Pre[addr].Storage, key)
			} else {
				modified = true
				if newVal != (common.Hash{}) {
					postAccount.Storage[key] = newVal
				}
			}
		}

		if modified {
			txState.Post[addr] = postAccount
		} else {
			// if state is not modified, then no need to include into the pre state
			delete(txState.Pre, addr)
		}
	}
}

// lookupAccount fetches details of an account and adds it to the prestate
// if it doesn't exist there.
func (t *prestateTracer) lookupAccount(tx common.Hash, addr common.Address) {
	if _, ok := t.states[tx].Pre[addr]; ok {
		return
	}

	acc := &PrestateAccount{
		Balance: (*hexutil.Big)(t.env.StateDB.GetBalance(addr).ToBig()),
		Nonce:   t.env.StateDB.GetNonce(addr),
		Code:    t.env.StateDB.GetCode(addr),
		Storage: make(map[common.Hash]common.Hash),
	}
	if !acc.exists() {
		acc.empty = true
	}

	t.states[tx].Pre[addr] = acc
}

// lookupStorage fetches the requested storage slot and adds
// it to the prestate of the given contract. It assumes `lookupAccount`
// has been performed on the contract before.
func (t *prestateTracer) lookupStorage(tx common.Hash, addr common.Address, key common.Hash) {
	if _, ok := t.states[tx].Pre[addr].Storage[key]; ok {
		return
	}
	t.states[tx].Pre[addr].Storage[key] = t.env.StateDB.GetState(addr, key)
}

func (t *prestateTracer) TraceFakeTx(tx *types.Transaction) (res json.RawMessage, err error) {
	if t.config.DiffMode {
		res, err = json.Marshal(PrestateDiffResult{
			Post: PrestateAccountMap{},
			Pre:  PrestateAccountMap{},
		})
	} else {
		res, err = json.Marshal(PrestateAccountMap{})
	}
	return res, err
}
