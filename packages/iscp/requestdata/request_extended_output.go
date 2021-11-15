package requestdata

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
)

type reqExtendedOutput struct {
	*iotago.ExtendedOutput
}

// implements RequestData interface
var _ RequestData = &reqExtendedOutput{}

func (r *reqExtendedOutput) Type() TypeCode {
	return TypeExtendedOutput
}

func (r *reqExtendedOutput) Request() Request {
	panic("implement me")
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
