package iscp

import (
	"time"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type OnLedgerRequestData struct {
	UTXOMetaData
	output iotago.Output

	// featureBlocksCache and requestMetadata originate from UTXOMetaData and output, and are created in `NewExtendedOutputData`
	featureBlocksCache iotago.FeatureBlocksSet
	requestMetadata    *RequestMetadata
}

func NewOnLedgerRequestFromUTXO(data UTXOMetaData, o iotago.Output) (*OnLedgerRequestData, error) {
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

	return &OnLedgerRequestData{
		output:             o,
		UTXOMetaData:       data,
		featureBlocksCache: fbSet,
		requestMetadata:    reqMetadata,
	}, nil
}

// implements Request interface
var _ Request = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) ID() RequestID {
	return r.UTXOMetaData.RequestID()
}

func (r *OnLedgerRequestData) Params() dict.Dict {
	return r.requestMetadata.Args()
}

func (r *OnLedgerRequestData) SenderAccount() *AgentID {
	if r.SenderAddress() == nil {
		return &NilAgentID
	}
	return NewAgentID(r.SenderAddress(), r.requestMetadata.SenderContract())
}

func (r *OnLedgerRequestData) SenderAddress() iotago.Address {
	senderBlock, has := r.featureBlocksCache[iotago.FeatureBlockSender]
	if !has {
		return nil
	}
	return senderBlock.(*iotago.SenderFeatureBlock).Address
}

func (r *OnLedgerRequestData) Target() RequestTarget {
	return RequestTarget{
		Contract:   r.requestMetadata.TargetContract(),
		EntryPoint: r.requestMetadata.EntryPoint(),
	}
}

func (r *OnLedgerRequestData) Transfer() *Assets {
	return r.requestMetadata.Transfer()
}

func (r *OnLedgerRequestData) Assets() *Assets {
	amount := r.output.Deposit()
	var tokens iotago.NativeTokens
	if output, ok := r.output.(iotago.NativeTokenOutput); ok {
		tokens = output.NativeTokenSet()
	}
	return NewAssets(amount, tokens)
}

func (r *OnLedgerRequestData) GasBudget() int64 {
	return r.requestMetadata.GasBudget()
}

// implements RequestData interface
var _ RequestData = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) Type() TypeCode {
	return TypeExtendedOutput
}

func (r *OnLedgerRequestData) Request() Request {
	return r
}

func (r *OnLedgerRequestData) TimeData() *TimeData {
	return &TimeData{
		MilestoneIndex: r.UTXOMetaData.MilestoneIndex,
		Time:           r.UTXOMetaData.MilestoneTimestamp,
	}
}

func (r *OnLedgerRequestData) Unwrap() unwrap {
	return r
}

func (r *OnLedgerRequestData) Features() Features {
	return r
}

func (r *OnLedgerRequestData) Bytes() []byte {
	outputBytes, err := r.output.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil
	}
	mu := marshalutil.New()
	mu.WriteBytes(outputBytes)
	mu.WriteBytes(r.requestMetadata.Bytes())
	return mu.Bytes()
}

func (r *OnLedgerRequestData) String() string {
	// TODO
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) OffLedger() *OffLedger {
	panic("not an off-ledger RequestData")
}

func (r *OnLedgerRequestData) UTXO() unwrapUTXO {
	return r
}

// implements unwrapUTXO interface
var _ unwrapUTXO = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) Output() iotago.Output {
	return r.output
}

func (r *OnLedgerRequestData) Metadata() *UTXOMetaData {
	return &r.UTXOMetaData
}

// implements Features interface
var _ Features = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) TimeLock() *TimeData {
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
		ret.Time = time.Unix(int64(timelockDeadlineFB.(*iotago.TimelockUnixFeatureBlock).UnixTime), 0)
	}
	return ret
}

func (r *OnLedgerRequestData) Expiry() (*TimeData, iotago.Address) {
	expiryMilestoneFB, hasMilestoneFB := r.featureBlocksCache[iotago.FeatureBlockExpirationMilestoneIndex]
	expiryDeadlineFB, hasDeadlineFB := r.featureBlocksCache[iotago.FeatureBlockExpirationUnix]
	if !hasMilestoneFB && !hasDeadlineFB {
		return nil, nil
	}
	ret := &TimeData{}
	if hasMilestoneFB {
		ret.MilestoneIndex = expiryMilestoneFB.(*iotago.ExpirationMilestoneIndexFeatureBlock).MilestoneIndex
	}
	if hasDeadlineFB {
		ret.Time = time.Unix(int64(expiryDeadlineFB.(*iotago.ExpirationUnixFeatureBlock).UnixTime), 0)
	}
	return ret, r.SenderAddress()
}

func (r *OnLedgerRequestData) ReturnAmount() (uint64, bool) {
	senderBlock, has := r.featureBlocksCache[iotago.FeatureBlockDustDepositReturn]
	if !has {
		return 0, false
	}
	return senderBlock.(*iotago.DustDepositReturnFeatureBlock).Amount, true
}
