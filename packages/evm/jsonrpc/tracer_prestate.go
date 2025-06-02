// Code on this file adapted from
// https://github.com/ethereum/go-ethereum/blob/master/eth/tracers/native/prestate.go

package jsonrpc

import (
	"bytes"
	"encoding/json"
	"sync/atomic"

	"fortio.org/safecast"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/log"
	"github.com/samber/lo"
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

type prestateTxTrace struct {
	txHash  common.Hash
	env     *tracing.VMContext
	pre     PrestateAccountMap
	post    PrestateAccountMap
	to      common.Address
	created map[common.Address]bool
	deleted map[common.Address]bool
}

type prestateTracer struct {
	txTraces  []*prestateTxTrace
	config    prestateTracerConfig
	interrupt atomic.Bool // Atomic flag to signal execution interruption
	reason    error       // Textual reason for the interruption
}

type prestateTracerConfig struct {
	DiffMode bool `json:"diffMode"` // If true, this tracer will return state modifications
}

func newPrestateTracer(
	ctx *tracers.Context,
	cfg json.RawMessage,
) (*tracers.Tracer, error) {
	var config prestateTracerConfig
	if err := json.Unmarshal(cfg, &config); err != nil {
		return nil, err
	}
	t := &prestateTracer{
		config: config,
	}
	return &tracers.Tracer{
		Hooks: &tracing.Hooks{
			OnTxStart: t.OnTxStart,
			OnTxEnd:   t.OnTxEnd,
			OnOpcode:  t.OnOpcode,
		},
		GetResult: t.GetResult,
		Stop:      t.Stop,
	}, nil
}

func (t *prestateTracer) currentTxTrace() *prestateTxTrace {
	return t.txTraces[len(t.txTraces)-1]
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
		t.lookupStorage(caller, slot)
	case stackLen >= 1 && (op == vm.EXTCODECOPY || op == vm.EXTCODEHASH || op == vm.EXTCODESIZE || op == vm.BALANCE || op == vm.SELFDESTRUCT):
		addr := common.Address(stackData[stackLen-1].Bytes20())
		t.lookupAccount(addr)
		if op == vm.SELFDESTRUCT {
			t.currentTxTrace().deleted[caller] = true
		}
	case stackLen >= 5 && (op == vm.DELEGATECALL || op == vm.CALL || op == vm.STATICCALL || op == vm.CALLCODE):
		addr := common.Address(stackData[stackLen-2].Bytes20())
		t.lookupAccount(addr)
	case op == vm.CREATE:
		nonce := t.currentTxTrace().env.StateDB.GetNonce(caller)
		addr := crypto.CreateAddress(caller, nonce)
		t.lookupAccount(addr)
		t.currentTxTrace().created[addr] = true
	case stackLen >= 4 && op == vm.CREATE2:
		offset := stackData[stackLen-2]
		size := stackData[stackLen-3]
		offsetConverted, convErr := safecast.Convert[int64](offset.Uint64())
		if convErr != nil {
			log.Warn("failed to copy CREATE2 input, offset conversion error:", convErr)
			return
		}
		sizeConverted, convErr := safecast.Convert[int64](size.Uint64())
		if convErr != nil {
			log.Warn("failed to copy CREATE2 input, size conversion error:", convErr)
			return
		}
		init, err := getMemoryCopyPadded(scope.MemoryData(), offsetConverted, sizeConverted)
		if err != nil {
			log.Warn("failed to copy CREATE2 input", "err", err, "tracer", "prestateTracer", "offset", offset, "size", size)
			return
		}
		inithash := crypto.Keccak256(init)
		salt := stackData[stackLen-4]
		addr := crypto.CreateAddress2(caller, salt.Bytes32(), inithash)
		t.lookupAccount(addr)
		t.currentTxTrace().created[addr] = true
	}
}

func (t *prestateTracer) OnTxStart(env *tracing.VMContext, tx *types.Transaction, from common.Address) {
	t.txTraces = append(t.txTraces, &prestateTxTrace{
		txHash:  tx.Hash(),
		env:     env,
		pre:     make(PrestateAccountMap),
		post:    make(PrestateAccountMap),
		created: make(map[common.Address]bool),
		deleted: make(map[common.Address]bool),
	})

	txTrace := t.currentTxTrace()

	if tx.To() == nil {
		createdAddr := crypto.CreateAddress(from, env.StateDB.GetNonce(from))
		txTrace.to = createdAddr
		txTrace.created[createdAddr] = true
	} else {
		txTrace.to = *tx.To()
	}

	t.lookupAccount(from)
	t.lookupAccount(txTrace.to)
	if env != nil {
		t.lookupAccount(env.Coinbase)
	}
}

func (t *prestateTracer) OnTxEnd(receipt *types.Receipt, err error) {
	defer func() {
		// don't keep pointer to a VMContext that is no longer needed
		t.currentTxTrace().env = nil
	}()
	if err != nil {
		return
	}
	if t.config.DiffMode {
		t.processDiffState()
	}
	// the new created contracts' prestate were empty, so delete them
	for a := range t.currentTxTrace().created {
		// the created contract maybe exists in statedb before the creating tx
		if s := t.currentTxTrace().pre[a]; s != nil && s.empty {
			delete(t.currentTxTrace().pre, a)
		}
	}
}

// GetResult returns the json-encoded nested list of call traces, and any
// error arising from the encoding or forceful termination (via `Stop`).
func (t *prestateTracer) GetResult() (json.RawMessage, error) {
	r := lo.Map(t.txTraces, func(tr *prestateTxTrace, _ int) TxTraceResult {
		var b json.RawMessage
		if t.config.DiffMode {
			b = lo.Must(json.Marshal(PrestateDiffResult{tr.post, tr.pre}))
		} else {
			b = lo.Must(json.Marshal(tr.pre))
		}
		return TxTraceResult{
			TxHash: tr.txHash,
			Result: b,
		}
	})
	return json.Marshal(r)
}

// Stop terminates execution of the tracer at the first opportune moment.
func (t *prestateTracer) Stop(err error) {
	t.reason = err
	t.interrupt.Store(true)
}

func (t *prestateTracer) processDiffState() {
	txTrace := t.currentTxTrace()
	for addr, state := range txTrace.pre {
		// The deleted account's state is pruned from `post` but kept in `pre`
		if _, ok := txTrace.deleted[addr]; ok {
			continue
		}
		modified := false
		postAccount := &PrestateAccount{Storage: make(map[common.Hash]common.Hash)}
		newBalance := t.currentTxTrace().env.StateDB.GetBalance(addr).ToBig()
		newNonce := t.currentTxTrace().env.StateDB.GetNonce(addr)
		newCode := t.currentTxTrace().env.StateDB.GetCode(addr)

		if newBalance.Cmp(txTrace.pre[addr].Balance.ToInt()) != 0 {
			modified = true
			postAccount.Balance = (*hexutil.Big)(newBalance)
		}
		if newNonce != txTrace.pre[addr].Nonce {
			modified = true
			postAccount.Nonce = newNonce
		}
		if !bytes.Equal(newCode, txTrace.pre[addr].Code) {
			modified = true
			postAccount.Code = newCode
		}

		for key, val := range state.Storage {
			// don't include the empty slot
			if val == (common.Hash{}) {
				delete(txTrace.pre[addr].Storage, key)
			}

			newVal := t.currentTxTrace().env.StateDB.GetState(addr, key)
			if val == newVal {
				// Omit unchanged slots
				delete(txTrace.pre[addr].Storage, key)
			} else {
				modified = true
				if newVal != (common.Hash{}) {
					postAccount.Storage[key] = newVal
				}
			}
		}

		if modified {
			txTrace.post[addr] = postAccount
		} else {
			// if state is not modified, then no need to include into the pre state
			delete(txTrace.pre, addr)
		}
	}
}

// lookupAccount fetches details of an account and adds it to the prestate
// if it doesn't exist there.
func (t *prestateTracer) lookupAccount(addr common.Address) {
	if t.currentTxTrace().env == nil {
		return
	}
	if _, ok := t.currentTxTrace().pre[addr]; ok {
		return
	}

	acc := &PrestateAccount{
		Balance: (*hexutil.Big)(t.currentTxTrace().env.StateDB.GetBalance(addr).ToBig()),
		Nonce:   t.currentTxTrace().env.StateDB.GetNonce(addr),
		Code:    t.currentTxTrace().env.StateDB.GetCode(addr),
		Storage: make(map[common.Hash]common.Hash),
	}
	if !acc.exists() {
		acc.empty = true
	}

	t.currentTxTrace().pre[addr] = acc
}

// lookupStorage fetches the requested storage slot and adds
// it to the prestate of the given contract. It assumes `lookupAccount`
// has been performed on the contract before.
func (t *prestateTracer) lookupStorage(addr common.Address, key common.Hash) {
	if _, ok := t.currentTxTrace().pre[addr].Storage[key]; ok {
		return
	}
	t.currentTxTrace().pre[addr].Storage[key] = t.currentTxTrace().env.StateDB.GetState(addr, key)
}
