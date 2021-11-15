package requestdata

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
)

type reqFoundryOutput struct {
	UTXOMetaData
	*iotago.FoundryOutput
}

// implements RequestData interface
var _ RequestData = &reqFoundryOutput{}

func (r *reqFoundryOutput) Type() TypeCode {
	return TypeNFTOutput
}

func (r *reqFoundryOutput) Request() Request {
	panic("implement me")
}

func (r *reqFoundryOutput) TimeData() *TimeData {
	panic("implement me")
}

func (r *reqFoundryOutput) MustUnwrap() unwrap {
	return r
}

func (r *reqFoundryOutput) Features() Features {
	return r
}

func (r *reqFoundryOutput) Bytes() []byte {
	panic("implement me")
}

func (r *reqFoundryOutput) String() string {
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &reqFoundryOutput{}

func (r *reqFoundryOutput) OffLedger() *OffLedger {
	panic("not an off-ledger RequestData")
}

func (r *reqFoundryOutput) UTXO() unwrapUTXO {
	return r
}

// implements unwrapUTXO interface
var _ unwrapUTXO = &reqFoundryOutput{}

func (r *reqFoundryOutput) MetaData() UTXOMetaData {
	return r.UTXOMetaData
}

func (r *reqFoundryOutput) Simple() *iotago.SimpleOutput {
	panic("not an Simple RequestData ")
}

func (r *reqFoundryOutput) Alias() *iotago.AliasOutput {
	panic("not an Alias RequestData ")
}

func (r *reqFoundryOutput) Extended() *iotago.ExtendedOutput {
	panic("not an Extended RequestData ")
}

func (r *reqFoundryOutput) NFT() *iotago.NFTOutput {
	panic("not a NFT RequestData ")
}

func (r *reqFoundryOutput) Foundry() *iotago.FoundryOutput {
	return r.FoundryOutput
}

func (r *reqFoundryOutput) Unknown() *placeholders.UnknownOutput {
	panic("not an Unknown RequestData ")
}

// implements Features interface
var _ Features = &reqFoundryOutput{}

func (r *reqFoundryOutput) TimeLock() (TimeLockOptions, bool) {
	panic("implement me")
}

func (r *reqFoundryOutput) Expiry() (ExpiryOptions, bool) {
	panic("implement me")
}

func (r *reqFoundryOutput) ReturnAmount() (ReturnAmountOptions, bool) {
	panic("implement me")
}

func (r *reqFoundryOutput) SwapOption() (SwapOptions, bool) {
	panic("implement me")
}
