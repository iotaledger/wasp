package isc

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const (
	requestKindTagOnLedger byte = iota
	requestKindTagOffLedgerISC
	requestKindTagOffLedgerEVM
	requestKindTagOffLedgerEVMEstimateGas
)

func NewRequestFromBytes(data []byte) (Request, error) {
	return NewRequestFromMarshalUtil(marshalutil.New(data))
}

func NewRequestFromMarshalUtil(mu *marshalutil.MarshalUtil) (Request, error) {
	kind, err := mu.ReadByte()
	if err != nil {
		return nil, err
	}
	var r Request
	switch kind {
	case requestKindTagOnLedger:
		r = &onLedgerRequestData{}
	case requestKindTagOffLedgerISC:
		r = &offLedgerRequestData{}
	case requestKindTagOffLedgerEVM:
		r = &evmOffLedgerRequest{}
	case requestKindTagOffLedgerEVMEstimateGas:
		r = &evmOffLedgerEstimateGasRequest{}
	default:
		panic(fmt.Sprintf("no handler for request kind %d", kind))
	}
	if err := r.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return r, nil
}

// region offLedgerRequestData  ////////////////////////////////////////////////////////////////////////////

type offLedgerRequestData struct {
	chainID         ChainID
	contract        Hname
	entryPoint      Hname
	params          dict.Dict
	signatureScheme *offLedgerSignatureScheme // nil if unsigned
	nonce           uint64
	allowance       *Assets
	gasBudget       uint64
}

type offLedgerSignatureScheme struct {
	publicKey *cryptolib.PublicKey
	signature []byte
}

func (s *offLedgerSignatureScheme) writeEssence(mu *marshalutil.MarshalUtil) {
	publicKey := s.publicKey.AsBytes()
	mu.WriteUint8(uint8(len(publicKey))).
		WriteBytes(publicKey)
}

func (s *offLedgerSignatureScheme) writeSignature(mu *marshalutil.MarshalUtil) {
	mu.WriteUint16(uint16(len(s.signature)))
	mu.WriteBytes(s.signature)
}

func (s *offLedgerSignatureScheme) readEssence(mu *marshalutil.MarshalUtil) error {
	pkLen, err := mu.ReadUint8()
	if err != nil {
		return err
	}
	publicKey, err := mu.ReadBytes(int(pkLen))
	if err != nil {
		return err
	}
	s.publicKey, err = cryptolib.NewPublicKeyFromBytes(publicKey)
	return err
}

func (s *offLedgerSignatureScheme) readSignature(mu *marshalutil.MarshalUtil) error {
	sigLength, err := mu.ReadUint16()
	if err != nil {
		return err
	}
	s.signature, err = mu.ReadBytes(int(sigLength))
	return err
}

func NewOffLedgerRequest(chainID ChainID, contract, entryPoint Hname, params dict.Dict, nonce uint64) UnsignedOffLedgerRequest {
	return &offLedgerRequestData{
		chainID:    chainID,
		contract:   contract,
		entryPoint: entryPoint,
		params:     params,
		nonce:      nonce,
		allowance:  NewEmptyAssets(),
		gasBudget:  gas.MaxGasPerRequest,
	}
}

// implement Request interface
var _ Request = &offLedgerRequestData{}

func (r *offLedgerRequestData) IsOffLedger() bool {
	return true
}

var (
	_ UnsignedOffLedgerRequest = &offLedgerRequestData{}
	_ OffLedgerRequest         = &offLedgerRequestData{}
)

func (r *offLedgerRequestData) ChainID() ChainID {
	return r.chainID
}

// implements Features interface
var _ Features = &offLedgerRequestData{}

func (r *offLedgerRequestData) TimeLock() time.Time {
	return time.Time{}
}

func (r *offLedgerRequestData) Expiry() (time.Time, iotago.Address) {
	return time.Time{}, nil
}

func (r *offLedgerRequestData) ReturnAmount() (uint64, bool) {
	return 0, false
}

// implements isc.Calldata interface
var _ Calldata = &offLedgerRequestData{}

func (r *offLedgerRequestData) Bytes() []byte {
	mu := marshalutil.New()
	r.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *offLedgerRequestData) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	r.writeEssenceToMarshalUtil(mu)
	r.signatureScheme.writeSignature(mu)
}

func (r *offLedgerRequestData) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := r.readEssenceFromMarshalUtil(mu); err != nil {
		return err
	}
	if err := r.signatureScheme.readSignature(mu); err != nil {
		return err
	}
	return nil
}

func (r *offLedgerRequestData) essenceBytes() []byte {
	mu := marshalutil.New()
	r.writeEssenceToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *offLedgerRequestData) writeEssenceToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.
		WriteByte(requestKindTagOffLedgerISC).
		Write(r.chainID).
		Write(r.contract).
		Write(r.entryPoint).
		Write(r.params).
		WriteUint64(r.nonce).
		WriteUint64(r.gasBudget)
	r.signatureScheme.writeEssence(mu)
	r.allowance.WriteToMarshalUtil(mu)
}

func (r *offLedgerRequestData) readEssenceFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error
	if r.chainID, err = ChainIDFromMarshalUtil(mu); err != nil {
		return err
	}
	if err := r.contract.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	if err := r.entryPoint.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	r.params, err = dict.FromMarshalUtil(mu)
	if err != nil {
		return err
	}
	if r.nonce, err = mu.ReadUint64(); err != nil {
		return err
	}
	if r.gasBudget, err = mu.ReadUint64(); err != nil {
		return err
	}
	r.signatureScheme = &offLedgerSignatureScheme{}
	if err := r.signatureScheme.readEssence(mu); err != nil {
		return err
	}
	if r.allowance, err = AssetsFromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
}

// Sign signs the essence
func (r *offLedgerRequestData) Sign(key *cryptolib.KeyPair) OffLedgerRequest {
	r.signatureScheme = &offLedgerSignatureScheme{
		publicKey: key.GetPublicKey(),
	}
	essence := r.essenceBytes()
	r.signatureScheme.signature = key.GetPrivateKey().Sign(essence)
	return r
}

// Assets is attached assets to the UTXO. Nil for off-ledger
func (r *offLedgerRequestData) Assets() *Assets {
	return nil
}

func (r *offLedgerRequestData) NFT() *NFT {
	return nil
}

// Allowance from the sender's account to the target smart contract. Nil mean no Allowance
func (r *offLedgerRequestData) Allowance() *Assets {
	return r.allowance
}

func (r *offLedgerRequestData) WithGasBudget(gasBudget uint64) UnsignedOffLedgerRequest {
	r.gasBudget = gasBudget
	return r
}

func (r *offLedgerRequestData) WithAllowance(allowance *Assets) UnsignedOffLedgerRequest {
	r.allowance = allowance.Clone()
	return r
}

// VerifySignature verifies essence signature
func (r *offLedgerRequestData) VerifySignature() error {
	if !r.signatureScheme.publicKey.Verify(r.essenceBytes(), r.signatureScheme.signature) {
		return errors.New("invalid signature")
	}
	return nil
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (r *offLedgerRequestData) ID() (requestID RequestID) {
	return NewRequestID(iotago.TransactionID(hashing.HashData(r.Bytes())), 0)
}

// Nonce incremental nonce used for replay protection
func (r *offLedgerRequestData) Nonce() uint64 {
	return r.nonce
}

func (r *offLedgerRequestData) WithNonce(nonce uint64) UnsignedOffLedgerRequest {
	r.nonce = nonce
	return r
}

func (r *offLedgerRequestData) Params() dict.Dict {
	return r.params
}

func (r *offLedgerRequestData) SenderAccount() AgentID {
	return NewAgentID(r.signatureScheme.publicKey.AsEd25519Address())
}

func (r *offLedgerRequestData) CallTarget() CallTarget {
	return CallTarget{
		Contract:   r.contract,
		EntryPoint: r.entryPoint,
	}
}

func (r *offLedgerRequestData) TargetAddress() iotago.Address {
	return r.chainID.AsAddress()
}

func (r *offLedgerRequestData) Timestamp() time.Time {
	// no request TX, return zero time
	return time.Time{}
}

func (r *offLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	return r.gasBudget, false
}

func (r *offLedgerRequestData) String() string {
	return fmt.Sprintf("offLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, nonce: %d }",
		r.ID().String(),
		r.SenderAccount().String(),
		r.contract.String(),
		r.entryPoint.String(),
		r.Params().String(),
		r.nonce,
	)
}

// endregion //////////////////////////////////////////////////////////

// region OnLedger ///////////////////////////////////////////////////////////////////

type onLedgerRequestData struct {
	outputID iotago.OutputID
	output   iotago.Output

	// the following originate from UTXOMetaData and output, and are created in `NewExtendedOutputData`

	featureBlocks    iotago.FeatureSet
	unlockConditions iotago.UnlockConditionSet
	requestMetadata  *RequestMetadata
}

func OnLedgerFromUTXO(output iotago.Output, outputID iotago.OutputID) (OnLedgerRequest, error) {
	r := &onLedgerRequestData{}
	if err := r.readFromUTXO(output, outputID); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *onLedgerRequestData) readFromUTXO(output iotago.Output, outputID iotago.OutputID) error {
	var reqMetadata *RequestMetadata
	var err error

	fbSet := output.FeatureSet()

	reqMetadata, err = RequestMetadataFromFeatureSet(fbSet)
	if err != nil {
		return err
	}

	if reqMetadata != nil {
		reqMetadata.Allowance.fillEmptyNFTIDs(output, outputID)
	}

	r.output = output
	r.outputID = outputID
	r.featureBlocks = fbSet
	r.unlockConditions = output.UnlockConditionSet()
	r.requestMetadata = reqMetadata
	return nil
}

func (r *onLedgerRequestData) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	outputID, err := OutputIDFromMarshalUtil(mu)
	if err != nil {
		return err
	}
	outputBytesLength, err := mu.ReadUint16()
	if err != nil {
		return err
	}
	outputBytes, err := mu.ReadBytes(int(outputBytesLength))
	if err != nil {
		return err
	}
	outputType, err := mu.ReadByte()
	if err != nil {
		return err
	}
	output, err := iotago.OutputSelector(uint32(outputType))
	if err != nil {
		return err
	}
	_, err = output.Deserialize(outputBytes, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return err
	}
	return r.readFromUTXO(output, outputID)
}

func (r *onLedgerRequestData) Clone() OnLedgerRequest {
	outputID := iotago.OutputID{}
	copy(outputID[:], r.outputID[:])

	return &onLedgerRequestData{
		outputID:         outputID,
		output:           r.output.Clone(),
		featureBlocks:    r.featureBlocks.Clone(),
		unlockConditions: util.CloneMap(r.unlockConditions),
		requestMetadata:  r.requestMetadata.Clone(),
	}
}

func (r *onLedgerRequestData) Bytes() []byte {
	mu := marshalutil.New()
	r.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *onLedgerRequestData) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.WriteByte(requestKindTagOnLedger)
	outputBytes, err := r.output.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
	if err != nil {
		return
	}
	mu = OutputIDToMarshalUtil(r.outputID, mu)
	mu.WriteUint16(uint16(len(outputBytes)))
	mu.WriteBytes(outputBytes)
	mu.WriteByte(byte(r.output.Type()))
}

// implements Calldata interface
var _ Calldata = &onLedgerRequestData{}

func (r *onLedgerRequestData) ID() RequestID {
	return RequestID(r.outputID)
}

func (r *onLedgerRequestData) Params() dict.Dict {
	return r.requestMetadata.Params
}

func (r *onLedgerRequestData) SenderAccount() AgentID {
	sender := r.SenderAddress()
	if sender == nil || r.requestMetadata == nil {
		return nil
	}
	if r.requestMetadata.SenderContract != 0 {
		if sender.Type() != iotago.AddressAlias {
			panic("inconsistency: non-alias address cannot have hname != 0")
		}
		chainID := ChainIDFromAddress(sender.(*iotago.AliasAddress))
		return NewContractAgentID(chainID, r.requestMetadata.SenderContract)
	}
	return NewAgentID(sender)
}

func (r *onLedgerRequestData) SenderAddress() iotago.Address {
	senderBlock := r.featureBlocks.SenderFeature()
	if senderBlock == nil {
		return nil
	}
	return senderBlock.Address
}

func (r *onLedgerRequestData) CallTarget() CallTarget {
	if r.requestMetadata == nil {
		return CallTarget{}
	}
	return CallTarget{
		Contract:   r.requestMetadata.TargetContract,
		EntryPoint: r.requestMetadata.EntryPoint,
	}
}

func (r *onLedgerRequestData) TargetAddress() iotago.Address {
	switch out := r.output.(type) {
	case *iotago.BasicOutput:
		return out.Ident()
	case *iotago.FoundryOutput:
		return out.Ident()
	case *iotago.NFTOutput:
		return out.Ident()
	case *iotago.AliasOutput:
		return out.AliasID.ToAddress()
	default:
		panic("onLedgerRequestData:TargetAddress implement me")
	}
}

func (r *onLedgerRequestData) NFT() *NFT {
	nftOutput, ok := r.output.(*iotago.NFTOutput)
	if !ok {
		return nil
	}

	ret := &NFT{}

	ret.ID = util.NFTIDFromNFTOutput(nftOutput, r.OutputID())

	for _, featureBlock := range nftOutput.ImmutableFeatures {
		if block, ok := featureBlock.(*iotago.IssuerFeature); ok {
			ret.Issuer = block.Address
		}
		if block, ok := featureBlock.(*iotago.MetadataFeature); ok {
			ret.Metadata = block.Data
		}
	}

	return ret
}

func (r *onLedgerRequestData) Allowance() *Assets {
	return r.requestMetadata.Allowance
}

func (r *onLedgerRequestData) Assets() *Assets {
	amount := r.output.Deposit()
	tokens := r.output.NativeTokenList()
	ret := NewAssets(amount, tokens)
	NFT := r.NFT()
	if NFT != nil {
		ret.AddNFTs(NFT.ID)
	}
	return ret
}

func (r *onLedgerRequestData) GasBudget() (gasBudget uint64, isEVM bool) {
	return r.requestMetadata.GasBudget, false
}

// implements Request interface
var _ Request = &onLedgerRequestData{}

func (r *onLedgerRequestData) IsOffLedger() bool {
	return false
}

func (r *onLedgerRequestData) Features() Features {
	return r
}

func (r *onLedgerRequestData) String() string {
	req := r.requestMetadata
	return fmt.Sprintf("onLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, GasBudget: %d }",
		r.ID().String(),
		req.SenderContract.String(),
		req.TargetContract.String(),
		req.EntryPoint.String(),
		req.Params.String(),
		req.GasBudget,
	)
}

var _ OnLedgerRequest = &onLedgerRequestData{}

func (r *onLedgerRequestData) OutputID() iotago.OutputID {
	return r.outputID
}

func (r *onLedgerRequestData) Output() iotago.Output {
	return r.output
}

// IsInternalUTXO if true the output cannot be interpreted as a request
func (r *onLedgerRequestData) IsInternalUTXO(chainID ChainID) bool {
	if r.output.Type() == iotago.OutputFoundry {
		return true
	}
	if r.SenderAddress() == nil {
		return false
	}
	if !r.SenderAddress().Equal(chainID.AsAddress()) {
		return false
	}
	if r.requestMetadata != nil {
		return false
	}
	return true
}

// implements Features interface
var _ Features = &onLedgerRequestData{}

func (r *onLedgerRequestData) TimeLock() time.Time {
	timelock := r.unlockConditions.Timelock()
	if timelock == nil {
		return time.Time{}
	}
	return time.Unix(int64(timelock.UnixTime), 0)
}

func (r *onLedgerRequestData) Expiry() (time.Time, iotago.Address) {
	expiration := r.unlockConditions.Expiration()
	if expiration == nil {
		return time.Time{}, nil
	}

	return time.Unix(int64(expiration.UnixTime), 0), expiration.ReturnAddress
}

func (r *onLedgerRequestData) ReturnAmount() (uint64, bool) {
	storageDepositReturn := r.unlockConditions.StorageDepositReturn()
	if storageDepositReturn == nil {
		return 0, false
	}
	return storageDepositReturn.Amount, true
}

// endregion

// region RequestID //////////////////////////////////////////////////////////////////

type RequestID iotago.OutputID

const RequestIDDigestLen = 6

const RequestIDSeparator = "-"

type RequestRef struct {
	ID   RequestID
	Hash hashing.HashValue
}

const RequestRefKeyLen = iotago.OutputIDLength + 32

type RequestRefKey [RequestRefKeyLen]byte

func RequestRefFromRequest(req Request) *RequestRef {
	return &RequestRef{ID: req.ID(), Hash: RequestHash(req)}
}

func RequestRefsFromRequests(reqs []Request) []*RequestRef {
	rr := make([]*RequestRef, len(reqs))
	for i := range rr {
		rr[i] = RequestRefFromRequest(reqs[i])
	}
	return rr
}

func (rr *RequestRef) AsKey() RequestRefKey {
	var key RequestRefKey
	copy(key[:], rr.Bytes())
	return key
}

func (rr *RequestRef) IsFor(req Request) bool {
	if rr.ID != req.ID() {
		return false
	}
	return rr.Hash == RequestHash(req)
}

func (rr *RequestRef) Bytes() []byte {
	ret := rr.Hash[:]
	ret = append(ret, rr.ID.Bytes()...)
	return ret
}

func RequestRefFromBytes(data []byte) (*RequestRef, error) {
	reqID, err := RequestIDFromBytes(data[hashing.HashSize:])
	if err != nil {
		return nil, err
	}
	ret := &RequestRef{
		ID: reqID,
	}
	copy(ret.Hash[:], data[:hashing.HashSize])

	return ret, nil
}

// RequestLookupDigest is shortened version of the request id. It is guaranteed to be unique
// within one block, however it may collide globally. Used for quick checking for most requests
// if it was never seen
type RequestLookupDigest [RequestIDDigestLen + 2]byte

func NewRequestID(txid iotago.TransactionID, index uint16) RequestID {
	return RequestID(iotago.OutputIDFromTransactionIDAndIndex(txid, index))
}

func RequestIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (RequestID, error) {
	outputIDData, err := mu.ReadBytes(iotago.OutputIDLength)
	if err != nil {
		return RequestID{}, err
	}

	outputID := iotago.OutputID{}
	copy(outputID[:], outputIDData)
	return RequestID(outputID), nil
}

func RequestIDFromBytes(data []byte) (RequestID, error) {
	return RequestIDFromMarshalUtil(marshalutil.New(data))
}

func RequestIDFromString(s string) (ret RequestID, err error) {
	data, err := iotago.DecodeHex(s)
	if err != nil {
		return RequestID{}, err
	}

	if len(data) != iotago.OutputIDLength {
		return ret, errors.New("error parsing requestID: wrong length")
	}

	requestID := RequestID{}
	copy(requestID[:], data)
	return requestID, nil
}

func (rid RequestID) OutputID() iotago.OutputID {
	return iotago.OutputID(rid)
}

func (rid RequestID) LookupDigest() RequestLookupDigest {
	ret := RequestLookupDigest{}
	copy(ret[:RequestIDDigestLen], rid[:RequestIDDigestLen])
	// last 2 bytes are the outputindex
	copy(ret[RequestIDDigestLen:RequestIDDigestLen+2], rid[len(rid)-2:])
	return ret
}

func (rid RequestID) Bytes() []byte {
	var buf bytes.Buffer
	buf.Write(rid[:])
	return buf.Bytes()
}

func (rid RequestID) Equals(other RequestID) bool {
	return rid == other
}

func (rid RequestID) String() string {
	return iotago.EncodeHex(rid[:])
}

func (rid RequestID) Short() string {
	ridString := rid.String()
	return fmt.Sprintf("%s..%s", ridString[2:6], ridString[len(ridString)-4:])
}

func ShortRequestIDs(ids []RequestID) []string {
	ret := make([]string, len(ids))
	for i := range ret {
		ret[i] = ids[i].Short()
	}
	return ret
}

func ShortRequestIDsFromRequests(reqs []Request) []string {
	requestIDs := make([]RequestID, len(reqs))
	for i := range reqs {
		requestIDs[i] = reqs[i].ID()
	}
	return ShortRequestIDs(requestIDs)
}

// endregion ////////////////////////////////////////////////////////////

// region RequestMetadata //////////////////////////////////////////////////

type RequestMetadata struct {
	SenderContract Hname `json:"senderContract"`
	// ID of the target smart contract
	TargetContract Hname `json:"targetContract"`
	// entry point code
	EntryPoint Hname `json:"entryPoint"`
	// request arguments
	Params dict.Dict `json:"params"`
	// Allowance intended to the target contract to take. Nil means zero allowance
	Allowance *Assets `json:"allowance"`
	// gas budget
	GasBudget uint64 `json:"gasBudget"`
}

func RequestMetadataFromFeatureSet(set iotago.FeatureSet) (*RequestMetadata, error) {
	metadataFeatBlock := set.MetadataFeature()
	if metadataFeatBlock == nil {
		return nil, nil
	}
	return RequestMetadataFromBytes(metadataFeatBlock.Data)
}

func RequestMetadataFromBytes(data []byte) (*RequestMetadata, error) {
	ret := &RequestMetadata{}
	err := ret.ReadFromMarshalUtil(marshalutil.New(data))
	return ret, err
}

// returns nil if nil pointer receiver is cloned
func (p *RequestMetadata) Clone() *RequestMetadata {
	if p == nil {
		return nil
	}

	return &RequestMetadata{
		SenderContract: p.SenderContract,
		TargetContract: p.TargetContract,
		EntryPoint:     p.EntryPoint,
		Params:         p.Params.Clone(),
		Allowance:      p.Allowance.Clone(),
		GasBudget:      p.GasBudget,
	}
}

func (p *RequestMetadata) Bytes() []byte {
	mu := marshalutil.New()
	p.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (p *RequestMetadata) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(p.SenderContract).
		Write(p.TargetContract).
		Write(p.EntryPoint).
		WriteUint64(p.GasBudget)
	p.Params.WriteToMarshalUtil(mu)
	p.Allowance.WriteToMarshalUtil(mu)
}

func (p *RequestMetadata) ReadFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error
	if p.SenderContract, err = HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.TargetContract, err = HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.EntryPoint, err = HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.GasBudget, err = mu.ReadUint64(); err != nil {
		return err
	}
	if p.Params, err = dict.FromMarshalUtil(mu); err != nil {
		return err
	}
	if p.Allowance, err = AssetsFromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
}

// endregion ///////////////////////////////////////////////////////////////
