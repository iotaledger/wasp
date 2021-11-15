package requestdata

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
)

type reqUnknownOutput struct {
	*placeholders.UnknownOutput
}

// implements RequestData interface
var _ RequestData = &reqUnknownOutput{}

func (r *reqUnknownOutput) Type() TypeCode {
	return TypeUnknown
}

func (r *reqUnknownOutput) Request() Request {
	panic("implement me")
}

func (r *reqUnknownOutput) TimeData() *TimeData {
	panic("implement me")
}

func (r *reqUnknownOutput) MustUnwrap() unwrap {
	return r
}

func (r *reqUnknownOutput) Features() Features {
	return r
}

func (r *reqUnknownOutput) Bytes() []byte {
	panic("implement me")
}

func (r *reqUnknownOutput) String() string {
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &reqUnknownOutput{}

func (r *reqUnknownOutput) OffLedger() *OffLedger {
	panic("not an off-ledger RequestData")
}

func (r *reqUnknownOutput) UTXO() unwrapUTXO {
	return r
}

// implements unwrapUTXO interface
var _ unwrapUTXO = &reqUnknownOutput{}

func (r *reqUnknownOutput) Simple() *iotago.SimpleOutput {
	panic("not an Simple RequestData ")
}

func (r *reqUnknownOutput) Alias() *iotago.AliasOutput {
	panic("not an Alias RequestData ")
}

func (r *reqUnknownOutput) Extended() *iotago.ExtendedOutput {
	panic("not an Extended RequestData ")
}

func (r *reqUnknownOutput) NFT() *iotago.NFTOutput {
	panic("not a NFT RequestData ")
}

func (r *reqUnknownOutput) Foundry() *iotago.FoundryOutput {
	panic("not an Foundry RequestData ")
}

func (r *reqUnknownOutput) Unknown() *placeholders.UnknownOutput {
	return r.UnknownOutput
}

// implements Features interface
var _ Features = &reqUnknownOutput{}

func (r *reqUnknownOutput) TimeLock() (TimeLockOptions, bool) {
	return nil, false
}

func (r *reqUnknownOutput) Expiry() (ExpiryOptions, bool) {
	return nil, false
}

func (r *reqUnknownOutput) ReturnAmount() (ReturnAmountOptions, bool) {
	return nil, false
}

func (r *reqUnknownOutput) SwapOption() (SwapOptions, bool) {
	return nil, false
}
