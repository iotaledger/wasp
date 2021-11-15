package requestdata

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
)

type reqAliasOutput struct {
	UTXOMetaData
	*iotago.AliasOutput
}

// implements RequestData interface
var _ RequestData = &reqAliasOutput{}

func (r *reqAliasOutput) Type() TypeCode {
	return TypeAliasOutput
}

// AliasOutput is not considered a valid request. We may implement is as a request in the future
func (r *reqAliasOutput) Request() Request {
	return nil
}

func (r *reqAliasOutput) TimeData() *TimeData {
	panic("implement me")
}

func (r *reqAliasOutput) MustUnwrap() unwrap {
	return r
}

func (r *reqAliasOutput) Features() Features {
	return r
}

func (r *reqAliasOutput) Bytes() []byte {
	panic("implement me")
}

func (r *reqAliasOutput) String() string {
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &reqAliasOutput{}

func (r *reqAliasOutput) OffLedger() *OffLedger {
	panic("not an off-ledger RequestData")
}

func (r *reqAliasOutput) UTXO() unwrapUTXO {
	return r
}

// implements unwrapUTXO interface
var _ unwrapUTXO = &reqAliasOutput{}

func (r *reqAliasOutput) MetaData() UTXOMetaData {
	return r.UTXOMetaData
}

func (r *reqAliasOutput) Simple() *iotago.SimpleOutput {
	panic("not an Simple RequestData ")
}

func (r *reqAliasOutput) Alias() *iotago.AliasOutput {
	return r.AliasOutput
}

func (r *reqAliasOutput) Extended() *iotago.ExtendedOutput {
	panic("not an Extended RequestData ")
}

func (r *reqAliasOutput) NFT() *iotago.NFTOutput {
	panic("not a NFT RequestData ")
}

func (r *reqAliasOutput) Foundry() *iotago.FoundryOutput {
	panic("not a Foundry RequestData ")
}

func (r *reqAliasOutput) Unknown() *placeholders.UnknownOutput {
	panic("not an Unknown RequestData ")
}

// implements Features interface
var _ Features = &reqAliasOutput{}

func (r *reqAliasOutput) TimeLock() (TimeLockOptions, bool) {
	panic("implement me")
}

func (r *reqAliasOutput) Expiry() (ExpiryOptions, bool) {
	panic("implement me")
}

func (r *reqAliasOutput) ReturnAmount() (ReturnAmountOptions, bool) {
	panic("implement me")
}

func (r *reqAliasOutput) SwapOption() (SwapOptions, bool) {
	panic("implement me")
}
