// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package sctransaction implements smart contract transaction.
// smart contract transaction is value transaction with special payload
package sctransaction_old

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/tangle/payload"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
	"io"
)

// TransactionEssence represents essence of ISCP transaction.
// ISCP transaction is a value transaction as defined by Goshimmer's ledgerstate.TransactionEssence
// Its payload contains metadata of the ISCP transaction essence.
// The payload is parsed and the checked validity of its semantic. The parsed payload contains information
// how elements of the overall value transaction are interpreted
type TransactionEssence struct {
	// value transaction with metadata as a payload (payload.GenericDataPayload)
	*ledgerstate.TransactionEssence

	// ------ metadata of the value transaction
	// if stateSection != nil, state section of the state anchor.
	// if stateSection == nil, the transaction is not a state anchor (it may contain requests)
	stateSection *StateSection
	// requestSection list of request sections.
	// it stateSection == nil, requestSection must contain at least 1 request section
	requestSection []*RequestSection

	// cachedProperties cached data of semantic analysis, for faster access
	cachedProperties coretypes.SCTransactionProperties
}

// NewTransactionEssence creates new sc transaction. It takes value transaction,
// appends metadata to, serialized metadata and put it into the payload
func NewTransactionEssence(vtx *ledgerstate.TransactionEssence, stateBlock *StateSection, requestBlocks []*RequestSection) (*TransactionEssence, error) {
	ret := &TransactionEssence{
		TransactionEssence: vtx,
		stateSection:       stateBlock,
		requestSection:     requestBlocks,
	}
	var buf bytes.Buffer
	if err := ret.writeDataPayload(&buf); err != nil {
		return nil, err
	}
	vtx.SetPayload(payload.NewGenericDataPayload(buf.Bytes()))
	return ret, nil
}

// ParseValueTransaction parses dataPayload of the value transaction performs semantic analysis
func ParseValueTransaction(vtx *ledgerstate.TransactionEssence) (*TransactionEssence, error) {
	// parse data payload as smart contract metadata
	p, ok := vtx.Payload().(*payload.GenericDataPayload)
	if !ok {
		return nil, xerrors.New("wrong payload type")
	}
	rdr := bytes.NewReader(p.Blob())
	ret := &TransactionEssence{TransactionEssence: vtx}
	if err := ret.readDataPayload(rdr); err != nil {
		return nil, err
	}
	// semantic validation
	if _, err := ret.Properties(); err != nil {
		return nil, err
	}
	return ret, nil
}

// Properties returns valid properties if sc transaction is semantically correct
func (tx *TransactionEssence) Properties() (coretypes.SCTransactionProperties, error) {
	if tx.cachedProperties != nil {
		return tx.cachedProperties, nil
	}
	var err error
	tx.cachedProperties, err = calcProperties(tx)
	return tx.cachedProperties, err
}

// MustProperties returns valid properties if sc transaction is semantically correct or panics otherwise
func (tx *TransactionEssence) MustProperties() coretypes.SCTransactionProperties {
	ret, err := tx.Properties()
	if err != nil {
		panic(err)
	}
	return ret
}

// State returns state section and existence flag
func (tx *TransactionEssence) State() (*StateSection, bool) {
	return tx.stateSection, tx.stateSection != nil
}

// MustState returns state section or panics if it does not exist
func (tx *TransactionEssence) MustState() *StateSection {
	if tx.stateSection == nil {
		panic("MustState: state block expected")
	}
	return tx.stateSection
}

// Requests returns requests
func (tx *TransactionEssence) Requests() []*RequestSection {
	return tx.requestSection
}

// function writes bytes of the SC transaction-specific part
func (tx *TransactionEssence) writeDataPayload(w io.Writer) error {
	if tx.stateSection == nil && len(tx.requestSection) == 0 {
		return errors.New("can't encode empty chain transaction")
	}
	if len(tx.requestSection) > 127 {
		return errors.New("max number of request sections 127 exceeded")
	}
	numRequests := byte(len(tx.requestSection))
	b, err := encodeMetaByte(tx.stateSection != nil, numRequests)
	if err != nil {
		return err
	}
	if err = util.WriteByte(w, b); err != nil {
		return err
	}
	if tx.stateSection != nil {
		if err := tx.stateSection.Write(w); err != nil {
			return err
		}
	}
	for _, reqBlk := range tx.requestSection {
		if err := reqBlk.Write(w); err != nil {
			return err
		}
	}
	return nil
}

// readDataPayload parses data stream of data payload to value transaction as smart contract meta data
func (tx *TransactionEssence) readDataPayload(r io.Reader) error {
	var hasState bool
	var numRequests byte
	if b, err := util.ReadByte(r); err != nil {
		return err
	} else {
		hasState, numRequests = decodeMetaByte(b)
	}
	var stateBlock *StateSection
	if hasState {
		stateBlock = &StateSection{}
		if err := stateBlock.Read(r); err != nil {
			return err
		}
	}
	reqBlks := make([]*RequestSection, numRequests)
	for i := range reqBlks {
		reqBlks[i] = &RequestSection{}
		if err := reqBlks[i].Read(r); err != nil {
			return err
		}
	}
	tx.stateSection = stateBlock
	tx.requestSection = reqBlks
	return nil
}

func (tx *TransactionEssence) String() string {
	ret := ""
	stateBlock, ok := tx.State()
	if ok {
		vh := stateBlock.StateHash()
		ret += fmt.Sprintf("State: color: %s statehash: %s, tx-ts: %s\n",
			stateBlock.Color().String(),
			vh.String(), tx.TransactionEssence.Timestamp(),
		)
	} else {
		ret += "State: none\n"
	}
	for i, reqBlk := range tx.Requests() {
		addr := reqBlk.Target()
		ret += fmt.Sprintf("Req #%d: addr: %s code: %s\n", i,
			util.Short(addr.String()), reqBlk.EntryPointCode().String())
	}
	return ret
}
