package iscp

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func RequestDataFromBytes(data []byte) (RequestData, error) {
	mu := marshalutil.New(data)

	return RequestDataFromMarshalUtil(mu)
}

// RequestDataFromMarshalUtil read RequestData from byte stream. First byte is interpreted as boolean flag if its an off-ledger request data
func RequestDataFromMarshalUtil(mu *marshalutil.MarshalUtil) (RequestData, error) {
	isOffLedger, err := mu.ReadBool()
	if err != nil {
		return nil, err
	}
	if isOffLedger {
		req := &OffLedgerRequestData{}
		if err := req.readFromMarshalUtil(mu); err != nil {
			return nil, err
		}
		return req, nil
	}
	// on-ledger
	return OnledgerRequestFromMarshalUtil(mu)
}

// region OffLedgerRequestData  ////////////////////////////////////////////////////////////////////////////

type OffLedgerRequestData struct {
	chainID        *ChainID
	contract       Hname
	entryPoint     Hname
	params         dict.Dict
	publicKey      cryptolib.PublicKey
	sender         *iotago.Ed25519Address
	signature      []byte
	nonce          uint64
	transferIotas  uint64
	transferTokens iotago.NativeTokens
	gasBudget      uint64
}

func NewOffLedgerRequest(chainID *ChainID, contract, entryPoint Hname, params dict.Dict, nonce uint64) *OffLedgerRequestData {
	return &OffLedgerRequestData{
		chainID:    chainID,
		contract:   contract,
		entryPoint: entryPoint,
		params:     params,
		nonce:      nonce,
	}
}

// implement RequestData interface
var _ RequestData = &OffLedgerRequestData{}

func (r *OffLedgerRequestData) IsOffLedger() bool {
	return true
}

func (r *OffLedgerRequestData) Unwrap() unwrap {
	return r
}

// implements unwrap interface
var _ unwrap = &OffLedgerRequestData{}

func (r *OffLedgerRequestData) OffLedger() *OffLedgerRequestData {
	return r
}

func (r *OffLedgerRequestData) ChainID() *ChainID {
	return r.chainID
}

func (r *OffLedgerRequestData) UTXO() unwrapUTXO {
	panic("not an UTXO RequestData")
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

// implements iscp.Request interface
var _ Request = &OffLedgerRequestData{}

func (r *OffLedgerRequestData) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteBool(true)
	r.writeToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *OffLedgerRequestData) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	r.writeEssenceToMarshalUtil(mu)
	mu.WriteUint8(uint8(len(r.signature)))
	mu.WriteBytes(r.signature)
}

func (r *OffLedgerRequestData) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := r.readEssenceFromMarshalUtil(mu); err != nil {
		return err
	}
	sigLength, err := mu.ReadUint8()
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
	mu.Write(r.chainID).
		Write(r.contract).
		Write(r.entryPoint).
		Write(r.params).
		WriteBytes(r.publicKey).
		WriteUint64(r.nonce).
		WriteUint64(r.gasBudget).
		WriteUint64(r.transferIotas)
	// TODO write native Tokens
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
	_, err = dict.FromMarshalUtil(mu)
	if err != nil {
		return err
	}
	r.params = dict.New()
	pk, err := mu.ReadBytes(len(r.publicKey))
	if err != nil {
		return err
	}
	copy(r.publicKey[:], pk)
	if r.nonce, err = mu.ReadUint64(); err != nil {
		return err
	}
	if r.gasBudget, err = mu.ReadUint64(); err != nil {
		return err
	}
	if r.transferIotas, err = mu.ReadUint64(); err != nil {
		return err
	}

	// TODO read native Tokens
	//if r.transferTokens, err = colored.BalancesFromMarshalUtil(mu); err != nil {
	//	return err
	//}
	return nil
}

// only used for consensus
func (r *OffLedgerRequestData) Hash() [32]byte {
	return hashing.HashData(r.Bytes())
}

// Sign signs essence
func (r *OffLedgerRequestData) Sign(key cryptolib.KeyPair) {
	r.publicKey = key.PublicKey
	r.signature, _ = key.PrivateKey.Sign(nil, r.essenceBytes(), nil)
}

// Assets is attached assets to the UTXO. Nil for off-ledger
func (r *OffLedgerRequestData) Assets() *Assets {
	return nil
}

// Transfer transfer of assets from the sender's account to the target smart contract. Nil mean no Transfer
func (r *OffLedgerRequestData) Transfer() *Assets {
	return NewAssets(r.transferIotas, r.transferTokens)
}

func (r *OffLedgerRequestData) WithGasBudget(gasBudget uint64) *OffLedgerRequestData {
	r.gasBudget = gasBudget
	return r
}

// Tokens sets the transfers passed to the request
func (r *OffLedgerRequestData) WithIotas(transferIotas uint64) *OffLedgerRequestData {
	r.transferIotas = transferIotas
	return r
}

// Tokens sets the transfers passed to the request
func (r *OffLedgerRequestData) WithTokens(tokens iotago.NativeTokens) *OffLedgerRequestData {
	r.transferTokens = tokens // TODO clone
	return r
}

// VerifySignature verifies essence signature
func (r *OffLedgerRequestData) VerifySignature() bool {
	return cryptolib.Verify(r.publicKey, r.essenceBytes(), r.signature)
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (r *OffLedgerRequestData) ID() (requestID RequestID) {
	return NewRequestID(iotago.TransactionID(hashing.HashData(r.Bytes())), 0)
}

// Order number used for replay protection
func (r *OffLedgerRequestData) Nonce() uint64 {
	return r.nonce
}

func (r *OffLedgerRequestData) WithNonce(nonce uint64) Request {
	r.nonce = nonce
	return r
}

func (r *OffLedgerRequestData) Params() dict.Dict {
	return r.params
}

func (r *OffLedgerRequestData) SenderAccount() *AgentID {
	// TODO return iscp.NewAgentID(r.SenderAddress(), 0)
	return nil
}

func (r *OffLedgerRequestData) SenderAddress() iotago.Address {
	if r.sender == nil {
		r.sender = cryptolib.Ed25519AddressFromPubKey(r.publicKey)
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
		r.ID().Base58(),
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
	UTXOMetaData
	output iotago.Output

	// featureBlocksCache and requestMetadata originate from UTXOMetaData and output, and are created in `NewExtendedOutputData`
	featureBlocksCache iotago.FeatureBlocksSet
	requestMetadata    *RequestMetadata
}

func OnLedgerFromUTXO(data *UTXOMetaData, o iotago.Output) (*OnLedgerRequestData, error) {
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
		UTXOMetaData:       *data,
		featureBlocksCache: fbSet,
		requestMetadata:    reqMetadata,
	}, nil
}

func OnledgerRequestFromMarshalUtil(mu *marshalutil.MarshalUtil) (*OnLedgerRequestData, error) {
	utxoMetadata, err := NewUTXOMetadataFromMarshalUtil(mu)
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
	outputType := binary.LittleEndian.Uint32(outputBytes)
	output, err := iotago.OutputSelector(outputType)
	if err != nil {
		return nil, err
	}
	_, err = output.Deserialize(outputBytes, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	return OnLedgerFromUTXO(utxoMetadata, output)
}

func (r *OnLedgerRequestData) Bytes() []byte {
	mu := marshalutil.New()
	r.writeToMarshalUtil(mu)
	return mu.Bytes()
}

func (r *OnLedgerRequestData) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	outputBytes, err := r.output.Serialize(serializer.DeSeriModePerformLexicalOrdering, nil)
	if err != nil {
		return
	}
	mu.WriteBytes(r.UTXOMetaData.Bytes())
	mu.WriteUint16(uint16(len(outputBytes)))
	mu.WriteBytes(outputBytes)
}

func (r *OnLedgerRequestData) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	panic("not implemented") // TODO
}

// implements Request interface
var _ Request = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) ID() RequestID {
	return r.UTXOMetaData.RequestID()
}

func (r *OnLedgerRequestData) Params() dict.Dict {
	return r.requestMetadata.Params
}

func (r *OnLedgerRequestData) SenderAccount() *AgentID {
	if r.SenderAddress() == nil {
		return &NilAgentID
	}
	return NewAgentID(r.SenderAddress(), r.requestMetadata.SenderContract)
}

func (r *OnLedgerRequestData) SenderAddress() iotago.Address {
	senderBlock := r.featureBlocksCache.SenderFeatureBlock()
	if senderBlock == nil {
		return nil
	}
	return senderBlock.Address
}

func (r *OnLedgerRequestData) CallTarget() CallTarget {
	return CallTarget{
		Contract:   r.requestMetadata.TargetContract,
		EntryPoint: r.requestMetadata.EntryPoint,
	}
}

func (r *OnLedgerRequestData) TargetAddress() iotago.Address {
	switch out := r.output.(type) {
	case *iotago.ExtendedOutput:
		return out.Address
	default:
		panic("OnLedgerRequestData:TargetAddress implement me")
	}
}

func (r *OnLedgerRequestData) Transfer() *Assets {
	return r.requestMetadata.Transfer
}

func (r *OnLedgerRequestData) Assets() *Assets {
	amount := r.output.Deposit()
	var tokens iotago.NativeTokens
	if output, ok := r.output.(iotago.NativeTokenOutput); ok {
		tokens = output.NativeTokenSet()
	}
	return NewAssets(amount, tokens)
}

func (r *OnLedgerRequestData) GasBudget() uint64 {
	return r.requestMetadata.GasBudget
}

// implements RequestData interface
var _ RequestData = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) IsOffLedger() bool {
	return false
}

func (r *OnLedgerRequestData) Unwrap() unwrap {
	return r
}

func (r *OnLedgerRequestData) Features() Features {
	return r
}

func (r *OnLedgerRequestData) String() string {
	// TODO
	panic("implement me")
}

// implements unwrap interface
var _ unwrap = &OnLedgerRequestData{}

func (r *OnLedgerRequestData) OffLedger() *OffLedgerRequestData {
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

// endregion

// region RequestID //////////////////////////////////////////////////////////////////

type RequestID iotago.UTXOInput

const RequestIDDigestLen = 6

// RequestLookupDigest is shortened version of the request id. It is guaranteed to be unique
// within one block, however it may collide globally. Used for quick checking for most requests
// if it was never seen
type RequestLookupDigest [RequestIDDigestLen + 2]byte

func NewRequestID(txid iotago.TransactionID, index uint16) RequestID {
	return RequestID{}
}

// TODO
func RequestIDFromMarshalUtil(mu *marshalutil.MarshalUtil) (RequestID, error) {
	// ret, err := ledgerstate.OutputIDFromMarshalUtil(mu)
	return RequestID{}, nil
}

func RequestIDFromBytes(data []byte) (RequestID, error) {
	return RequestIDFromMarshalUtil(marshalutil.New(data))
}

// TODO change all Base58 to Bech
func RequestIDFromBase58(b58 string) (ret RequestID, err error) {
	//var oid ledgerstate.OutputID
	//oid, err = ledgerstate.OutputIDFromBase58(b58)
	//if err != nil {
	//	return
	//}
	//ret = RequestID(oid)
	ret = RequestID{}
	return
}

func (rid RequestID) OutputID() *iotago.UTXOInput {
	r := iotago.UTXOInput(rid)
	return &r
}

func (rid RequestID) LookupDigest() RequestLookupDigest {
	ret := RequestLookupDigest{}
	// copy(ret[:RequestIDDigestLen], rid[:RequestIDDigestLen])
	// copy(ret[RequestIDDigestLen:RequestIDDigestLen+2], util.Uint16To2Bytes(rid.OutputID().OutputIndex()))
	return ret
}

// TODO change all Base58 to Bech
// Base58 returns a base58 encoded version of the request id.
func (rid RequestID) Base58() string {
	// return ledgerstate.OutputID(rid).Base58()
	return ""
}

func (rid RequestID) Bytes() []byte {
	// TODO
	return nil
}

func (rid RequestID) String() string {
	return OID(rid.OutputID())
}

func (rid RequestID) Short() string {
	oid := rid.OutputID()
	txid := hex.EncodeToString(oid.TransactionID[:])
	return fmt.Sprintf("[%d]%s", oid.TransactionOutputIndex, txid[:6]+"..")
}

func OID(o *iotago.UTXOInput) string {
	return fmt.Sprintf("[%d]%s", o.TransactionOutputIndex, hex.EncodeToString(o.TransactionID[:]))
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
	// Transfer intended to the target contract. Always taken from the sender's account. Nil mean no Transfer
	Transfer *Assets
	// gas budget
	GasBudget uint64
}

func RequestMetadataFromFeatureBlocksSet(set iotago.FeatureBlocksSet) (*RequestMetadata, error) {
	metadataFeatBlock := set.MetadataFeatureBlock()
	if metadataFeatBlock == nil {
		return nil, nil
	}
	bytes := metadataFeatBlock.Data
	return RequestMetadataFromBytes(bytes)
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
	mu.WriteBool(p.Transfer != nil)
	if p.Transfer != nil {
		p.Transfer.WriteToMarshalUtil(mu)
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
	if err := (p.Params).ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	if transferPresent, err := mu.ReadBool(); err != nil {
		if transferPresent {
			if p.Transfer, err = NewAssetsFromMarshalUtil(mu); err != nil {
				return err
			}
		}
	}
	return nil
}

// endregion ///////////////////////////////////////////////////////////////
