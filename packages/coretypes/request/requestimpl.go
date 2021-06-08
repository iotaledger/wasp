package request

import (
	"bytes"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/crypto/blake2b"
)

// region RequestMetadata  ///////////////////////////////////////////////////////

// RequestMetadata represents content of the data payload of the output
type RequestMetadata struct {
	err            error
	senderContract coretypes.Hname
	// ID of the target smart contract
	targetContract coretypes.Hname
	// entry point code
	entryPoint coretypes.Hname
	// request arguments, not decoded yet wrt blobRefs
	args requestargs.RequestArgs
}

func NewRequestMetadata() *RequestMetadata {
	return &RequestMetadata{
		args: requestargs.RequestArgs(dict.New()),
	}
}

func RequestMetadataFromBytes(data []byte) *RequestMetadata {
	ret := NewRequestMetadata()
	ret.err = ret.Read(bytes.NewReader(data))
	return ret
}

func (p *RequestMetadata) WithSender(s coretypes.Hname) *RequestMetadata {
	p.senderContract = s
	return p
}

func (p *RequestMetadata) WithTarget(t coretypes.Hname) *RequestMetadata {
	p.targetContract = t
	return p
}

func (p *RequestMetadata) WithEntryPoint(ep coretypes.Hname) *RequestMetadata {
	p.entryPoint = ep
	return p
}

func (p *RequestMetadata) WithArgs(args requestargs.RequestArgs) *RequestMetadata {
	p.args = args.Clone()
	return p
}

func (p *RequestMetadata) Clone() *RequestMetadata {
	ret := *p
	ret.args = p.args.Clone()
	return &ret
}

func (p *RequestMetadata) ParsedOk() bool {
	return p.err == nil
}

func (p *RequestMetadata) ParsedError() error {
	return p.err
}

func (p *RequestMetadata) SenderContract() coretypes.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.senderContract
}

func (p *RequestMetadata) TargetContract() coretypes.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.targetContract
}

func (p *RequestMetadata) EntryPoint() coretypes.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.entryPoint
}

func (p *RequestMetadata) Args() requestargs.RequestArgs {
	if !p.ParsedOk() {
		return requestargs.RequestArgs(dict.New())
	}
	return p.args
}

func (p *RequestMetadata) Bytes() []byte {
	var buf bytes.Buffer
	_ = p.Write(&buf)
	return buf.Bytes()
}

func (p *RequestMetadata) Write(w io.Writer) error {
	if err := p.senderContract.Write(w); err != nil {
		return err
	}
	if err := p.targetContract.Write(w); err != nil {
		return err
	}
	if err := p.entryPoint.Write(w); err != nil {
		return err
	}
	if err := p.args.Write(w); err != nil {
		return err
	}
	return nil
}

func (p *RequestMetadata) Read(r io.Reader) error {
	if err := p.senderContract.Read(r); err != nil {
		return err
	}
	if err := p.targetContract.Read(r); err != nil {
		return err
	}
	if err := p.entryPoint.Read(r); err != nil {
		return err
	}
	if err := p.args.Read(r); err != nil {
		return err
	}
	return nil
}

// endregion

// region RequestOnLedger //////////////////////////////////////////////////////////////////

type RequestOnLedger struct {
	timestamp       time.Time
	minted          map[ledgerstate.Color]uint64
	outputObj       *ledgerstate.ExtendedLockedOutput
	requestMetadata *RequestMetadata
	senderAddress   ledgerstate.Address
	solidArgs       dict.Dict
}

// implements coretypes.Request interface
var _ coretypes.Request = &RequestOnLedger{}

// RequestOnLedgerFromOutput
func RequestOnLedgerFromOutput(output *ledgerstate.ExtendedLockedOutput, timestamp time.Time, senderAddr ledgerstate.Address, minted ...map[ledgerstate.Color]uint64) *RequestOnLedger {
	ret := &RequestOnLedger{
		outputObj:     output,
		timestamp:     timestamp,
		senderAddress: senderAddr,
	}
	ret.requestMetadata = RequestMetadataFromBytes(output.GetPayload())
	ret.minted = make(map[ledgerstate.Color]uint64, 0)
	if len(minted) > 0 {
		for k, v := range minted[0] {
			ret.minted[k] = v
		}
	}
	return ret
}

// RequestsOnLedgerFromTransaction creates RequestOnLedger object from transaction and output index
func RequestsOnLedgerFromTransaction(tx *ledgerstate.Transaction, targetAddr ledgerstate.Address) ([]*RequestOnLedger, error) {
	senderAddr, err := utxoutil.GetSingleSender(tx)
	if err != nil {
		return nil, err
	}
	mintedAmounts := utxoutil.GetMintedAmounts(tx)
	ret := make([]*RequestOnLedger, 0)
	for _, o := range tx.Essence().Outputs() {
		if out, ok := o.(*ledgerstate.ExtendedLockedOutput); ok {
			if out.Address().Equals(targetAddr) {
				out1 := out.UpdateMintingColor().(*ledgerstate.ExtendedLockedOutput)
				ret = append(ret, RequestOnLedgerFromOutput(out1, tx.Essence().Timestamp(), senderAddr, mintedAmounts))
			}
		}
	}
	return ret, nil
}

func (req *RequestOnLedger) ID() coretypes.RequestID {
	return coretypes.RequestID(req.Output().ID())
}

func (req *RequestOnLedger) IsFeePrepaid() bool {
	return false
}

func (req *RequestOnLedger) Nonce() uint64 {
	return uint64(req.timestamp.UnixNano())
}

func (req *RequestOnLedger) Output() ledgerstate.Output {
	return req.outputObj
}

// Params returns solid args if decoded already or nil otherwise
func (req *RequestOnLedger) Params() (dict.Dict, bool) {
	return req.solidArgs, req.solidArgs != nil
}

func (req *RequestOnLedger) SenderAccount() *coretypes.AgentID {
	return coretypes.NewAgentID(req.senderAddress, req.requestMetadata.SenderContract())
}

func (req *RequestOnLedger) SenderAddress() ledgerstate.Address {
	return req.senderAddress
}

// SolidifyArgs return true if solidified successfully
func (req *RequestOnLedger) SolidifyArgs(reg coretypes.BlobCache) (bool, error) {
	if req.solidArgs != nil {
		return true, nil
	}
	solid, ok, err := req.requestMetadata.Args().SolidifyRequestArguments(reg)
	if err != nil || !ok {
		return ok, err
	}
	req.solidArgs = solid
	if req.solidArgs == nil {
		panic("req.solidArgs == nil")
	}
	return true, nil
}

// Target returns target contract and target entry point
func (req *RequestOnLedger) Target() (coretypes.Hname, coretypes.Hname) {
	return req.requestMetadata.TargetContract(), req.requestMetadata.EntryPoint()
}

func (req *RequestOnLedger) TimeLock() time.Time {
	return req.outputObj.TimeLock()
}

func (req *RequestOnLedger) Tokens() *ledgerstate.ColoredBalances {
	return req.outputObj.Balances()
}

// coming from the transaction timestamp
func (req *RequestOnLedger) Timestamp() time.Time {
	return req.timestamp
}

func (req *RequestOnLedger) SetMetadata(d *RequestMetadata) {
	req.requestMetadata = d.Clone()
}

func (req *RequestOnLedger) GetMetadata() *RequestMetadata {
	return req.requestMetadata
}

func (req *RequestOnLedger) MintColor() ledgerstate.Color {
	return blake2b.Sum256(req.Output().ID().Bytes())
}

func (req *RequestOnLedger) MintedAmounts() map[ledgerstate.Color]uint64 {
	return req.minted
}

func (req *RequestOnLedger) Short() string {
	return req.outputObj.ID().Base58()[:6] + ".."
}

// endregion /////////////////////////////////////////////////////////////////

// region RequestOffLedger  ///////////////////////////////////////////////////////

type RequestOffLedger struct {
	args       requestargs.RequestArgs
	contract   coretypes.Hname
	entryPoint coretypes.Hname
	params     dict.Dict
	publicKey  ed25519.PublicKey
	sender     ledgerstate.Address
	signature  ed25519.Signature
	timestamp  time.Time
	transfer   *ledgerstate.ColoredBalances
}

// implements coretypes.Request interface
var _ coretypes.Request = &RequestOffLedger{}

// NewRequestOffLedger creates a basic request
func NewRequestOffLedger(contract coretypes.Hname, entryPoint coretypes.Hname, args requestargs.RequestArgs) *RequestOffLedger {
	return &RequestOffLedger{
		args:       args.Clone(),
		contract:   contract,
		entryPoint: entryPoint,
		timestamp:  time.Now(),
	}
}

// NewRequestOffLedgerFromBytes creates a basic request from previously serialized bytes
func NewRequestOffLedgerFromBytes(data []byte) (request *RequestOffLedger, err error) {
	req := &RequestOffLedger{
		args: requestargs.New(nil),
	}
	buf := bytes.NewBuffer(data)
	if err = req.contract.Read(buf); err != nil {
		return
	}
	if err = req.entryPoint.Read(buf); err != nil {
		return
	}
	if err = req.args.Read(buf); err != nil {
		return
	}
	var n int
	n, err = buf.Read(req.publicKey[:])
	if err != nil || n != len(req.publicKey) {
		return nil, io.EOF
	}
	if err = util.ReadTime(buf, &req.timestamp); err != nil {
		return
	}
	var colors uint32
	if err = util.ReadUint32(buf, &colors); err != nil {
		return
	}
	if colors != 0 {
		balances := make(map[ledgerstate.Color]uint64)
		for i := uint32(0); i < colors; i++ {
			var color ledgerstate.Color
			n, err = buf.Read(color[:])
			if err != nil || n != len(color) {
				return nil, io.EOF
			}
			var balance uint64
			if err = util.ReadUint64(buf, &balance); err != nil {
				return
			}
			balances[color] = balance
		}
		req.transfer = ledgerstate.NewColoredBalances(balances)
	}
	n, err = buf.Read(req.signature[:])
	if err != nil || n != len(req.signature) {
		return nil, io.EOF
	}
	return req, nil
}

// Essence encodes request essence as bytes
func (req *RequestOffLedger) Essence() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	_ = req.contract.Write(buf)
	_ = req.entryPoint.Write(buf)
	_ = req.args.Write(buf)
	_, _ = buf.Write(req.publicKey[:])
	_ = util.WriteTime(buf, req.timestamp)
	if req.transfer == nil {
		_ = util.WriteUint32(buf, 0)
		return buf.Bytes()
	}
	_, _ = buf.Write(req.transfer.Bytes())
	return buf.Bytes()
}

// Bytes encodes request as bytes
func (req *RequestOffLedger) Bytes() []byte {
	return append(req.Essence(), req.signature[:]...)
}

// Sign signs essence
func (req *RequestOffLedger) Sign(keyPair *ed25519.KeyPair) {
	req.publicKey = keyPair.PublicKey
	req.signature = keyPair.PrivateKey.Sign(req.Essence())
}

// Transfer returns transfers passed to request
func (req *RequestOffLedger) Transfer(transfer *ledgerstate.ColoredBalances) {
	req.transfer = transfer
}

// VerifySignature verifies essence signature
func (req *RequestOffLedger) VerifySignature() bool {
	return req.publicKey.VerifySignature(req.Essence(), req.signature)
}

// ID returns request id for this request
// index part of request id is always 0 for off ledger requests
// note that request needs to have been signed before this value is
// considered valid
func (req *RequestOffLedger) ID() (requestId coretypes.RequestID) {
	txid := ledgerstate.TransactionID(hashing.HashData(req.Bytes()))
	return coretypes.RequestID(ledgerstate.NewOutputID(txid, 0))
}

// IsFeePrepaid always true for off-ledger
func (req *RequestOffLedger) IsFeePrepaid() bool {
	return true
}

// Order number used for ordering requests in the mempool. Priority order is a descending order
func (req *RequestOffLedger) Nonce() uint64 {
	return uint64(req.timestamp.UnixNano())
}

// Output nil for off-ledger requests
func (req *RequestOffLedger) Output() ledgerstate.Output {
	return nil
}

func (req *RequestOffLedger) Params() (dict.Dict, bool) {
	return req.params, req.params != nil
}

func (req *RequestOffLedger) SenderAccount() *coretypes.AgentID {
	return coretypes.NewAgentID(req.SenderAddress(), 0)
}

func (req *RequestOffLedger) SenderAddress() ledgerstate.Address {
	if req.sender == nil {
		req.sender = ledgerstate.NewED25519Address(req.publicKey)
	}
	return req.sender
}

// SolidifyArgs return true if solidified successfully
func (req *RequestOffLedger) SolidifyArgs(reg coretypes.BlobCache) (bool, error) {
	if req.params != nil {
		return true, nil
	}
	solid, ok, err := req.args.SolidifyRequestArguments(reg)
	if err != nil || !ok {
		return ok, err
	}
	req.params = solid
	if req.params == nil {
		panic("req.solidArgs == nil")
	}
	return true, nil
}

func (req *RequestOffLedger) Target() (coretypes.Hname, coretypes.Hname) {
	return req.contract, req.entryPoint
}

// TimeLock returns time lock time or zero time if no time lock
func (req *RequestOffLedger) TimeLock() time.Time {
	// no time lock, return zero time
	return time.Time{}
}

func (req *RequestOffLedger) Tokens() *ledgerstate.ColoredBalances {
	return req.transfer
}

// endregion /////////////////////////////////////////////////////////////////
