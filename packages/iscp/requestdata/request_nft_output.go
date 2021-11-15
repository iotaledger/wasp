package requestdata

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
)

type reqNFTOutput struct {
	*iotago.NFTOutput
	UTXOMetaData
}

// implements RequestData interface
var _ RequestData = &reqNFTOutput{}

func (r *reqNFTOutput) Type() TypeCode {
	return TypeNFTOutput
}

func (r *reqNFTOutput) Request() Request {
	panic("implement me")
}

func (r *reqNFTOutput) TimeData() *TimeData {
	panic("implement me")
}

func (r *reqNFTOutput) MustUnwrap() unwrap {
	return r
}

func (r *reqNFTOutput) Features() Features {
	return r
}

func (r *reqNFTOutput) Bytes() []byte {
	panic("implement me")
}

func (r *reqNFTOutput) String() string {
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &reqNFTOutput{}

func (r *reqNFTOutput) OffLedger() *OffLedger {
	panic("not an off-ledger RequestData")
}

func (r *reqNFTOutput) UTXO() unwrapUTXO {
	return r
}

// implements unwrapUTXO interface
var _ unwrapUTXO = &reqNFTOutput{}

func (r *reqNFTOutput) MetaData() UTXOMetaData {
	return r.UTXOMetaData
}

func (r *reqNFTOutput) Simple() *iotago.SimpleOutput {
	panic("not an Simple RequestData ")
}

func (r *reqNFTOutput) Alias() *iotago.AliasOutput {
	panic("not an Alias RequestData ")
}

func (r *reqNFTOutput) Extended() *iotago.ExtendedOutput {
	panic("not an Extended RequestData ")
}

func (r *reqNFTOutput) NFT() *iotago.NFTOutput {
	return r.NFTOutput
}

func (r *reqNFTOutput) Foundry() *iotago.FoundryOutput {
	panic("not a Foundry RequestData ")
}

func (r *reqNFTOutput) Unknown() *placeholders.UnknownOutput {
	panic("not an Unknown RequestData ")
}

// implements Features interface
var _ Features = &reqNFTOutput{}

func (r *reqNFTOutput) TimeLock() (TimeLockOptions, bool) {
	panic("implement me")
}

func (r *reqNFTOutput) Expiry() (ExpiryOptions, bool) {
	panic("implement me")
}

func (r *reqNFTOutput) ReturnAmount() (ReturnAmountOptions, bool) {
	panic("implement me")
}

func (r *reqNFTOutput) SwapOption() (SwapOptions, bool) {
	panic("implement me")
}
