package request

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/color"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"go.uber.org/atomic"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"
)

const (
	OnLedgerRequestType  byte = 0
	OffLedgerRequestType byte = 1
)

func FromMarshalUtil(mu *marshalutil.MarshalUtil) (iscp.Request, error) {
	b, err := mu.ReadByte()
	if err != nil {
		return nil, xerrors.Errorf("Request.FromMarshalUtil: %w", err)
	}
	// first byte is the request type
	switch b {
	case OnLedgerRequestType:
		return onLedgerFromMarshalUtil(mu)
	case OffLedgerRequestType:
		return offLedgerFromMarshalUtil(mu)
	}
	return nil, xerrors.Errorf("invalid Request Type")
}

func FromBytes(data []byte) (iscp.Request, error) {
	return FromMarshalUtil(marshalutil.New(data))
}

func OffLedgerFromBytes(b []byte) (*RequestOffLedger, error) {
	mu := marshalutil.New(b)
	return offLedgerFromMarshalUtil(mu)
}

// region Metadata  ///////////////////////////////////////////////////////

// Metadata represents content of the data payload of the output
type Metadata struct {
	err            error
	senderContract iscp.Hname
	// ID of the target smart contract
	targetContract iscp.Hname
	// entry point code
	entryPoint iscp.Hname
	// request arguments, not decoded yet wrt blobRefs
	args requestargs.RequestArgs
}

func NewMetadata() *Metadata {
	return &Metadata{
		args: requestargs.RequestArgs(dict.New()),
	}
}

func MetadataFromBytes(data []byte) *Metadata {
	return MetadataFromMarshalUtil(marshalutil.New(data))
}

func MetadataFromMarshalUtil(mu *marshalutil.MarshalUtil) *Metadata {
	ret := NewMetadata()
	ret.err = ret.ReadFromMarshalUtil(mu)
	return ret
}

func (p *Metadata) WithSender(s iscp.Hname) *Metadata {
	p.senderContract = s
	return p
}

func (p *Metadata) WithTarget(t iscp.Hname) *Metadata {
	p.targetContract = t
	return p
}

func (p *Metadata) WithEntryPoint(ep iscp.Hname) *Metadata {
	p.entryPoint = ep
	return p
}

func (p *Metadata) WithArgs(args requestargs.RequestArgs) *Metadata {
	p.args = args.Clone()
	return p
}

func (p *Metadata) Clone() *Metadata {
	ret := *p
	ret.args = p.args.Clone()
	return &ret
}

func (p *Metadata) ParsedOk() bool {
	return p.err == nil
}

func (p *Metadata) ParsedError() error {
	return p.err
}

func (p *Metadata) SenderContract() iscp.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.senderContract
}

func (p *Metadata) TargetContract() iscp.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.targetContract
}

func (p *Metadata) EntryPoint() iscp.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.entryPoint
}

func (p *Metadata) Args() requestargs.RequestArgs {
	if !p.ParsedOk() {
		return requestargs.RequestArgs(dict.New())
	}
	return p.args
}

func (p *Metadata) Bytes() []byte {
	mu := marshalutil.New()
	p.WriteToMarshalUtil(mu)
	return mu.Bytes()
}

func (p *Metadata) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(p.senderContract).
		Write(p.targetContract).
		Write(p.entryPoint).
		Write(p.args)
}

func (p *Metadata) ReadFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error
	if p.senderContract, err = iscp.HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.targetContract, err = iscp.HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.entryPoint, err = iscp.HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	if p.entryPoint, err = iscp.HnameFromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
}

// endregion

// region RequestOnLedger //////////////////////////////////////////////////////////////////

type RequestOnLedger struct {
	outputObj       *ledgerstate.ExtendedLockedOutput
	requestMetadata *Metadata
	senderAddress   ledgerstate.Address
	params          atomic.Value // this part is mutable
	minted          color.Balances
}

// implements iscp.Request interface
var _ iscp.Request = &RequestOnLedger{}

// OnLedgerFromOutput
//nolint:revive // TODO refactor stutter request.request
func OnLedgerFromOutput(output *ledgerstate.ExtendedLockedOutput, senderAddr ledgerstate.Address, minted ...color.Balances) *RequestOnLedger {
	ret := &RequestOnLedger{
		outputObj:     output,
		senderAddress: senderAddr,
	}
	ret.requestMetadata = MetadataFromBytes(output.GetPayload())
	if len(minted) > 0 {
		ret.minted = minted[0].Clone()
	}
	return ret
}

// OnLedgerFromTransaction creates RequestOnLedger object from transaction and output index
func OnLedgerFromTransaction(tx *ledgerstate.Transaction, targetAddr ledgerstate.Address) ([]*RequestOnLedger, error) {
	senderAddr, err := utxoutil.GetSingleSender(tx)
	if err != nil {
		return nil, err
	}
	mintedAmounts := color.BalancesFromLedgerstate2(utxoutil.GetMintedAmounts(tx))
	ret := make([]*RequestOnLedger, 0)
	for _, o := range tx.Essence().Outputs() {
		if out, ok := o.(*ledgerstate.ExtendedLockedOutput); ok {
			if out.Address().Equals(targetAddr) {
				out1 := out.UpdateMintingColor().(*ledgerstate.ExtendedLockedOutput)
				ret = append(ret, OnLedgerFromOutput(out1, senderAddr, mintedAmounts))
			}
		}
	}
	return ret, nil
}

func onLedgerFromMarshalUtil(mu *marshalutil.MarshalUtil) (*RequestOnLedger, error) {
	ret := &RequestOnLedger{}
	var err error

	if ret.outputObj, err = ledgerstate.ExtendedOutputFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	if ret.senderAddress, err = ledgerstate.AddressFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	ret.requestMetadata = MetadataFromMarshalUtil(mu)
	if ret.minted, err = color.BalancesFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

func (req *RequestOnLedger) Bytes() []byte {
	return marshalutil.New().WriteByte(OnLedgerRequestType).
		Write(req.Output()).
		Write(req.senderAddress).
		Write(req.requestMetadata).
		Write(req.minted).
		Bytes()
}

func (req *RequestOnLedger) ID() iscp.RequestID {
	return iscp.RequestID(req.Output().ID())
}

func (req *RequestOnLedger) IsFeePrepaid() bool {
	return false
}

func (req *RequestOnLedger) Output() ledgerstate.Output {
	return req.outputObj
}

// Params returns solid args if decoded already or nil otherwise
func (req *RequestOnLedger) Params() (dict.Dict, bool) {
	par := req.params.Load()
	if par == nil {
		return nil, false
	}
	return par.(dict.Dict), true
}

func (req *RequestOnLedger) SenderAccount() *iscp.AgentID {
	return iscp.NewAgentID(req.senderAddress, req.requestMetadata.SenderContract())
}

func (req *RequestOnLedger) SenderAddress() ledgerstate.Address {
	return req.senderAddress
}

// Target returns target contract and target entry point
func (req *RequestOnLedger) Target() (iscp.Hname, iscp.Hname) {
	return req.requestMetadata.TargetContract(), req.requestMetadata.EntryPoint()
}

func (req *RequestOnLedger) TimeLock() time.Time {
	return req.outputObj.TimeLock()
}

func (req *RequestOnLedger) SetMetadata(d *Metadata) {
	req.requestMetadata = d.Clone()
}

func (req *RequestOnLedger) GetMetadata() *Metadata {
	return req.requestMetadata
}

func (req *RequestOnLedger) MintedAmounts() color.Balances {
	return req.minted
}

func (req *RequestOnLedger) Short() string {
	return req.outputObj.ID().Base58()[:6] + ".."
}

// only used for consensus
func (req *RequestOnLedger) Hash() [32]byte {
	return blake2b.Sum256(req.Bytes())
}

func (req *RequestOnLedger) SetParams(params dict.Dict) {
	req.params.Store(params)
}

func (req *RequestOnLedger) Args() requestargs.RequestArgs {
	return req.requestMetadata.Args()
}

// endregion /////////////////////////////////////////////////////////////////

// region RequestOffLedger  ///////////////////////////////////////////////////////

type RequestOffLedger struct {
	args       requestargs.RequestArgs
	contract   iscp.Hname
	entryPoint iscp.Hname
	params     atomic.Value // mutable
	publicKey  ed25519.PublicKey
	sender     ledgerstate.Address
	signature  ed25519.Signature
	nonce      uint64
	transfer   color.Balances
}

// implements iscp.Request interface
var _ iscp.Request = &RequestOffLedger{}

// NewRequestOffLedger creates a basic request
func NewRequestOffLedger(contract, entryPoint iscp.Hname, args requestargs.RequestArgs) *RequestOffLedger {
	return &RequestOffLedger{
		args:       args.Clone(),
		contract:   contract,
		entryPoint: entryPoint,
		nonce:      uint64(time.Now().UnixNano()),
	}
}

// Bytes encodes request as bytes
func (req *RequestOffLedger) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(OffLedgerRequestType)
	req.WriteEssenceToMarshalUtil(mu)
	mu.WriteBytes(req.signature[:])
	return mu.Bytes()
}

// offLedgerFromBytes creates a basic request from previously serialized bytes (except type byte)
func offLedgerFromMarshalUtil(mu *marshalutil.MarshalUtil) (req *RequestOffLedger, err error) {
	req = &RequestOffLedger{}
	if err := req.ReadEssenceFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	sig, err := mu.ReadBytes(len(req.signature))
	if err != nil {
		return nil, err
	}
	copy(req.signature[:], sig)
	return req, nil
}

func (req *RequestOffLedger) WriteEssenceToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(req.contract).
		Write(req.entryPoint).
		Write(req.args).
		WriteBytes(req.publicKey[:]).
		WriteUint64(req.nonce).
		Write(req.transfer)
}

func (req *RequestOffLedger) ReadEssenceFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := req.contract.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	if err := req.entryPoint.ReadFromMarshalUtil(mu); err != nil {
		return err
	}
	a, err := dict.FromMarshalUtil(mu)
	if err != nil {
		return err
	}
	req.args = requestargs.New(a)
	pk, err := mu.ReadBytes(len(req.publicKey))
	if err != nil {
		return err
	}
	copy(req.publicKey[:], pk)
	if req.nonce, err = mu.ReadUint64(); err != nil {
		return err
	}
	if req.transfer, err = color.BalancesFromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
}

// only used for consensus
func (req *RequestOffLedger) Hash() [32]byte {
	return hashing.HashData(req.Bytes())
}

// Sign signs essence
func (req *RequestOffLedger) Sign(keyPair *ed25519.KeyPair) {
	req.publicKey = keyPair.PublicKey
	mu := marshalutil.New()
	req.WriteEssenceToMarshalUtil(mu)
	req.signature = keyPair.PrivateKey.Sign(mu.Bytes())
}

// Transfer returns the transfers passed to the request
func (req *RequestOffLedger) Transfer() color.Balances {
	return req.transfer
}

// Transfer sets the transfers passed to the request
func (req *RequestOffLedger) WithTransfer(transfer color.Balances) *RequestOffLedger {
	req.transfer = transfer
	return req
}

// VerifySignature verifies essence signature
func (req *RequestOffLedger) VerifySignature() bool {
	mu := marshalutil.New()
	req.WriteEssenceToMarshalUtil(mu)
	return req.publicKey.VerifySignature(mu.Bytes(), req.signature)
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (req *RequestOffLedger) ID() (requestID iscp.RequestID) {
	txid := ledgerstate.TransactionID(hashing.HashData(req.Bytes()))
	return iscp.RequestID(ledgerstate.NewOutputID(txid, 0))
}

// IsFeePrepaid always true for off-ledger
func (req *RequestOffLedger) IsFeePrepaid() bool {
	return true
}

// Order number used for replay protection
func (req *RequestOffLedger) Nonce() uint64 {
	return req.nonce
}

func (req *RequestOffLedger) WithNonce(nonce uint64) iscp.Request {
	req.nonce = nonce
	return req
}

// Output nil for off-ledger requests
func (req *RequestOffLedger) Output() ledgerstate.Output {
	return nil
}

func (req *RequestOffLedger) Params() (dict.Dict, bool) {
	par := req.params.Load()
	if par == nil {
		return nil, false
	}
	return par.(dict.Dict), true
}

func (req *RequestOffLedger) SenderAccount() *iscp.AgentID {
	return iscp.NewAgentID(req.SenderAddress(), 0)
}

func (req *RequestOffLedger) SenderAddress() ledgerstate.Address {
	if req.sender == nil {
		req.sender = ledgerstate.NewED25519Address(req.publicKey)
	}
	return req.sender
}

func (req *RequestOffLedger) Target() (iscp.Hname, iscp.Hname) {
	return req.contract, req.entryPoint
}

// TimeLock returns time lock time or zero time if no time lock
func (req *RequestOffLedger) TimeLock() time.Time {
	// no time lock, return zero time
	return time.Time{}
}

func (req *RequestOffLedger) SetParams(params dict.Dict) {
	req.params.Store(params)
}

func (req *RequestOffLedger) Args() requestargs.RequestArgs {
	return req.args
}

// endregion /////////////////////////////////////////////////////////////////

// SolidifiableRequest is the minimal interface required for SolidifyArgs
type SolidifiableRequest interface {
	Params() (dict.Dict, bool)
	SetParams(params dict.Dict)
	Args() requestargs.RequestArgs
}

var (
	_ SolidifiableRequest = &RequestOnLedger{}
	_ SolidifiableRequest = &RequestOffLedger{}
)

// SolidifyArgs solidifies the request arguments
func SolidifyArgs(req iscp.Request, reg registry.BlobCache) (bool, error) {
	sreq := req.(SolidifiableRequest)
	par, _ := sreq.Params()
	if par != nil {
		return true, nil
	}
	solid, ok, err := sreq.Args().SolidifyRequestArguments(reg)
	if err != nil || !ok {
		return ok, err
	}
	if solid == nil {
		panic("solid == nil")
	}
	sreq.SetParams(solid)
	return true, nil
}
