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
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
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
	case *OnLedgerRequestData:
		mu.WriteBool(false)
		req.writeToMarshalUtil(mu)
	case *OffLedgerRequestData:
		mu.WriteBool(true)
		req.writeToMarshalUtil(mu)
	default:
		panic(fmt.Sprintf("RequestDataToMarshalUtil: no handler for type %T", req))
	}
}

// region OffLedgerRequestData  ////////////////////////////////////////////////////////////////////////////

type OffLedgerRequestData struct {
	chainID    *ChainID
	contract   Hname
	entryPoint Hname
	params     dict.Dict
	publicKey  *cryptolib.PublicKey
	sender     *iotago.Ed25519Address
	signature  []byte
	nonce      uint64
	allowance  *Allowance
	gasBudget  uint64
}

func NewOffLedgerRequest(chainID *ChainID, contract, entryPoint Hname, params dict.Dict, nonce, gasBudget uint64) *OffLedgerRequestData {
	return &OffLedgerRequestData{
		chainID:    chainID,
		contract:   contract,
		entryPoint: entryPoint,
		params:     params,
		nonce:      nonce,
		gasBudget:  gasBudget,
		publicKey:  cryptolib.NewEmptyPublicKey(),
	}
}

// implement Request interface
var _ Request = &OffLedgerRequestData{}

func (r *OffLedgerRequestData) IsOffLedger() bool {
	return true
}

// implements unwrap interface
var _ AsOffLedger = &OffLedgerRequestData{}

func (r *OffLedgerRequestData) AsOffLedger() AsOffLedger {
	return r
}

func (r *OffLedgerRequestData) ChainID() *ChainID {
	return r.chainID
}

func (r *OffLedgerRequestData) AsOnLedger() AsOnLedger {
	panic("not an UTXO Request")
}

// implements Features interface
var _ Features = &OffLedgerRequestData{}

func (r *OffLedgerRequestData) TimeLock() *TimeData {
	return nil
}

func (r *OffLedgerRequestData) Expiry() (*TimeData, iotago.Address) {
	return nil, nil
}

func (r *OffLedgerRequestData) ReturnAmount() (uint64, bool) {
	return 0, false
}

// implements iscp.Calldata interface
var _ Calldata = &OffLedgerRequestData{}

func (r *OffLedgerRequestData) Bytes() []byte {
	mu := marshalutil.New()
	RequestDataToMarshalUtil(r, mu)
	return mu.Bytes()
}

func OffLedgerRequestDataFromMarshalUtil(mu *marshalutil.MarshalUtil) (*OffLedgerRequestData, error) {
	ret := &OffLedgerRequestData{}
	if err := ret.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *OffLedgerRequestData) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	r.writeEssenceToMarshalUtil(mu)
	mu.WriteUint16(uint16(len(r.signature)))
	mu.WriteBytes(r.signature)
}

func (r *OffLedgerRequestData) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := r.readEssenceFromMarshalUtil(mu); err != nil {
		return err
	}
	sigLength, err := mu.ReadUint16()
	if err != nil {
		return err
	}
	r.signature, err = mu.ReadBytes(int(sigLength))
	if err != nil {
		return err
	}
	return nil
}

func (r *OffLedgerRequestData) essenceBytes() []byte {
	mu := marshalutil.New()
	r.writeEssenceToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *OffLedgerRequestData) writeEssenceToMarshalUtil(mu *marshalutil.MarshalUtil) {
	publicKey := r.publicKey.AsBytes()
	mu.Write(r.chainID).
		Write(r.contract).
		Write(r.entryPoint).
		Write(r.params).
		WriteUint8(uint8(len(publicKey))).
		WriteBytes(publicKey).
		WriteUint64(r.nonce).
		WriteUint64(r.gasBudget)
	mu.WriteBool(r.allowance != nil)
	if r.allowance != nil {
		r.allowance.WriteToMarshalUtil(mu)
	}
}

func (r *OffLedgerRequestData) readEssenceFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
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
	pkLen, err := mu.ReadUint8()
	if err != nil {
		return err
	}
	if publicKey, err := mu.ReadBytes(int(pkLen)); err != nil {
		return err
	} else if r.publicKey, err = cryptolib.NewPublicKeyFromBytes(publicKey); err != nil {
		return err
	}
	if r.nonce, err = mu.ReadUint64(); err != nil {
		return err
	}
	if r.gasBudget, err = mu.ReadUint64(); err != nil {
		return err
	}
	var transferNotNil bool
	if transferNotNil, err = mu.ReadBool(); err != nil {
		return err
	}
	r.allowance = nil
	if transferNotNil {
		if r.allowance, err = AllowanceFromMarshalUtil(mu); err != nil {
			return err
		}
	}
	return nil
}

// only used for consensus
func (r *OffLedgerRequestData) Hash() [32]byte {
	return hashing.HashData(r.Bytes())
}

// Sign signs essence
func (r *OffLedgerRequestData) Sign(key *cryptolib.KeyPair) {
	r.publicKey = key.GetPublicKey()
	r.signature = key.GetPrivateKey().Sign(r.essenceBytes())
}

// FungibleTokens is attached assets to the UTXO. Nil for off-ledger
func (r *OffLedgerRequestData) FungibleTokens() *FungibleTokens {
	return nil
}

func (r *OffLedgerRequestData) NFT() *NFT {
	return nil
}

// Allowance from the sender's account to the target smart contract. Nil mean no Allowance
func (r *OffLedgerRequestData) Allowance() *Allowance {
	return r.allowance
}

func (r *OffLedgerRequestData) WithGasBudget(gasBudget uint64) *OffLedgerRequestData {
	r.gasBudget = gasBudget
	return r
}

func (r *OffLedgerRequestData) WithTransfer(transfer *Allowance) *OffLedgerRequestData {
	r.allowance = transfer.Clone()
	return r
}

// VerifySignature verifies essence signature
func (r *OffLedgerRequestData) VerifySignature() bool {
	return r.publicKey.Verify(r.essenceBytes(), r.signature)
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (r *OffLedgerRequestData) ID() (requestID RequestID) {
	return NewRequestID(iotago.TransactionID(hashing.HashData(r.Bytes())), 0)
}

// Nonce incremental nonce used for replay protection
func (r *OffLedgerRequestData) Nonce() uint64 {
	return r.nonce
}

func (r *OffLedgerRequestData) WithNonce(nonce uint64) Calldata {
	r.nonce = nonce
	return r
}

func (r *OffLedgerRequestData) Params() dict.Dict {
	return r.params
}

func (r *OffLedgerRequestData) SenderAccount() *AgentID {
	return NewAgentID(r.SenderAddress(), 0)
}

func (r *OffLedgerRequestData) SenderAddress() iotago.Address {
	if r.sender == nil {
		r.sender = r.publicKey.AsEd25519Address()
	}
	return r.sender
}

func (r *OffLedgerRequestData) CallTarget() CallTarget {
	return CallTarget{
		Contract:   r.contract,
		EntryPoint: r.entryPoint,
	}
}

func (r *OffLedgerRequestData) TargetAddress() iotago.Address {
	return r.chainID.AsAddress()
}

func (r *OffLedgerRequestData) Timestamp() time.Time {
	// no request TX, return zero time
	return time.Time{}
}

func (r *OffLedgerRequestData) GasBudget() uint64 {
	return r.gasBudget
}

func (r *OffLedgerRequestData) String() string {
	return fmt.Sprintf("OffLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, nonce: %d }",
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

type OnLedgerRequestData struct {
	inputID iotago.UTXOInput
	output  iotago.Output

	// the following originate from UTXOMetaData and output, and are created in `NewExtendedOutputData`

	featureBlocks    iotago.FeatureBlocksSet
	unlockConditions iotago.UnlockConditionsSet
	requestMetadata  *RequestMetadata
}

func OnLedgerFromUTXO(o iotago.Output, id *iotago.UTXOInput) (*OnLedgerRequestData, error) {
	var fbSet iotago.FeatureBlocksSet
	var reqMetadata *RequestMetadata
	var err error

	fbSet, err = o.FeatureBlocks().Set()
	if err != nil {
		return nil, err
	}

	reqMetadata, err = RequestMetadataFromFeatureBlocksSet(fbSet)
	if err != nil {
		return nil, err
	}

	if reqMetadata != nil {
		reqMetadata.Allowance.fillEmptyNFTIDs(o, id)
	}

	ucSet, err := o.UnlockConditions().Set()
	if err != nil {
		return nil, err
	}

	return &OnLedgerRequestData{
		output:           o,
		inputID:          *id,
		featureBlocks:    fbSet,
		unlockConditions: ucSet,
		requestMetadata:  reqMetadata,
	}, nil
}

func OnLedgerRequestFromMarshalUtil(mu *marshalutil.MarshalUtil) (*OnLedgerRequestData, error) {
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

func (r *OnLedgerRequestData) Bytes() []byte {
	mu := marshalutil.New()
	RequestDataToMarshalUtil(r, mu)
	return mu.Bytes()
}

func (r *OnLedgerRequestData) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
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
var _ Calldata = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) ID() RequestID {
	return RequestID(r.inputID)
}

func (r *OnLedgerRequestData) Params() dict.Dict {
	return r.requestMetadata.Params
}

func (r *OnLedgerRequestData) SenderAccount() *AgentID {
	if r.SenderAddress() == nil || r.requestMetadata == nil {
		return nil
	}
	return NewAgentID(r.SenderAddress(), r.requestMetadata.SenderContract)
}

func (r *OnLedgerRequestData) SenderAddress() iotago.Address {
	senderBlock := r.featureBlocks.SenderFeatureBlock()
	if senderBlock == nil {
		return nil
	}
	return senderBlock.Address
}

func (r *OnLedgerRequestData) CallTarget() CallTarget {
	if r.requestMetadata == nil {
		return CallTarget{}
	}
	return CallTarget{
		Contract:   r.requestMetadata.TargetContract,
		EntryPoint: r.requestMetadata.EntryPoint,
	}
}

func (r *OnLedgerRequestData) TargetAddress() iotago.Address {
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
		panic("OnLedgerRequestData:TargetAddress implement me")
	}
}

func (r *OnLedgerRequestData) NFT() *NFT {
	out, ok := r.output.(*iotago.NFTOutput)
	if !ok {
		return nil
	}

	ret := &NFT{}

	utxoInput := r.UTXOInput()
	ret.ID = util.NFTIDFromNFTOutput(out, utxoInput.ID())

	for _, featureBlock := range out.ImmutableBlocks {
		if block, ok := featureBlock.(*iotago.IssuerFeatureBlock); ok {
			ret.Issuer = block.Address
		}
		if block, ok := featureBlock.(*iotago.MetadataFeatureBlock); ok {
			ret.Metadata = block.Data
		}
	}

	return ret
}

func (r *OnLedgerRequestData) Allowance() *Allowance {
	return r.requestMetadata.Allowance
}

func (r *OnLedgerRequestData) FungibleTokens() *FungibleTokens {
	amount := r.output.Deposit()
	tokens := r.output.NativeTokenSet()
	return NewFungibleTokens(amount, tokens)
}

func (r *OnLedgerRequestData) GasBudget() uint64 {
	return r.requestMetadata.GasBudget
}

// implements Request interface
var _ Request = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) IsOffLedger() bool {
	return false
}

func (r *OnLedgerRequestData) Features() Features {
	return r
}

func (r *OnLedgerRequestData) String() string {
	req := r.requestMetadata
	return fmt.Sprintf("OnLedgerRequestData::{ ID: %s, sender: %s, target: %s, entrypoint: %s, Params: %s, GasBudget: %d }",
		r.ID().String(),
		req.SenderContract.String(),
		req.TargetContract.String(),
		req.EntryPoint.String(),
		req.Params.String(),
		req.GasBudget,
	)
}

func (r *OnLedgerRequestData) AsOffLedger() AsOffLedger {
	panic("not an off-ledger Request")
}

func (r *OnLedgerRequestData) AsOnLedger() AsOnLedger {
	return r
}

// implements AsOnLedger interface
var _ AsOnLedger = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) UTXOInput() iotago.UTXOInput {
	return r.inputID
}

func (r *OnLedgerRequestData) Output() iotago.Output {
	return r.output
}

// IsInternalUTXO if true the output cannot be interpreted as a request
func (r *OnLedgerRequestData) IsInternalUTXO(chinID *ChainID) bool {
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
var _ Features = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) TimeLock() *TimeData {
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

func (r *OnLedgerRequestData) Expiry() (*TimeData, iotago.Address) {
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

func (r *OnLedgerRequestData) ReturnAmount() (uint64, bool) {
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
	txid := TxID(&oid.TransactionID)
	return fmt.Sprintf("%d%s%s", oid.TransactionOutputIndex, RequestIDSeparator, txid[:6]+"..")
}

func (rid RequestID) Equals(reqID2 RequestID) bool {
	if rid.TransactionOutputIndex != reqID2.TransactionOutputIndex {
		return false
	}
	return rid.TransactionID == reqID2.TransactionID
}

func OID(o *iotago.UTXOInput) string {
	return fmt.Sprintf("%d%s%s", o.TransactionOutputIndex, RequestIDSeparator, TxID(&o.TransactionID))
}

func TxID(txID *iotago.TransactionID) string {
	return hex.EncodeToString(txID[:])
}

func ShortRequestIDs(ids []RequestID) []string {
	ret := make([]string, len(ids))
	for i := range ret {
		ret[i] = ids[i].Short()
	}
	return ret
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

func RequestMetadataFromFeatureBlocksSet(set iotago.FeatureBlocksSet) (*RequestMetadata, error) {
	metadataFeatBlock := set.MetadataFeatureBlock()
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
