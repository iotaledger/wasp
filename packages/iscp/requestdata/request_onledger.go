package requestdata

import (
	"time"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/requestdata/placeholders"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type onLedgerRequestData struct {
	UTXOMetaData
	output          iotago.Output
	featureBlocks   iotago.FeatureBlocksSet
	requestMetadata *RequestMetadata
}

func NewExtendedOutputData(data UTXOMetaData, o iotago.Output) (*onLedgerRequestData, error) {
	var fbSet iotago.FeatureBlocksSet
	var reqMetadata *RequestMetadata
	var err error

	if fbo, ok := o.(iotago.FeatureBlockOutput); ok {
		fbSet, err = fbo.FeatureBlocks().Set()
		if err != nil {
			return nil, err
		}
		reqMetadata, err = RequestMetadataFromFeatureBlocksSet(fbSet)
		if err != nil {
			return nil, err
		}
	}

	return &onLedgerRequestData{
		output:          o,
		UTXOMetaData:    data,
		featureBlocks:   fbSet,
		requestMetadata: reqMetadata,
	}, nil
}

// implements Request interface
var _ Request = &onLedgerRequestData{}

func (r *onLedgerRequestData) ID() RequestID {
	return r.UTXOMetaData.RequestID()
}

func (r *onLedgerRequestData) Params() dict.Dict {
	return r.requestMetadata.Args()
}

func (r *onLedgerRequestData) SenderAccount() *iscp.AgentID {
	return iscp.NewAgentID(r.SenderAddress(), r.requestMetadata.SenderContract())
}

func (r *onLedgerRequestData) SenderAddress() iotago.Address {
	senderBlock, has := r.featureBlocks[iotago.FeatureBlockSender]
	if !has {
		return nil
	}
	return senderBlock.(*iotago.SenderFeatureBlock).Address
}

func (r *onLedgerRequestData) Target() iscp.Target {
	return iscp.Target{
		Contract:   r.requestMetadata.TargetContract(),
		Entrypoint: r.requestMetadata.EntryPoint(),
	}
}

func (r *onLedgerRequestData) Assets() Transfer {
	amount, _ := r.output.Deposit()
	var tokens iotago.NativeTokens
	if output, ok := r.output.(iotago.NativeTokenOutput); ok {
		tokens = output.NativeTokenSet()
	}
	return Transfer{
		amount: amount,
		tokens: tokens,
	}
}

func (r *onLedgerRequestData) GasBudget() int64 {
	panic("implement me") // TODO
}

// implements RequestData interface
var _ RequestData = &onLedgerRequestData{}

func (r *onLedgerRequestData) Type() TypeCode {
	return TypeExtendedOutput
}

func (r *onLedgerRequestData) Request() Request {
	return r
}

func (r *onLedgerRequestData) TimeData() *TimeData {
	return &TimeData{
		ConfirmingMilestoneIndex: r.UTXOMetaData.MilestoneIndex,
		ConfirmationTime:         r.UTXOMetaData.MilestoneTimestamp,
	}
}

func (r *onLedgerRequestData) MustUnwrap() unwrap {
	return r
}

func (r *onLedgerRequestData) Features() Features {
	return r
}

func (r *onLedgerRequestData) Bytes() []byte {
	outputBytes, err := r.output.Serialize(serializer.DeSeriModeNoValidation)
	if err != nil {
		return nil
	}
	mu := marshalutil.New()
	mu.WriteBytes(outputBytes)
	mu.WriteBytes(r.requestMetadata.Bytes())
	return mu.Bytes()
}

func (r *onLedgerRequestData) String() string {
	// TODO
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &onLedgerRequestData{}

func (r *onLedgerRequestData) OffLedger() *OffLedger {
	panic("not an off-ledger RequestData")
}

func (r *onLedgerRequestData) UTXO() unwrapUTXO {
	return r
}

// implements unwrapUTXO interface
var _ unwrapUTXO = &onLedgerRequestData{}

func (r *onLedgerRequestData) MetaData() UTXOMetaData {
	return r.UTXOMetaData
}

func (r *onLedgerRequestData) Simple() *iotago.SimpleOutput {
	return r.output.(*iotago.SimpleOutput)
}

func (r *onLedgerRequestData) Alias() *iotago.AliasOutput {
	return r.output.(*iotago.AliasOutput)
}

func (r *onLedgerRequestData) Extended() *iotago.ExtendedOutput {
	return r.output.(*iotago.ExtendedOutput)
}

func (r *onLedgerRequestData) NFT() *iotago.NFTOutput {
	return r.output.(*iotago.NFTOutput)
}

func (r *onLedgerRequestData) Foundry() *iotago.FoundryOutput {
	return r.output.(*iotago.FoundryOutput)
}

func (r *onLedgerRequestData) Unknown() *placeholders.UnknownOutput {
	// TODO not sure what to do here
	panic("not an Unknown RequestData ")
}

// implements Features interface
var _ Features = &onLedgerRequestData{}

func (r *onLedgerRequestData) TimeLock() *TimeInstant {
	timelockMilestoneFB, hasMilestoneFB := r.featureBlocks[iotago.FeatureBlockTimelockMilestoneIndex]
	timelockDeadlineFB, hasDeadlineFB := r.featureBlocks[iotago.FeatureBlockTimelockUnix]
	if !hasMilestoneFB && !hasDeadlineFB {
		return nil
	}
	ret := &TimeInstant{}
	if hasMilestoneFB {
		ret.MilestoneIndex = timelockMilestoneFB.(*iotago.TimelockMilestoneIndexFeatureBlock).MilestoneIndex
	}
	if hasDeadlineFB {
		ret.Timestamp = time.Unix(int64(timelockDeadlineFB.(*iotago.TimelockUnixFeatureBlock).UnixTime), 0)
	}
	return ret
}

func (r *onLedgerRequestData) Expiry() *TimeInstant {
	expiryMilestoneFB, hasMilestoneFB := r.featureBlocks[iotago.FeatureBlockExpirationMilestoneIndex]
	expiryDeadlineFB, hasDeadlineFB := r.featureBlocks[iotago.FeatureBlockExpirationUnix]
	if !hasMilestoneFB && !hasDeadlineFB {
		return nil
	}
	ret := &TimeInstant{}
	if hasMilestoneFB {
		ret.MilestoneIndex = expiryMilestoneFB.(*iotago.ExpirationMilestoneIndexFeatureBlock).MilestoneIndex
	}
	if hasDeadlineFB {
		ret.Timestamp = time.Unix(int64(expiryDeadlineFB.(*iotago.ExpirationUnixFeatureBlock).UnixTime), 0)
	}
	return ret
}

func (r *onLedgerRequestData) ReturnAmount() (uint64, bool) {
	senderBlock, has := r.featureBlocks[iotago.FeatureBlockReturn]
	if !has {
		return 0, false
	}
	return senderBlock.(*iotago.ReturnFeatureBlock).Amount, true
}

func (r *onLedgerRequestData) SwapOption() (SwapOptions, bool) {
	// TODO
	panic("implement me")
}
