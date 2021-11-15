package requestdata

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
)

type reqSimpleOutput struct {
	*iotago.SimpleOutput
}

// implements RequestData interface
var _ RequestData = &reqSimpleOutput{}

func (r *reqSimpleOutput) Type() TypeCode {
	return TypeSimpleOutput
}

func (r *reqSimpleOutput) Request() Request {
	panic("implement me")
}

func (r *reqSimpleOutput) TimeData() *TimeData {
	panic("implement me")
}

func (r *reqSimpleOutput) MustUnwrap() unwrap {
	return r
}

func (r *reqSimpleOutput) Features() Features {
	return r
}

func (r *reqSimpleOutput) Bytes() []byte {
	panic("implement me")
}

func (r *reqSimpleOutput) String() string {
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &reqSimpleOutput{}

func (r *reqSimpleOutput) OffLedger() *OffLedger {
	panic("not an off-ledger RequestData")
}

func (r *reqSimpleOutput) UTXO() unwrapUTXO {
	return r
}

// implements unwrapUTXO interface
var _ unwrapUTXO = &reqSimpleOutput{}

func (r *reqSimpleOutput) Simple() *iotago.SimpleOutput {
	return r.SimpleOutput
}

func (r *reqSimpleOutput) Alias() *iotago.AliasOutput {
	panic("not an Alias RequestData ")
}

func (r *reqSimpleOutput) Extended() *iotago.ExtendedOutput {
	panic("not an Extended RequestData ")
}

func (r *reqSimpleOutput) NFT() *iotago.NFTOutput {
	panic("not a NFT RequestData ")
}

func (r *reqSimpleOutput) Foundry() *iotago.FoundryOutput {
	panic("not a Foundry RequestData ")
}

func (r *reqSimpleOutput) Unknown() *placeholders.UnknownOutput {
	panic("not an Unknown RequestData ")
}

// implements Features interface
var _ Features = &reqSimpleOutput{}

func (r *reqSimpleOutput) TimeLock() (TimeLockOptions, bool) {
	return nil, false
}

func (r *reqSimpleOutput) Expiry() (ExpiryOptions, bool) {
	return nil, false
}

func (r *reqSimpleOutput) ReturnAmount() (ReturnAmountOptions, bool) {
	return nil, false
}

func (r *reqSimpleOutput) SwapOption() (SwapOptions, bool) {
	return nil, false
}
