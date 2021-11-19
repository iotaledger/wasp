package iscp

import (
	"github.com/iotaledger/hive.go/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type onLedgerRequestData struct {
	UTXOMetaData
	output iotago.Output

	// featureBlocksCache and requestMetadata originate from UTXOMetaData and output, and are created in `NewExtendedOutputData`
	featureBlocksCache iotago.FeatureBlocksSet
	requestMetadata    *RequestMetadata
}

func NewOnLedgerRequestData(data UTXOMetaData, o iotago.Output) (*onLedgerRequestData, error) {
	var fbSet iotago.FeatureBlocksSet
	var reqMetadata *RequestMetadata
	var err error

	fbo, ok := o.(iotago.FeatureBlockOutput)
	if !ok {
		panic("wrong type. Expected iotago.FeatureBlockOutput")
	}
	fbSet, err = fbo.FeatureBlocks().Set()
	if err != nil {
		return nil, err
	}
	reqMetadata, err = RequestMetadataFromFeatureBlocksSet(fbSet)
	if err != nil {
		return nil, err
	}

	return &onLedgerRequestData{
		output:             o,
		UTXOMetaData:       data,
		featureBlocksCache: fbSet,
		requestMetadata:    reqMetadata,
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

func (r *onLedgerRequestData) SenderAccount() *AgentID {
	return NewAgentID(r.SenderAddress(), r.requestMetadata.SenderContract())
}

func (r *onLedgerRequestData) SenderAddress() iotago.Address {
	senderBlock, has := r.featureBlocksCache[iotago.FeatureBlockSender]
	if !has {
		return nil
	}
	return senderBlock.(*iotago.SenderFeatureBlock).Address
}

func (r *onLedgerRequestData) Target() RequestTarget {
	return RequestTarget{
		Contract:   r.requestMetadata.TargetContract(),
		EntryPoint: r.requestMetadata.EntryPoint(),
	}
}

func (r *onLedgerRequestData) Transfer() *Assets {
	return r.requestMetadata.Transfer()
}

func (r *onLedgerRequestData) Assets() *Assets {
	amount := r.output.Deposit()
	var tokens iotago.NativeTokens
	if output, ok := r.output.(iotago.NativeTokenOutput); ok {
		tokens = output.NativeTokenSet()
	}
	return &Assets{
		Amount: amount,
		Tokens: tokens,
	}
}

func (r *onLedgerRequestData) GasBudget() int64 {
	return r.requestMetadata.GasBudget()
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
		MilestoneIndex: r.UTXOMetaData.MilestoneIndex,
		Timestamp:      r.UTXOMetaData.MilestoneTimestamp,
	}
}

func (r *onLedgerRequestData) Unwrap() unwrap {
	return r
}

func (r *onLedgerRequestData) Features() Features {
	return r
}

func (r *onLedgerRequestData) Bytes() []byte {
	outputBytes, err := []byte{}, error(nil) // r.output.Serialize(serializer.DeSeriModeNoValidation)
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

func (r *onLedgerRequestData) UTXO() iotago.Output {
	return r.output
}

// implements unwrapUTXO interface
var _ unwrap = &onLedgerRequestData{}

func (r *onLedgerRequestData) MetaData() UTXOMetaData {
	return r.UTXOMetaData
}

// implements Features interface
var _ Features = &onLedgerRequestData{}

func (r *onLedgerRequestData) TimeLock() *TimeData {
	timelockMilestoneFB, hasMilestoneFB := r.featureBlocksCache[iotago.FeatureBlockTimelockMilestoneIndex]
	timelockDeadlineFB, hasDeadlineFB := r.featureBlocksCache[iotago.FeatureBlockTimelockUnix]
	if !hasMilestoneFB && !hasDeadlineFB {
		return nil
	}
	ret := &TimeData{}
	if hasMilestoneFB {
		ret.MilestoneIndex = timelockMilestoneFB.(*iotago.TimelockMilestoneIndexFeatureBlock).MilestoneIndex
	}
	if hasDeadlineFB {
		ret.Timestamp = timelockDeadlineFB.(*iotago.TimelockUnixFeatureBlock).UnixTime
	}
	return ret
}

func (r *onLedgerRequestData) Expiry() *TimeData {
	expiryMilestoneFB, hasMilestoneFB := r.featureBlocksCache[iotago.FeatureBlockExpirationMilestoneIndex]
	expiryDeadlineFB, hasDeadlineFB := r.featureBlocksCache[iotago.FeatureBlockExpirationUnix]
	if !hasMilestoneFB && !hasDeadlineFB {
		return nil
	}
	ret := &TimeData{}
	if hasMilestoneFB {
		ret.MilestoneIndex = expiryMilestoneFB.(*iotago.ExpirationMilestoneIndexFeatureBlock).MilestoneIndex
	}
	if hasDeadlineFB {
		ret.Timestamp = expiryDeadlineFB.(*iotago.ExpirationUnixFeatureBlock).UnixTime
	}
	return ret
}

func (r *onLedgerRequestData) ReturnAmount() (uint64, bool) {
	//senderBlock, has := r.featureBlocksCache[iotago.FeatureBlockReturn]
	//if !has {
	//	return 0, false
	//}
	//return senderBlock.(*iotago.ReturnFeatureBlock).Amount, true
	return 0, false
}
