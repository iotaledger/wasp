package request

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"go.uber.org/atomic"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"
)

const (
	onLedgerRequestType byte = iota
	offLedgerRequestType
)

// FromMarshalUtil re-creates request from bytes. First byte is treated as type of the request
func FromMarshalUtil(mu *marshalutil.MarshalUtil) (iscp.Request, error) {
	b, err := mu.ReadByte()
	if err != nil {
		return nil, xerrors.Errorf("Request.FromMarshalUtil: %w", err)
	}
	// first byte is the request type
	switch b {
	case onLedgerRequestType:
		return onLedgerFromMarshalUtil(mu)
	case offLedgerRequestType:
		return offLedgerFromMarshalUtil(mu)
	}
	return nil, xerrors.Errorf("invalid Request Type")
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
		Write(p.entryPoint)
	p.args.WriteToMarshalUtil(mu)
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
	if p.args, err = requestargs.FromMarshalUtil(mu); err != nil {
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
	minted          colored.Balances
}

// implements iscp.Request interface
var _ iscp.Request = &RequestOnLedger{}

// OnLedgerFromOutput
//nolint:revive // TODO refactor stutter request.request
func OnLedgerFromOutput(output *ledgerstate.ExtendedLockedOutput, senderAddr ledgerstate.Address, minted ...colored.Balances) *RequestOnLedger {
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
	mintedAmounts := colored.BalancesFromLedgerstate2(utxoutil.GetMintedAmounts(tx))
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

// onLedgerFromMarshalUtil unmarshals requestOnLedger
func onLedgerFromMarshalUtil(mu *marshalutil.MarshalUtil) (*RequestOnLedger, error) {
	ret := &RequestOnLedger{}
	if err := ret.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

// Bytes serializes with the request type in the first byte
func (req *RequestOnLedger) Bytes() []byte {
	mu := marshalutil.New().WriteByte(onLedgerRequestType)
	req.writeToMarshalUtil(mu)
	return mu.Bytes()
}

func (req *RequestOnLedger) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(req.Output()).
		Write(req.senderAddress).
		Write(req.requestMetadata).
		Write(req.minted)
}

func (req *RequestOnLedger) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error

	if req.outputObj, err = ledgerstate.ExtendedOutputFromMarshalUtil(mu); err != nil {
		return err
	}
	if req.senderAddress, err = ledgerstate.AddressFromMarshalUtil(mu); err != nil {
		return err
	}
	req.requestMetadata = MetadataFromMarshalUtil(mu)
	if req.minted, err = colored.BalancesFromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
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

func (req *RequestOnLedger) MintedAmounts() colored.Balances {
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
	transfer   colored.Balances
}

// implements iscp.Request interface
var _ iscp.Request = &RequestOffLedger{}

// NewOffLedger creates a basic request
func NewOffLedger(contract, entryPoint iscp.Hname, args requestargs.RequestArgs) *RequestOffLedger {
	return &RequestOffLedger{
		args:       args.Clone(),
		contract:   contract,
		entryPoint: entryPoint,
		nonce:      uint64(time.Now().UnixNano()),
	}
}

// Bytes encodes request as bytes with first type byte
func (req *RequestOffLedger) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(offLedgerRequestType)
	req.writeToMarshalUtil(mu)
	return mu.Bytes()
}

// offLedgerFromMarshalUtil creates a request from previously serialized bytes. Does not expects type byte
func offLedgerFromMarshalUtil(mu *marshalutil.MarshalUtil) (req *RequestOffLedger, err error) {
	req = &RequestOffLedger{}
	if err := req.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return req, nil
}

func (req *RequestOffLedger) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	req.writeEssenceToMarshalUtil(mu)
	mu.WriteBytes(req.signature[:])
}

func (req *RequestOffLedger) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	if err := req.readEssenceFromMarshalUtil(mu); err != nil {
		return err
	}
	sig, err := mu.ReadBytes(len(req.signature))
	if err != nil {
		return err
	}
	copy(req.signature[:], sig)
	return nil
}

func (req *RequestOffLedger) writeEssenceToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(req.contract).
		Write(req.entryPoint).
		Write(req.args).
		WriteBytes(req.publicKey[:]).
		WriteUint64(req.nonce).
		Write(req.transfer)
}

func (req *RequestOffLedger) readEssenceFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
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
	if req.transfer, err = colored.BalancesFromMarshalUtil(mu); err != nil {
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
	req.writeEssenceToMarshalUtil(mu)
	req.signature = keyPair.PrivateKey.Sign(mu.Bytes())
}

// Tokens returns the transfers passed to the request
func (req *RequestOffLedger) Tokens() colored.Balances {
	return req.transfer
}

// Tokens sets the transfers passed to the request
func (req *RequestOffLedger) WithTransfer(transfer colored.Balances) *RequestOffLedger {
	req.transfer = transfer
	return req
}

// VerifySignature verifies essence signature
func (req *RequestOffLedger) VerifySignature() bool {
	mu := marshalutil.New()
	req.writeEssenceToMarshalUtil(mu)
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
