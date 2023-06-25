// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// ContractFeesRecord is a structure which contains the fee information for a contract
type ContractFeesRecord struct {
	// Chain owner part of the fee. If it is 0, it means chain-global default is in effect
	OwnerFee uint64
	// Validator part of the fee. If it is 0, it means chain-global default is in effect
	ValidatorFee uint64
}

func ContractFeesRecordFromBytes(data []byte) (*ContractFeesRecord, error) {
	return rwutil.ReadFromBytes(data, new(ContractFeesRecord))
}

func (p *ContractFeesRecord) Bytes() []byte {
	return rwutil.WriteToBytes(p)
}

func (p *ContractFeesRecord) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	p.OwnerFee = rr.ReadAmount64()
	p.ValidatorFee = rr.ReadAmount64()
	return rr.Err
}

func (p *ContractFeesRecord) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteAmount64(p.OwnerFee)
	ww.WriteAmount64(p.ValidatorFee)
	return ww.Err
}
