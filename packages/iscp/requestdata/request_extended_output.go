package requestdata

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type reqExtendedOutput struct {
	*iotago.ExtendedOutput
	utxoID             iotago.UTXOInput
	milestoneIndex     uint32
	milestoneTimestamp time.Time
}

// implements Request interface
var _ Request = &reqExtendedOutput{}

func (r *reqExtendedOutput) ID() RequestID {
	return RequestID(r.utxoID)
}

func (r *reqExtendedOutput) Params() dict.Dict {
	panic("implement me")
}

func (r *reqExtendedOutput) SenderAccount() *iscp.AgentID {
	panic("implement me")
}

func (r *reqExtendedOutput) SenderAddress() iotago.Address {
	panic("implement me")
}

func (r *reqExtendedOutput) Target() (iscp.Hname, iscp.Hname) {
	panic("implement me")
}

func (r *reqExtendedOutput) Assets() (uint64, iotago.NativeTokens) {
	panic("implement me")
}

func (r *reqExtendedOutput) GasBudget() int64 {
	panic("implement me")
}

// implements RequestData interface
var _ RequestData = &reqExtendedOutput{}

func (r *reqExtendedOutput) Type() TypeCode {
	return TypeExtendedOutput
}

func (r *reqExtendedOutput) Request() Request {
	return r
}

func (r *reqExtendedOutput) TimeData() *TimeData {
	panic("implement me")
}

func (r *reqExtendedOutput) MustUnwrap() unwrap {
	return r
}

func (r *reqExtendedOutput) Features() Features {
	return r
}

func (r *reqExtendedOutput) Bytes() []byte {
	panic("implement me")
}

func (r *reqExtendedOutput) String() string {
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &reqExtendedOutput{}

func (r *reqExtendedOutput) OffLedger() *OffLedger {
	panic("not an off-ledger RequestData")
}

func (r *reqExtendedOutput) UTXO() unwrapUTXO {
	return r
}

// implements unwrapUTXO interface
var _ unwrapUTXO = &reqExtendedOutput{}

func (r *reqExtendedOutput) Simple() *iotago.SimpleOutput {
	panic("not an Simple RequestData ")
}

func (r *reqExtendedOutput) Alias() *iotago.AliasOutput {
	panic("not an Alias RequestData ")
}

func (r *reqExtendedOutput) Extended() *iotago.ExtendedOutput {
	return r.ExtendedOutput
}

func (r *reqExtendedOutput) NFT() *iotago.NFTOutput {
	panic("not a NFT RequestData ")
}

func (r *reqExtendedOutput) Foundry() *iotago.FoundryOutput {
	panic("not a Foundry RequestData ")
}

func (r *reqExtendedOutput) Unknown() *placeholders.UnknownOutput {
	panic("not an Unknown RequestData ")
}

// implements Features interface
var _ Features = &reqExtendedOutput{}

func (r *reqExtendedOutput) TimeLock() (TimeLockOptions, bool) {
	panic("implement me")
}

func (r *reqExtendedOutput) Expiry() (ExpiryOptions, bool) {
	panic("implement me")
}

func (r *reqExtendedOutput) ReturnAmount() (ReturnAmountOptions, bool) {
	panic("implement me")
}

func (r *reqExtendedOutput) SwapOption() (SwapOptions, bool) {
	panic("implement me")
}
