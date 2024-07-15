package isc

import (
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type onLedgerRequestData struct {
	outputID sui.ObjectID
	output   iscmove.Request

	// the following originate from UTXOMetaData and output, and are created in `NewExtendedOutputData`

	featureBlocks    iotago.FeatureSet
	unlockConditions iotago.UnlockConditionSet
	requestMetadata  *RequestMetadata
}

var (
	_ Request         = new(onLedgerRequestData)
	_ OnLedgerRequest = new(onLedgerRequestData)
	_ Calldata        = new(onLedgerRequestData)
	_ Features        = new(onLedgerRequestData)
)

func OnLedgerFromUTXO(output iotago.Output, outputID iotago.OutputID) (OnLedgerRequest, error) {
	r := &onLedgerRequestData{}
	if err := r.readFromUTXO(output, outputID); err != nil {
		return nil, err
	}
	return r, nil
}

func (req *onLedgerRequestData) readFromUTXO(output iotago.Output, outputID iotago.OutputID) error {
	var reqMetadata *RequestMetadata
	var err error

	fbSet := output.FeatureSet()

	reqMetadata, err = requestMetadataFromFeatureSet(fbSet)
	if err != nil {
		reqMetadata = nil // bad metadata. // we must handle these request, so that those funds are not lost forever
	}

	req.output = output
	req.outputID = outputID
	req.featureBlocks = fbSet
	req.unlockConditions = output.UnlockConditionSet()
	req.requestMetadata = reqMetadata
	return nil
}

func (req *onLedgerRequestData) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadKindAndVerify(rwutil.Kind(requestKindOnLedger))
	rr.ReadN(req.outputID[:])
	outputData := rr.ReadBytes()
	if rr.Err != nil {
		return rr.Err
	}
	req.output, rr.Err = util.OutputFromBytes(outputData)
	if rr.Err != nil {
		return rr.Err
	}
	return req.readFromUTXO(req.output, req.outputID)
}

func (req *onLedgerRequestData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteKind(rwutil.Kind(requestKindOnLedger))
	ww.WriteN(req.outputID[:])
	if ww.Err != nil {
		return ww.Err
	}
	outputData, err := req.output.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
	ww.Err = err
	ww.WriteBytes(outputData)
	return ww.Err
}

func (req *onLedgerRequestData) Allowance() *Assets {
	if req.requestMetadata == nil {
		return NewEmptyAssets()
	}
	return req.requestMetadata.Allowance
}

func (req *onLedgerRequestData) Assets() *Assets {
	amount := req.output.Deposit()
	// FIXME work on the SUI version
	// tokens := req.output.NativeTokenList()
	ret := NewAssets(new(big.Int).SetUint64(amount), nil)
	return ret
}

func (req *onLedgerRequestData) Bytes() []byte {
	return rwutil.WriteToBytes(req)
}

func (req *onLedgerRequestData) Message() Message {
	if req.requestMetadata == nil {
		return Message{}
	}
	return req.requestMetadata.Message
}

func (req *onLedgerRequestData) Clone() OnLedgerRequest {
	outputID := iotago.OutputID{}
	copy(outputID[:], req.outputID[:])

	ret := &onLedgerRequestData{
		outputID:         outputID,
		output:           req.output.Clone(),
		featureBlocks:    req.featureBlocks.Clone(),
		unlockConditions: util.CloneMap(req.unlockConditions),
	}
	if req.requestMetadata != nil {
		ret.requestMetadata = req.requestMetadata.Clone()
	}
	return ret
}

func (req *onLedgerRequestData) Expiry() (time.Time, *cryptolib.Address) {
	expiration := req.unlockConditions.Expiration()
	if expiration == nil {
		return time.Time{}, nil
	}

	return time.Unix(int64(expiration.UnixTime), 0), cryptolib.NewAddressFromIotago(expiration.ReturnAddress)
}

func (req *onLedgerRequestData) Features() Features {
	return req
}

func (req *onLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	if req.requestMetadata == nil {
		return 0, false
	}
	return req.requestMetadata.GasBudget, false
}

func (req *onLedgerRequestData) ID() RequestID {
	return RequestID(req.outputID)
}

// IsInternalUTXO if true the output cannot be interpreted as a request
func (req *onLedgerRequestData) IsInternalUTXO(chainID ChainID) bool {
	if req.output.Type() == iotago.OutputFoundry {
		return true
	}
	if req.senderAddress() == nil {
		return false
	}
	if !req.senderAddress().Equals(chainID.AsAddress()) {
		return false
	}
	if req.requestMetadata != nil {
		return false
	}
	return true
}

func (req *onLedgerRequestData) IsOffLedger() bool {
	return false
}

func (req *onLedgerRequestData) NFT() *NFT {
	nftOutput, ok := req.output.(*iotago.NFTOutput)
	if !ok {
		return nil
	}

	ret := &NFT{}

	ret.ID = util.NFTIDFromNFTOutput(nftOutput, req.OutputID())

	for _, featureBlock := range nftOutput.ImmutableFeatures {
		if block, ok := featureBlock.(*iotago.IssuerFeature); ok {
			ret.Issuer = cryptolib.NewAddressFromIotago(block.Address)
		}
		if block, ok := featureBlock.(*iotago.MetadataFeature); ok {
			ret.Metadata = block.Data
		}
	}

	return ret
}

func (req *onLedgerRequestData) Output() iotago.Output {
	return req.output
}

func (req *onLedgerRequestData) OutputID() iotago.OutputID {
	return req.outputID
}

func (req *onLedgerRequestData) ReturnAmount() (uint64, bool) {
	storageDepositReturn := req.unlockConditions.StorageDepositReturn()
	if storageDepositReturn == nil {
		return 0, false
	}
	return storageDepositReturn.Amount, true
}

func (req *onLedgerRequestData) SenderAccount() AgentID {
	sender := req.senderAddress()
	if sender == nil {
		return nil
	}
	if req.requestMetadata != nil && !req.requestMetadata.SenderContract.Empty() {
		// if sender.Type() == iotago.AddressAlias {	// TODO: is it needed?
		chainID := ChainIDFromAddress(sender)
		return req.requestMetadata.SenderContract.AgentID(chainID)
		//}
	}
	return NewAgentID(sender)
}

func (req *onLedgerRequestData) senderAddress() *cryptolib.Address {
	senderBlock := req.featureBlocks.SenderFeature()
	if senderBlock == nil {
		return nil
	}
	return cryptolib.NewAddressFromIotago(senderBlock.Address)
}

func (req *onLedgerRequestData) String() string {
	metadata := req.requestMetadata
	if metadata == nil {
		return "onledger request without metadata"
	}
	return fmt.Sprintf("onLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, GasBudget: %d }",
		req.ID().String(),
		metadata.SenderContract.String(),
		metadata.Message.Target.Contract.String(),
		metadata.Message.Target.EntryPoint.String(),
		metadata.Message.Params.String(),
		metadata.GasBudget,
	)
}

func (req *onLedgerRequestData) TargetAddress() *cryptolib.Address {
	switch out := req.output.(type) {
	case *iotago.BasicOutput:
		return cryptolib.NewAddressFromIotago(out.Ident())
	case *iotago.FoundryOutput:
		return cryptolib.NewAddressFromIotago(out.Ident())
	case *iotago.NFTOutput:
		return cryptolib.NewAddressFromIotago(out.Ident())
	case *iotago.AliasOutput:
		return cryptolib.NewAddressFromIotago(out.AliasID.ToAddress())
	default:
		panic("onLedgerRequestData:TargetAddress implement me")
	}
}

func (req *onLedgerRequestData) TimeLock() time.Time {
	timelock := req.unlockConditions.Timelock()
	if timelock == nil {
		return time.Time{}
	}
	return time.Unix(int64(timelock.UnixTime), 0)
}

func (req *onLedgerRequestData) EVMCallMsg() *ethereum.CallMsg {
	return nil
}

// region RetryOnLedgerRequest //////////////////////////////////////////////////////////////////

type RetryOnLedgerRequest struct {
	OnLedgerRequest
	retryOutputID sui.ObjectID
}

func NewRetryOnLedgerRequest(req OnLedgerRequest, retryOutput sui.ObjectID) *RetryOnLedgerRequest {
	return &RetryOnLedgerRequest{
		OnLedgerRequest: req,
		retryOutputID:   retryOutput,
	}
}

func (r *RetryOnLedgerRequest) RetryOutputID() sui.ObjectID {
	return r.retryOutputID
}

func (r *RetryOnLedgerRequest) SetRetryOutputID(oid sui.ObjectID) {
	r.retryOutputID = oid
}
