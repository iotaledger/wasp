package requestdata

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
)

type reqExtendedOutput struct {
	output *iotago.ExtendedOutput
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
	panic("implement me")
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

// implements RequestData interface
var _ unwrapUTXO = &reqExtendedOutput{}

func (r *reqExtendedOutput) Simple() *iotago.SimpleOutput {
	panic("not an Simple RequestData ")
}

func (r *reqExtendedOutput) Alias() *iotago.AliasOutput {
	panic("not an Alias RequestData ")
}

func (r *reqExtendedOutput) Extended() *iotago.ExtendedOutput {
	return r.output
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
