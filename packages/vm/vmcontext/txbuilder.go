package vmcontext

import (
	"math/big"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/requestdata"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
)

// implements transaction builder used internally by the VM during batch run

// TODO dust protection not covered yet !!!

type txbuilder struct {
	vmctx *VMContext
	// anchor output of the chain
	chainOutput *iotago.AliasOutput
	// already consumed inputs
	consumedInputs []iotago.UTXOInput
	// balance of iotas is kept in the chain output
	balanceIotas uint64
	// balances of each native token is kept each in separate ExtendedOutput
	balanceNativeTokens map[iotago.NativeTokenID]*nativeTokenBalance
	// posted requests
	postedRequests []*postedRequest
	// TODO
}

type nativeTokenBalance struct {
	initial *big.Int
	balance *big.Int
}

// postedRequest represents on-ledger request posted in the anchor
type postedRequest struct {
	targetAddress iotago.Address
	assets        *requestdata.Assets
	metadata      *iscp.SendMetadata
	options       *iscp.SendOptions
}

// error codes used for handled panics
var (
	ErrInputLimitExceeded  = xerrors.Errorf("exceeded maximum number of inputs in transaction: %d", iotago.MaxInputsCount)
	ErrOutputLimitExceeded = xerrors.Errorf("exceeded maximum number of outputs in transaction: %d", iotago.MaxOutputsCount)
)

func (txb *txbuilder) clone() *txbuilder {
	ret := txb.vmctx.newTxBuilder()
	ret.consumedInputs = append(ret.consumedInputs, txb.consumedInputs...)
	ret.balanceIotas = txb.balanceIotas
	for k, v := range txb.balanceNativeTokens {
		ret.balanceNativeTokens[k] = &nativeTokenBalance{
			initial: new(big.Int).Set(v.balance),
			balance: new(big.Int).Set(v.balance),
		}
	}
	ret.postedRequests = append(ret.postedRequests, txb.postedRequests...)
	// TODO the rest
	return ret
}

func (txb *txbuilder) addConsumedInput(inp iotago.UTXOInput) int {
	if txb.inputsFull() {
		panic(ErrInputLimitExceeded)
	}
	txb.consumedInputs = append(txb.consumedInputs, inp)
	return len(txb.consumedInputs) - 1
}

func (txb *txbuilder) numConsumedInputs() int {
	return len(txb.consumedInputs)
}

func (txb *txbuilder) inputs() iotago.Inputs {
	ret := make(iotago.Inputs, len(txb.consumedInputs))
	for i := range txb.consumedInputs {
		ret[i] = &txb.consumedInputs[i]
	}
	// TODO
	return ret
}

func (txb *txbuilder) numInputsConsumed() int {
	ret := len(txb.consumedInputs)
	for _, v := range txb.balanceNativeTokens {
		if v.balance.Cmp(v.initial) != 0 {
			ret++
		}
	}
	// TODO the rest
	return ret
}

func (txb *txbuilder) inputsFull() bool {
	return txb.numInputsConsumed() >= iotago.MaxInputsCount
}

func (txb *txbuilder) numOutputsConsumed() int {
	ret := 1 // for chain output
	for _, v := range txb.balanceNativeTokens {
		// see https://stackoverflow.com/questions/64257065/is-there-another-way-of-testing-if-a-big-int-is-0
		if v.balance.Cmp(v.initial) != 0 && len(v.balance.Bits()) > 0 {
			ret++
		}
	}
	ret += len(txb.postedRequests)
	return ret
}

func (txb *txbuilder) outputsFull() bool {
	return txb.numOutputsConsumed() >= iotago.MaxOutputsCount
}

func (txb *txbuilder) addDeltaIotas(delta uint64) {
	// safe arithmetics
	n := txb.balanceIotas + delta
	if n < txb.balanceIotas {
		panic("addDeltaIotas: overflow")
	}
	txb.balanceIotas += n
}

func (txb *txbuilder) subDeltaIotas(delta uint64) {
	// safe arithmetics
	if delta > txb.balanceIotas {
		panic("subDeltaIotas: overflow")
	}
	txb.balanceIotas -= delta
}

func (txb *txbuilder) ensureNativeTokenBalance(id iotago.NativeTokenID) *nativeTokenBalance {
	if b, ok := txb.balanceNativeTokens[id]; ok {
		return b
	}
	b := &nativeTokenBalance{
		balance: txb.vmctx.loadNativeTokensOnChain(id),
	}
	txb.balanceNativeTokens[id] = b
	if txb.outputsFull() {
		panic(ErrOutputLimitExceeded)
	}
	return b
}

// use negative to subtract
func (txb *txbuilder) addDeltaNativeToken(id iotago.NativeTokenID, delta *big.Int) {
	b := txb.ensureNativeTokenBalance(id)
	// TODO safe arithmetic
	b.balance.Add(b.balance, delta)
}

func (txb *txbuilder) addPostedRequest(
	targetAddress iotago.Address,
	assets *requestdata.Assets,
	metadata *iscp.SendMetadata,
	options *iscp.SendOptions) {
	txb.postedRequests = append(txb.postedRequests, &postedRequest{
		targetAddress: targetAddress,
		assets:        assets,
		metadata:      metadata,
		options:       options,
	})
}

/////////// vmcontext methods

func (vmctx *VMContext) newTxBuilder() *txbuilder {
	return &txbuilder{
		vmctx:               vmctx,
		chainOutput:         vmctx.chainInput,
		consumedInputs:      []iotago.UTXOInput{vmctx.chainInputID},
		balanceIotas:        vmctx.chainInput.Amount,
		balanceNativeTokens: make(map[iotago.NativeTokenID]*nativeTokenBalance),
		postedRequests:      make([]*postedRequest, 0),
	}
}

func (vmctx *VMContext) BuildTransactionEssence(stateHash hashing.HashValue, timestamp time.Time) (*iotago.TransactionEssence, error) {
	//if err := vmctx.txBuilder.AddAliasOutputAsRemainder(vmctx.chainID.AsAddress(), stateHash[:]); err != nil {
	//	return nil, xerrors.Errorf("mustFinalizeRequestCall: %v", err)
	//}
	//tx, _, err := vmctx.txBuilder.WithTimestamp(timestamp).BuildEssence()
	//if err != nil {
	//	return nil, err
	//}
	//return tx, nil
	return nil, nil
}

func (vmctx *VMContext) createTxBuilderSnapshot(id int) *txbuilder {
	return vmctx.txbuilder.clone()
}

func (vmctx *VMContext) restoreTxBuilderSnapshot(snapshot *txbuilder) {
	vmctx.txbuilder = snapshot
}

func (vmctx *VMContext) loadNativeTokensOnChain(id iotago.NativeTokenID) *big.Int {
	panic("not implemented")
}
