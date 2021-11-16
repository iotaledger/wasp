package request

import (
	"fmt"
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
		return nil, xerrors.Errorf("Request.ColorFromMarshalUtil: %w", err)
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
	// used to prevent identical outputs from being generated
	requestNonce uint8
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

func (p *Metadata) WithRequestNonce(nonce uint8) *Metadata {
	p.requestNonce = nonce
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
		WriteByte(p.requestNonce)
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
	if p.requestNonce, err = mu.ReadByte(); err != nil {
		return err
	}
	if p.args, err = requestargs.FromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
}

// endregion

// region OnLedger //////////////////////////////////////////////////////////////////

type OnLedger struct {
	outputObj       *ledgerstate.ExtendedLockedOutput
	requestMetadata *Metadata
	senderAddress   ledgerstate.Address
	txTimestamp     time.Time    // Timestamp of the TX contaning this request.
	params          atomic.Value // this part is mutable
	minted          colored.Balances
}

// implements iscp.Request interface
var _ iscp.Request = &OnLedger{}

func OnLedgerFromOutput(output *ledgerstate.ExtendedLockedOutput, senderAddr ledgerstate.Address, txTimestamp time.Time, minted ...colored.Balances) *OnLedger {
	ret := &OnLedger{
		outputObj:     output,
		senderAddress: senderAddr,
		txTimestamp:   txTimestamp,
	}
	ret.requestMetadata = MetadataFromBytes(output.GetPayload())
	if len(minted) > 0 {
		ret.minted = minted[0].Clone()
	}
	return ret
}

// OnLedgerFromTransaction creates OnLedger object from transaction and output index
func OnLedgerFromTransaction(tx *ledgerstate.Transaction, targetAddr ledgerstate.Address) ([]*OnLedger, error) {
	senderAddr, err := utxoutil.GetSingleSender(tx)
	if err != nil {
		return nil, err
	}
	mintedAmounts := colored.BalancesFromL1Map(utxoutil.GetMintedAmounts(tx))
	ret := make([]*OnLedger, 0)
	for _, o := range tx.Essence().Outputs() {
		if out, ok := o.(*ledgerstate.ExtendedLockedOutput); ok {
			if out.Address().Equals(targetAddr) {
				out1 := out.UpdateMintingColor().(*ledgerstate.ExtendedLockedOutput)
				ret = append(ret, OnLedgerFromOutput(out1, senderAddr, tx.Essence().Timestamp(), mintedAmounts))
			}
		}
	}
	return ret, nil
}

// onLedgerFromMarshalUtil unmarshals requestOnLedger
func onLedgerFromMarshalUtil(mu *marshalutil.MarshalUtil) (*OnLedger, error) {
	ret := &OnLedger{}
	if err := ret.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return ret, nil
}

// Bytes serializes with the request type in the first byte
func (req *OnLedger) Bytes() []byte {
	mu := marshalutil.New().WriteByte(onLedgerRequestType)
	req.writeToMarshalUtil(mu)
	return mu.Bytes()
}

func (req *OnLedger) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(req.Output()).
		Write(req.ID()). // Goshimmer doesnt include outputID in serialization, so we neeed to add it manually
		Write(req.senderAddress).
		WriteTime(req.txTimestamp).
		Write(req.requestMetadata).
		Write(req.minted)
}

func (req *OnLedger) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error

	if req.outputObj, err = ledgerstate.ExtendedOutputFromMarshalUtil(mu); err != nil {
		return err
	}
	outputID, err := ledgerstate.OutputIDFromMarshalUtil(mu)
	if err != nil {
		return err
	}
	req.outputObj.SetID(outputID)

	if req.senderAddress, err = ledgerstate.AddressFromMarshalUtil(mu); err != nil {
		return err
	}
	if req.txTimestamp, err = mu.ReadTime(); err != nil {
		return err
	}
	req.requestMetadata = MetadataFromMarshalUtil(mu)
	if req.minted, err = colored.BalancesFromMarshalUtil(mu); err != nil {
		return err
	}
	return nil
}

func (req *OnLedger) ID() iscp.RequestID {
	return iscp.RequestID(req.Output().ID())
}

func (req *OnLedger) IsFeePrepaid() bool {
	return false
}

func (req *OnLedger) IsOffLedger() bool {
	return false
}

func (req *OnLedger) Output() ledgerstate.Output {
	return req.outputObj
}

// Params returns solid args if decoded already or nil otherwise
func (req *OnLedger) Params() (dict.Dict, bool) {
	// FIXME: this returns nil after deserializing a processed request (see tools/wasp-cli/chain/blocklog.go)
	par := req.params.Load()
	if par == nil {
		return nil, false
	}
	return par.(dict.Dict), true
}

func (req *OnLedger) SenderAccount() *iscp.AgentID {
	return iscp.NewAgentID(req.senderAddress, req.requestMetadata.SenderContract())
}

func (req *OnLedger) SenderAddress() ledgerstate.Address {
	return req.senderAddress
}

// Target returns target contract and target entry point
func (req *OnLedger) Target() iscp.RequestTarget {
	return iscp.NewRequestTarget(req.requestMetadata.TargetContract(), req.requestMetadata.EntryPoint())
}

func (req *OnLedger) Timestamp() time.Time {
	return req.txTimestamp
}

func (req *OnLedger) TimeLock() time.Time {
	return req.outputObj.TimeLock()
}

func (req *OnLedger) FallbackAddress() ledgerstate.Address {
	return req.outputObj.FallbackAddress()
}

func (req *OnLedger) FallbackDeadline() time.Time {
	_, t := req.outputObj.FallbackOptions()
	return t
}

func (req *OnLedger) SetMetadata(d *Metadata) {
	req.requestMetadata = d.Clone()
}

func (req *OnLedger) GetMetadata() *Metadata {
	return req.requestMetadata
}

func (req *OnLedger) MintedAmounts() colored.Balances {
	return req.minted
}

func (req *OnLedger) Short() string {
	return req.outputObj.ID().Base58()[:6] + ".."
}

// only used for consensus
func (req *OnLedger) Hash() [32]byte {
	return blake2b.Sum256(req.Bytes())
}

func (req *OnLedger) SetParams(params dict.Dict) {
	req.params.Store(params)
}

func (req *OnLedger) Args() requestargs.RequestArgs {
	return req.requestMetadata.Args()
}

func (req *OnLedger) String() string {
	fallbackStr := "none"
	if req.FallbackAddress() != nil {
		fallbackStr = fmt.Sprintf(
			"{Address: %s, Deadline: %s}",
			req.FallbackAddress().Base58(),
			req.FallbackDeadline().String(),
		)
	}
	timelockStr := "none"
	if !req.TimeLock().IsZero() {
		timelockStr = req.TimeLock().String()
	}
	return fmt.Sprintf(
		"OnLedger::{ ID: %s, sender: %s, senderHname: %s, target: %s, entrypoint: %s, args: %s, nonce: %d, timestamp: %s, fallback: %s, timelock: %s }",
		req.ID().Base58(),
		req.senderAddress.Base58(),
		req.requestMetadata.senderContract.String(),
		req.requestMetadata.targetContract.String(),
		req.requestMetadata.entryPoint.String(),
		req.Args().String(),
		req.requestMetadata.requestNonce,
		req.txTimestamp.String(),
		fallbackStr,
		timelockStr,
	)
}

// endregion /////////////////////////////////////////////////////////////////

// region OffLedger  ///////////////////////////////////////////////////////

type OffLedger struct {
	args       requestargs.RequestArgs
	chainID    *iscp.ChainID
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
var _ iscp.Request = &OffLedger{}

// NewOffLedger creates a basic request
func NewOffLedger(chainID *iscp.ChainID, contract, entryPoint iscp.Hname, args requestargs.RequestArgs) *OffLedger {
	return &OffLedger{
		chainID:    chainID,
		args:       args.Clone(),
		contract:   contract,
		entryPoint: entryPoint,
		nonce:      uint64(time.Now().UnixNano()),
	}
}

// Bytes encodes request as bytes with first type byte
func (req *OffLedger) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(offLedgerRequestType)
	req.writeToMarshalUtil(mu)
	return mu.Bytes()
}

// offLedgerFromMarshalUtil creates a request from previously serialized bytes. Does not expects type byte
func offLedgerFromMarshalUtil(mu *marshalutil.MarshalUtil) (req *OffLedger, err error) {
	req = &OffLedger{}
	if err := req.readFromMarshalUtil(mu); err != nil {
		return nil, err
	}
	return req, nil
}

func (req *OffLedger) writeToMarshalUtil(mu *marshalutil.MarshalUtil) {
	req.writeEssenceToMarshalUtil(mu)
	mu.WriteBytes(req.signature[:])
}

func (req *OffLedger) readFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
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

func (req *OffLedger) writeEssenceToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(req.chainID).
		Write(req.contract).
		Write(req.entryPoint).
		Write(req.args).
		WriteBytes(req.publicKey[:]).
		WriteUint64(req.nonce).
		Write(req.transfer)
}

func (req *OffLedger) readEssenceFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	var err error
	if req.chainID, err = iscp.ChainIDFromMarshalUtil(mu); err != nil {
		return err
	}

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
func (req *OffLedger) Hash() [32]byte {
	return hashing.HashData(req.Bytes())
}

// Sign signs essence
func (req *OffLedger) Sign(keyPair *ed25519.KeyPair) {
	req.publicKey = keyPair.PublicKey
	mu := marshalutil.New()
	req.writeEssenceToMarshalUtil(mu)
	req.signature = keyPair.PrivateKey.Sign(mu.Bytes())
}

// Tokens returns the transfers passed to the request
func (req *OffLedger) Tokens() colored.Balances {
	return req.transfer
}

// Tokens sets the transfers passed to the request
func (req *OffLedger) WithTransfer(transfer colored.Balances) *OffLedger {
	req.transfer = transfer
	return req
}

// VerifySignature verifies essence signature
func (req *OffLedger) VerifySignature() bool {
	mu := marshalutil.New()
	req.writeEssenceToMarshalUtil(mu)
	return req.publicKey.VerifySignature(mu.Bytes(), req.signature)
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (req *OffLedger) ID() (requestID iscp.RequestID) {
	txid := ledgerstate.TransactionID(hashing.HashData(req.Bytes()))
	return iscp.RequestID(ledgerstate.NewOutputID(txid, 0))
}

func (req *OffLedger) ChainID() (chainID *iscp.ChainID) {
	return req.chainID
}

// IsFeePrepaid always true for off-ledger
func (req *OffLedger) IsFeePrepaid() bool {
	return true
}

// Order number used for replay protection
func (req *OffLedger) Nonce() uint64 {
	return req.nonce
}

func (req *OffLedger) WithNonce(nonce uint64) iscp.Request {
	req.nonce = nonce
	return req
}

func (req *OffLedger) IsOffLedger() bool {
	return true
}

func (req *OffLedger) Params() (dict.Dict, bool) {
	// FIXME: this returns nil after deserializing a processed request (see tools/wasp-cli/chain/blocklog.go)
	par := req.params.Load()
	if par == nil {
		return nil, false
	}
	return par.(dict.Dict), true
}

func (req *OffLedger) SenderAccount() *iscp.AgentID {
	return iscp.NewAgentID(req.SenderAddress(), 0)
}

func (req *OffLedger) SenderAddress() ledgerstate.Address {
	if req.sender == nil {
		req.sender = ledgerstate.NewED25519Address(req.publicKey)
	}
	return req.sender
}

func (req *OffLedger) Target() iscp.RequestTarget {
	return iscp.NewRequestTarget(req.contract, req.entryPoint)
}

func (req *OffLedger) Timestamp() time.Time {
	// no request TX, return zero time
	return time.Time{}
}

func (req *OffLedger) SetParams(params dict.Dict) {
	req.params.Store(params)
}

func (req *OffLedger) Args() requestargs.RequestArgs {
	return req.args
}

func (req *OffLedger) String() string {
	return fmt.Sprintf("OffLedger::{ ID: %s, sender: %s, target: %s, entrypoint: %s, args: %s, nonce: %d }",
		req.ID().Base58(),
		req.SenderAddress().Base58(),
		req.contract.String(),
		req.entryPoint.String(),
		req.Args().String(),
		req.nonce,
	)
}

// endregion /////////////////////////////////////////////////////////////////

// SolidifiableRequest is the minimal interface required for SolidifyArgs
type SolidifiableRequest interface {
	Params() (dict.Dict, bool)
	SetParams(params dict.Dict)
	Args() requestargs.RequestArgs
}

var (
	_ SolidifiableRequest = &OnLedger{}
	_ SolidifiableRequest = &OffLedger{}
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

func RequestsInTransaction(chainID *iscp.ChainID, tx *ledgerstate.Transaction) []iscp.RequestID {
	var reqs []iscp.RequestID
	for _, out := range tx.Essence().Outputs() {
		if !out.Address().Equals(chainID.AsAddress()) {
			continue
		}
		out, ok := out.(*ledgerstate.ExtendedLockedOutput)
		if !ok {
			continue
		}
		reqs = append(reqs, iscp.RequestID(out.ID()))
	}
	return reqs
}
