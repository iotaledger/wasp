package request

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/util"
	"go.uber.org/atomic"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"
)

const (
	OnLedgerRequestType  byte = 0
	OffLedgerRequestType byte = 1
)

func FromBytes(b []byte) (iscp.Request, error) {
	// first byte is the request type
	switch b[0] {
	case OnLedgerRequestType:
		return onLedgerFromBytes(b[1:])
	case OffLedgerRequestType:
		return offLedgerFromBytes(b[1:])
	}
	return nil, xerrors.Errorf("invalid Request Type")
}

func OffLedgerFromBytes(b []byte) (*RequestOffLedger, error) {
	return offLedgerFromBytes(b[1:])
}

// region RequestMetadata  ///////////////////////////////////////////////////////

// RequestMetadata represents content of the data payload of the output
type RequestMetadata struct {
	err            error
	senderContract iscp.Hname
	// ID of the target smart contract
	targetContract iscp.Hname
	// entry point code
	entryPoint iscp.Hname
	// request arguments, not decoded yet wrt blobRefs
	args requestargs.RequestArgs
}

func NewRequestMetadata() *RequestMetadata {
	return &RequestMetadata{
		args: requestargs.RequestArgs(dict.New()),
	}
}

//nolint:revive // TODO refactor stutter request.request
func RequestMetadataFromBytes(data []byte) *RequestMetadata {
	ret := NewRequestMetadata()
	ret.err = ret.Read(bytes.NewReader(data))
	return ret
}

func (p *RequestMetadata) WithSender(s iscp.Hname) *RequestMetadata {
	p.senderContract = s
	return p
}

func (p *RequestMetadata) WithTarget(t iscp.Hname) *RequestMetadata {
	p.targetContract = t
	return p
}

func (p *RequestMetadata) WithEntryPoint(ep iscp.Hname) *RequestMetadata {
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

func (p *RequestMetadata) SenderContract() iscp.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.senderContract
}

func (p *RequestMetadata) TargetContract() iscp.Hname {
	if !p.ParsedOk() {
		return 0
	}
	return p.targetContract
}

func (p *RequestMetadata) EntryPoint() iscp.Hname {
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
	return p.args.Write(w)
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
	return p.args.Read(r)
}

// endregion

// region RequestOnLedger //////////////////////////////////////////////////////////////////

type RequestOnLedger struct {
	nonce           uint64
	minted          map[ledgerstate.Color]uint64
	outputObj       *ledgerstate.ExtendedLockedOutput
	requestMetadata *RequestMetadata
	senderAddress   ledgerstate.Address
	params          atomic.Value // this part is mutable
}

// implements iscp.Request interface
var _ iscp.Request = &RequestOnLedger{}

// RequestOnLedgerFromOutput
//nolint:revive // TODO refactor stutter request.request
func RequestOnLedgerFromOutput(output *ledgerstate.ExtendedLockedOutput, timestamp time.Time, senderAddr ledgerstate.Address, minted ...map[ledgerstate.Color]uint64) *RequestOnLedger {
	ret := &RequestOnLedger{
		outputObj:     output,
		nonce:         uint64(timestamp.UnixNano()),
		senderAddress: senderAddr,
	}
	ret.requestMetadata = RequestMetadataFromBytes(output.GetPayload())
	ret.minted = make(map[ledgerstate.Color]uint64)
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

func (req *RequestOnLedger) ID() iscp.RequestID {
	return iscp.RequestID(req.Output().ID())
}

func (req *RequestOnLedger) IsFeePrepaid() bool {
	return false
}

func (req *RequestOnLedger) Nonce() uint64 {
	return req.nonce
}

func (req *RequestOnLedger) WithNonce(nonce uint64) iscp.Request {
	req.nonce = nonce
	return req
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

func (req *RequestOnLedger) Tokens() *ledgerstate.ColoredBalances {
	return req.outputObj.Balances()
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

func (req *RequestOnLedger) readMinted(r io.Reader) error {
	req.minted = make(map[ledgerstate.Color]uint64)
	var length uint64
	err := util.ReadUint64(r, &length)
	if err != nil {
		return err
	}
	i := uint64(0)
	for i < length {
		var colorBytes [ledgerstate.ColorLength]byte
		_, err := r.Read(colorBytes[:])
		if err != nil {
			return err
		}

		var amount int64
		err = util.ReadInt64(r, &amount)
		if err != nil {
			return err
		}

		req.minted[colorBytes] = uint64(amount)
		i++
	}
	return nil
}

func (req *RequestOnLedger) mintedBytes() []byte {
	var buf bytes.Buffer
	_ = util.WriteInt64(&buf, int64(len(req.minted)))
	for k, v := range req.minted {
		buf.Write(k.Bytes())
		_ = util.WriteInt64(&buf, int64(v))
	}
	return buf.Bytes()
}

func (req *RequestOnLedger) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteByte(OnLedgerRequestType)
	buf.Write(req.Output().Bytes())
	buf.Write(req.senderAddress.Bytes())
	buf.Write(req.mintedBytes())
	_ = util.WriteUint64(&buf, req.nonce)
	buf.Write(req.requestMetadata.Bytes())
	return buf.Bytes()
}

func onLedgerFromBytes(buf []byte) (*RequestOnLedger, error) {
	req := &RequestOnLedger{}
	r := bytes.NewReader(buf)

	// output
	outputObj, offset, err := ledgerstate.ExtendedOutputFromBytes(buf)
	if err != nil {
		return nil, err
	}
	req.outputObj = outputObj
	_, err = r.Seek(int64(offset), 0)
	if err != nil {
		return nil, err
	}

	// sender address
	addrBytes := make([]byte, ledgerstate.AddressLength)
	_, err = r.Read(addrBytes)
	if err != nil {
		return nil, err
	}
	addr, _, err := ledgerstate.AddressFromBytes(addrBytes)
	if err != nil {
		return nil, err
	}
	req.senderAddress = addr

	// minted
	err = req.readMinted(r)
	if err != nil {
		return nil, err
	}

	// timestamp
	err = util.ReadUint64(r, &req.nonce)
	if err != nil {
		return nil, err
	}

	// metadata
	metadataBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	req.requestMetadata = RequestMetadataFromBytes(metadataBytes)

	return req, nil
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
	transfer   *ledgerstate.ColoredBalances
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

// offLedgerFromBytes creates a basic request from previously serialized bytes
func offLedgerFromBytes(data []byte) (req *RequestOffLedger, err error) {
	req = &RequestOffLedger{
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
	if err = util.ReadUint64(buf, &req.nonce); err != nil {
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
	buf := bytes.NewBuffer(make([]byte, 0, 1024)) //nolint:gomnd
	_ = req.contract.Write(buf)
	_ = req.entryPoint.Write(buf)
	_ = req.args.Write(buf)
	_, _ = buf.Write(req.publicKey[:])
	_ = util.WriteUint64(buf, req.nonce)
	if req.transfer == nil {
		_ = util.WriteUint32(buf, 0)
		return buf.Bytes()
	}
	_, _ = buf.Write(req.transfer.Bytes())
	return buf.Bytes()
}

// Bytes encodes request as bytes
func (req *RequestOffLedger) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteByte(OffLedgerRequestType)
	buf.Write(req.Essence())
	buf.Write(req.signature[:])
	return buf.Bytes()
}

// only used for consensus
func (req *RequestOffLedger) Hash() [32]byte {
	return hashing.HashData(req.Bytes())
}

// Sign signs essence
func (req *RequestOffLedger) Sign(keyPair *ed25519.KeyPair) {
	req.publicKey = keyPair.PublicKey
	req.signature = keyPair.PrivateKey.Sign(req.Essence())
}

// Transfer returns the transfers passed to the request
func (req *RequestOffLedger) Transfer() *ledgerstate.ColoredBalances {
	return req.transfer
}

// Transfer sets the transfers passed to the request
func (req *RequestOffLedger) WithTransfer(transfer *ledgerstate.ColoredBalances) *RequestOffLedger {
	req.transfer = transfer
	return req
}

// VerifySignature verifies essence signature
func (req *RequestOffLedger) VerifySignature() bool {
	return req.publicKey.VerifySignature(req.Essence(), req.signature)
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

// Order number used for ordering requests in the mempool. Priority order is a descending order
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

func (req *RequestOffLedger) Tokens() *ledgerstate.ColoredBalances {
	return req.transfer
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
