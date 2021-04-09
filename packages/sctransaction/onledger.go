package sctransaction

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"golang.org/x/crypto/blake2b"
	"time"
)

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

func (req *RequestOnLedger) Output() ledgerstate.Output {
	return req.outputObj
}

func (req *RequestOnLedger) Order() uint64 {
	return uint64(req.timestamp.UnixNano())
}

// Args returns solid args if decoded already or nil otherwise
func (req *RequestOnLedger) Params() (dict.Dict, bool) {
	return req.solidArgs, req.solidArgs != nil
}

func (req *RequestOnLedger) SenderAccount() *coretypes.AgentID {
	return coretypes.NewAgentID(req.senderAddress, req.requestMetadata.SenderContract())
}

func (req *RequestOnLedger) SenderAddress() ledgerstate.Address {
	return req.senderAddress
}

// Target returns target contract and target entry point
func (req *RequestOnLedger) Target() (coretypes.Hname, coretypes.Hname) {
	return req.requestMetadata.TargetContract(), req.requestMetadata.EntryPoint()
}

func (req *RequestOnLedger) Tokens() *ledgerstate.ColoredBalances {
	return req.outputObj.Balances()
}

// coming from the transaction timestamp
func (req *RequestOnLedger) Timestamp() time.Time {
	return req.timestamp
}

func (req *RequestOnLedger) TimeLock() time.Time {
	return req.outputObj.TimeLock()
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

func (req *RequestOnLedger) Short() string {
	return req.outputObj.ID().Base58()[:6] + ".."
}

// endregion /////////////////////////////////////////////////////////////////
