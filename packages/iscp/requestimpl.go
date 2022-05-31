package iscp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmnames"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// RequestDataFromMarshalUtil read Request from byte stream. First byte is interpreted as boolean flag if its an off-ledger request data
func RequestDataFromMarshalUtil(mu *marshalutil.MarshalUtil) (Request, error) {
	isOffLedger, err := mu.ReadBool()
	if err != nil {
		return nil, err
	}
	if isOffLedger {
		return OffLedgerRequestDataFromMarshalUtil(mu)
	}
	// on-ledger
	return OnLedgerRequestFromMarshalUtil(mu)
}

func RequestDataToMarshalUtil(req Request, mu *marshalutil.MarshalUtil) {
	switch req := req.(type) {
	case *onLedgerRequestData:
		mu.WriteBool(false)
		req.writeToMarshalUtil(mu)
	case *offLedgerRequestData:
		mu.WriteBool(true)
		req.writeToMarshalUtil(mu)
	default:
		panic(fmt.Sprintf("RequestDataToMarshalUtil: no handler for type %T", req))
	}
}

// region offLedgerRequestData  ////////////////////////////////////////////////////////////////////////////

type offLedgerRequestData struct {
	chainID         *ChainID
	contract        Hname
	entryPoint      Hname
	params          dict.Dict
	signatureScheme OffLedgerSignatureScheme
	nonce           uint64
	allowance       *Allowance
	gasBudget       uint64
}

type iscOffLedgerSignatureScheme struct {
	publicKey *cryptolib.PublicKey
	signature []byte
}

var _ OffLedgerSignatureScheme = &iscOffLedgerSignatureScheme{}

func (s *iscOffLedgerSignatureScheme) writeEssence(mu *marshalutil.MarshalUtil) {
	publicKey := s.publicKey.AsBytes()
	mu.WriteUint8(uint8(len(publicKey))).
		WriteBytes(publicKey)
}

func (s *iscOffLedgerSignatureScheme) writeSignature(mu *marshalutil.MarshalUtil) {
	mu.WriteUint16(uint16(len(s.signature)))
	mu.WriteBytes(s.signature)
}

func (s *iscOffLedgerSignatureScheme) readEssence(mu *marshalutil.MarshalUtil) error {
	pkLen, err := mu.ReadUint8()
	if err != nil {
		return err
	}
	if publicKey, err := mu.ReadBytes(int(pkLen)); err != nil {
		return err
	} else {
		s.publicKey, err = cryptolib.NewPublicKeyFromBytes(publicKey)
	}
	return err
}

func (s *iscOffLedgerSignatureScheme) readSignature(mu *marshalutil.MarshalUtil) error {
	sigLength, err := mu.ReadUint16()
	if err != nil {
		return err
	}
	s.signature, err = mu.ReadBytes(int(sigLength))
	return err
}

func (s *iscOffLedgerSignatureScheme) setPublicKey(key *cryptolib.PublicKey) {
	s.publicKey = key
}

func (s *iscOffLedgerSignatureScheme) sign(key *cryptolib.KeyPair, data []byte) {
	s.signature = key.GetPrivateKey().Sign(data)
}

func (s *iscOffLedgerSignatureScheme) verify(data []byte) bool {
	return s.publicKey.Verify(data, s.signature)
}

func (s *iscOffLedgerSignatureScheme) Sender() AgentID {
	return NewAgentID(s.publicKey.AsEd25519Address())
}

func NewOffLedgerRequest(chainID *ChainID, contract, entryPoint Hname, params dict.Dict, nonce uint64) UnsignedOffLedgerRequest {
	return &offLedgerRequestData{
		chainID:    chainID,
		contract:   contract,
		entryPoint: entryPoint,
		params:     params,
		signatureScheme: &iscOffLedgerSignatureScheme{
			publicKey: cryptolib.NewEmptyPublicKey(),
			signature: []byte{},
		},
		nonce:     nonce,
		allowance: NewEmptyAllowance(),
		gasBudget: gas.MaxGasPerCall,
	}
}

// implement Request interface
var _ Request = &offLedgerRequestData{}

func (r *offLedgerRequestData) IsOffLedger() bool {
	return true
}

var _ UnsignedOffLedgerRequest = &offLedgerRequestData{}
var _ OffLedgerRequest = &offLedgerRequestData{}

func (r *offLedgerRequestData) ChainID() *ChainID {
	return r.chainID
}

// implements Features interface
var _ Features = &offLedgerRequestData{}

func (r *offLedgerRequestData) TimeLock() *TimeData {
	return nil
}

func (r *offLedgerRequestData) Expiry() (*TimeData, iotago.Address) {
	return nil, nil
}

func (r *offLedgerRequestData) ReturnAmount() (uint64, bool) {
	return 0, false
}

// implements iscp.Calldata interface
var _ Calldata = &offLedgerRequestData{}

func (r *offLedgerRequestData) Bytes() []byte {
	mu := marshalutil.New()
	RequestDataToMarshalUtil(r, mu)
	return mu.Bytes()
}

func OffLedgerRequestDataFromMarshalUtil(mu *marshalutil.MarshalUtil) (OffLedgerRequest, error) {
	ret := &offLedgerRequestData{}
	if err := ret.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *offLedgerRequestData) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
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
	mu.Write(r.chainID).
		Write(r.contract).
		Write(r.entryPoint).
		Write(r.params).
		WriteUint64(r.nonce).
		WriteUint64(r.gasBudget)
	r.signatureScheme.writeEssence(mu)
	if !r.IsEVM() {
		mu.WriteBool(r.allowance != nil)
		if r.allowance != nil {
			r.allowance.WriteToMarshalUtil(mu)
		}
	}
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

	if r.IsEVMSendTransaction() {
		tx, err := evmtypes.DecodeTransaction(r.params.MustGet(evmnames.FieldTransaction))
		if err != nil {
			return err
		}
		signatureScheme, err := newEVMOffLedferSignatureSchemeFromTransaction(tx)
		if err != nil {
			return err
		}
		r.signatureScheme = signatureScheme
	} else if r.IsEVMEstimateGas() {
		callMsg, err := evmtypes.DecodeCallMsg(r.params.MustGet(evmnames.FieldCallMsg))
		if err != nil {
			return err
		}
		r.signatureScheme = newEVMOffLedgerSignatureScheme(callMsg.From)
	} else {
		r.signatureScheme = &iscOffLedgerSignatureScheme{}
	}
	if err = r.signatureScheme.readEssence(mu); err != nil {
		return err
	}

	if !r.IsEVM() {
		var hasAllowance bool
		if hasAllowance, err = mu.ReadBool(); err != nil {
			return err
		}
		r.allowance = nil
		if hasAllowance {
			if r.allowance, err = AllowanceFromMarshalUtil(mu); err != nil {
				return err
			}
		}
	}
	return nil
}

// only used for consensus
func (r *offLedgerRequestData) Hash() [32]byte {
	return hashing.HashData(r.Bytes())
}

// Sign signs the essence
func (r *offLedgerRequestData) Sign(key *cryptolib.KeyPair) OffLedgerRequest {
	r.signatureScheme.setPublicKey(key.GetPublicKey())
	r.signatureScheme.sign(key, r.essenceBytes())
	return r
}

// FungibleTokens is attached assets to the UTXO. Nil for off-ledger
func (r *offLedgerRequestData) FungibleTokens() *FungibleTokens {
	return nil
}

func (r *offLedgerRequestData) NFT() *NFT {
	return nil
}

// Allowance from the sender's account to the target smart contract. Nil mean no Allowance
func (r *offLedgerRequestData) Allowance() *Allowance {
	return r.allowance
}

func (r *offLedgerRequestData) WithGasBudget(gasBudget uint64) UnsignedOffLedgerRequest {
	r.gasBudget = gasBudget
	return r
}

func (r *offLedgerRequestData) WithAllowance(allowance *Allowance) UnsignedOffLedgerRequest {
	if r.IsEVM() && !allowance.IsEmpty() {
		panic("allowance is not supported in EVM requests")
	}
	r.allowance = allowance.Clone()
	return r
}

// VerifySignature verifies essence signature
func (r *offLedgerRequestData) VerifySignature() bool {
	return r.signatureScheme.verify(r.essenceBytes())
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
	if r.IsEVM() {
		panic("nonce in EVM requests is specified in the Ethereum tx")
	}
	r.nonce = nonce
	return r
}

func (r *offLedgerRequestData) Params() dict.Dict {
	return r.params
}

func (r *offLedgerRequestData) SenderAccount() AgentID {
	return r.signatureScheme.Sender()
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

func (r *offLedgerRequestData) GasBudget() uint64 {
	return r.gasBudget
}

func (r *offLedgerRequestData) String() string {
	return fmt.Sprintf("offLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, nonce: %d }",
		r.ID().String(),
		"**not impl**", // TODO r.SenderAddress().Base58(),
		r.contract.String(),
		r.entryPoint.String(),
		r.Params().String(),
		r.nonce,
	)
}

// endregion //////////////////////////////////////////////////////////

// region OnLedger ///////////////////////////////////////////////////////////////////

type onLedgerRequestData struct {
	inputID iotago.UTXOInput
	output  iotago.Output

	// the following originate from UTXOMetaData and output, and are created in `NewExtendedOutputData`

	featureBlocks    iotago.FeatureSet
	unlockConditions iotago.UnlockConditionSet
	requestMetadata  *RequestMetadata
}

func OnLedgerFromUTXO(o iotago.Output, id *iotago.UTXOInput) (OnLedgerRequest, error) {
	var reqMetadata *RequestMetadata
	var err error

	fbSet := o.FeatureSet()

	reqMetadata, err = RequestMetadataFromFeatureSet(fbSet)
	if err != nil {
		return nil, err
	}

	if reqMetadata != nil {
		reqMetadata.Allowance.fillEmptyNFTIDs(o, id)
	}

	return &onLedgerRequestData{
		output:           o,
		inputID:          *id,
		featureBlocks:    fbSet,
		unlockConditions: o.UnlockConditionSet(),
		requestMetadata:  reqMetadata,
	}, nil
}

func OnLedgerRequestFromMarshalUtil(mu *marshalutil.MarshalUtil) (OnLedgerRequest, error) {
	utxoID, err := UTXOInputFromMarshalUtil(mu)
	if err != nil {
		return nil, err
	}
	outputBytesLength, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	outputBytes, err := mu.ReadBytes(int(outputBytesLength))
	if err != nil {
		return nil, err
	}
	outputType, err := mu.ReadByte()
	if err != nil {
		return nil, err
	}
	output, err := iotago.OutputSelector(uint32(outputType))
	if err != nil {
		return nil, err
	}
	_, err = output.Deserialize(outputBytes, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	return OnLedgerFromUTXO(output, utxoID)
}

func (r *onLedgerRequestData) Bytes() []byte {
	mu := marshalutil.New()
	RequestDataToMarshalUtil(r, mu)
	return mu.Bytes()
}

func (r *onLedgerRequestData) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	outputBytes, err := r.output.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
	if err != nil {
		return
	}
	UTXOInputToMarshalUtil(&r.inputID, mu)
	mu.WriteUint16(uint16(len(outputBytes)))
	mu.WriteBytes(outputBytes)
	mu.WriteByte(byte(r.output.Type()))
}

// implements Calldata interface
var _ Calldata = &onLedgerRequestData{}

func (r *onLedgerRequestData) ID() RequestID {
	return RequestID(r.inputID)
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
		chid := ChainIDFromAddress(sender.(*iotago.AliasAddress))
		return NewContractAgentID(&chid, r.requestMetadata.SenderContract)
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
	out, ok := r.output.(*iotago.NFTOutput)
	if !ok {
		return nil
	}

	ret := &NFT{}

	utxoInput := r.UTXOInput()
	ret.ID = util.NFTIDFromNFTOutput(out, utxoInput.ID())

	for _, featureBlock := range out.ImmutableFeatures {
		if block, ok := featureBlock.(*iotago.IssuerFeature); ok {
			ret.Issuer = block.Address
		}
		if block, ok := featureBlock.(*iotago.MetadataFeature); ok {
			ret.Metadata = block.Data
		}
	}

	return ret
}

func (r *onLedgerRequestData) Allowance() *Allowance {
	return r.requestMetadata.Allowance
}

func (r *onLedgerRequestData) FungibleTokens() *FungibleTokens {
	amount := r.output.Deposit()
	tokens := r.output.NativeTokenSet()
	return NewFungibleTokens(amount, tokens)
}

func (r *onLedgerRequestData) GasBudget() uint64 {
	return r.requestMetadata.GasBudget
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

func (r *onLedgerRequestData) UTXOInput() iotago.UTXOInput {
	return r.inputID
}

func (r *onLedgerRequestData) Output() iotago.Output {
	return r.output
}

// IsInternalUTXO if true the output cannot be interpreted as a request
func (r *onLedgerRequestData) IsInternalUTXO(chinID *ChainID) bool {
	if r.output.Type() == iotago.OutputFoundry {
		return true
	}
	if r.SenderAddress() == nil {
		return false
	}
	if !r.SenderAddress().Equal(chinID.AsAddress()) {
		return false
	}
	if r.requestMetadata != nil {
		return false
	}
	return true
}

// implements Features interface
var _ Features = &onLedgerRequestData{}

func (r *onLedgerRequestData) TimeLock() *TimeData {
	timelock := r.unlockConditions.Timelock()
	if timelock == nil {
		return nil
	}
	ret := &TimeData{}
	ret.MilestoneIndex = timelock.MilestoneIndex
	if timelock.UnixTime != 0 {
		ret.Time = time.Unix(int64(timelock.UnixTime), 0)
	}
	return ret
}

func (r *onLedgerRequestData) Expiry() (*TimeData, iotago.Address) {
	expiration := r.unlockConditions.Expiration()
	if expiration == nil {
		return nil, nil
	}
	ret := &TimeData{}
	ret.MilestoneIndex = expiration.MilestoneIndex
	if expiration.UnixTime != 0 {
		ret.Time = time.Unix(int64(expiration.UnixTime), 0)
	}
	return ret, expiration.ReturnAddress
}

func (r *onLedgerRequestData) ReturnAmount() (uint64, bool) {
	senderBlock := r.unlockConditions.StorageDepositReturn()
	if senderBlock == nil {
		return 0, false
	}
	return senderBlock.Amount, true
}

// endregion

// region RequestID //////////////////////////////////////////////////////////////////

type RequestID iotago.UTXOInput

const RequestIDDigestLen = 6

const RequestIDSeparator = "-"

// RequestLookupDigest is shortened version of the request id. It is guaranteed to be unique
// within one block, however it may collide globally. Used for quick checking for most requests
// if it was never seen
type RequestLookupDigest [RequestIDDigestLen + 2]byte

func NewRequestID(txid iotago.TransactionID, index uint16) RequestID {
	return RequestID(iotago.UTXOInput{
		TransactionID:          txid,
		TransactionOutputIndex: index,
	})
}

func RequestIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (RequestID, error) {
	var ret RequestID
	txidData, err := mu.ReadBytes(iotago.TransactionIDLength)
	if err != nil {
		return RequestID{}, err
	}
	ret.TransactionOutputIndex, err = mu.ReadUint16()
	if err != nil {
		return RequestID{}, err
	}
	copy(ret.TransactionID[:], txidData)
	return ret, nil
}

func RequestIDFromBytes(data []byte) (RequestID, error) {
	return RequestIDFromMarshalUtil(marshalutil.New(data))
}

func RequestIDFromString(s string) (ret RequestID, err error) {
	split := strings.Split(s, RequestIDSeparator)
	if len(split) != 2 {
		return ret, fmt.Errorf("error parsing requestID")
	}
	txOutputIndex, err := strconv.ParseUint(split[0], 10, 16)
	if err != nil {
		return ret, err
	}
	ret.TransactionOutputIndex = uint16(txOutputIndex)
	txID, err := hex.DecodeString(split[1])
	if err != nil {
		return ret, err
	}
	copy(ret.TransactionID[:], txID)
	return ret, nil
}

func (rid RequestID) UTXOInput() *iotago.UTXOInput {
	r := iotago.UTXOInput(rid)
	return &r
}

func (rid RequestID) OutputID() iotago.OutputID {
	r := iotago.UTXOInput(rid)
	return r.ID()
}

func (rid RequestID) LookupDigest() RequestLookupDigest {
	ret := RequestLookupDigest{}
	copy(ret[:RequestIDDigestLen], rid.TransactionID[:RequestIDDigestLen])
	copy(ret[RequestIDDigestLen:RequestIDDigestLen+2], util.Uint16To2Bytes(rid.TransactionOutputIndex))
	return ret
}

func (rid RequestID) Bytes() []byte {
	var buf bytes.Buffer
	buf.Write(rid.TransactionID[:])
	buf.Write(util.Uint16To2Bytes(rid.TransactionOutputIndex))
	return buf.Bytes()
}

func (rid RequestID) String() string {
	return OID(rid.UTXOInput())
}

func (rid RequestID) Short() string {
	oid := rid.UTXOInput()
	txid := TxID(oid.TransactionID)
	return fmt.Sprintf("%d%s%s", oid.TransactionOutputIndex, RequestIDSeparator, txid[:6]+"..")
}

func (rid RequestID) Equals(reqID2 RequestID) bool {
	if rid.TransactionOutputIndex != reqID2.TransactionOutputIndex {
		return false
	}
	return rid.TransactionID == reqID2.TransactionID
}

func OID(o *iotago.UTXOInput) string {
	return fmt.Sprintf("%d%s%s", o.TransactionOutputIndex, RequestIDSeparator, TxID(o.TransactionID))
}

func TxID(txID iotago.TransactionID) string {
	return hex.EncodeToString(txID[:])
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
	SenderContract Hname
	// ID of the target smart contract
	TargetContract Hname
	// entry point code
	EntryPoint Hname
	// request arguments
	Params dict.Dict
	// Allowance intended to the target contract to take. Nil means zero allowance
	Allowance *Allowance
	// gas budget
	GasBudget uint64
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
	mu.WriteBool(!p.Allowance.IsEmpty())
	if !p.Allowance.IsEmpty() {
		p.Allowance.WriteToMarshalUtil(mu)
	}
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
	allowanceNotEmpty, err := mu.ReadBool()
	if err != nil {
		return err
	}
	if allowanceNotEmpty {
		if p.Allowance, err = AllowanceFromMarshalUtil(mu); err != nil {
			return err
		}
	}
	return nil
}

// endregion ///////////////////////////////////////////////////////////////
