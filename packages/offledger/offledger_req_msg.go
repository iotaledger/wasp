package offledger

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"golang.org/x/xerrors"
)

type OffLedgerRequestMsg struct {
	chainID *coretypes.ChainID
	req     *request.RequestOffLedger
}

func NewOffledgerRequestMsg(chainID *coretypes.ChainID, req *request.RequestOffLedger) *OffLedgerRequestMsg {
	return &OffLedgerRequestMsg{
		chainID: chainID,
		req:     req,
	}
}

func (msg *OffLedgerRequestMsg) write(w io.Writer) error {
	if _, err := w.Write(msg.chainID.Bytes()); err != nil {
		return xerrors.Errorf("failed to write chainID: %w", err)
	}
	if _, err := w.Write(msg.req.Bytes()); err != nil {
		return xerrors.Errorf("failed to write reqDuest data")
	}
	return nil
}

func (msg *OffLedgerRequestMsg) Bytes() []byte {
	var buf bytes.Buffer
	_ = msg.write(&buf)
	return buf.Bytes()
}

func (msg *OffLedgerRequestMsg) read(r io.Reader) error {
	// read chainID
	var chainIDBytes [ledgerstate.AddressLength]byte
	_, err := r.Read(chainIDBytes[:])
	if err != nil {
		return xerrors.Errorf("failed to read chainID: %w", err)
	}
	if msg.chainID, err = coretypes.ChainIDFromBytes(chainIDBytes[:]); err != nil {
		return xerrors.Errorf("failed to read chainID: %w", err)
	}
	// read off-ledger request
	reqBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return xerrors.Errorf("failed to read request data: %w", err)
	}
	if msg.req, err = request.NewRequestOffLedgerFromBytes(reqBytes); err != nil {
		return xerrors.Errorf("failed to read request data: %w", err)
	}
	return nil
}

func OffLedgerRequestMsgFromBytes(buf []byte) (OffLedgerRequestMsg, error) {
	r := bytes.NewReader(buf)
	msg := OffLedgerRequestMsg{}
	err := msg.read(r)
	return msg, err
}
