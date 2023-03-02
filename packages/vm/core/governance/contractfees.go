// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

// ContractFeesRecord is a structure which contains the fee information for a contract
type ContractFeesRecord struct {
	// Chain owner part of the fee. If it is 0, it means chain-global default is in effect
	OwnerFee uint64
	// Validator part of the fee. If it is 0, it means chain-global default is in effect
	ValidatorFee uint64
}

func ContractFeesRecordFromMarshalUtil(mu *marshalutil.MarshalUtil) (*ContractFeesRecord, error) {
	ret := &ContractFeesRecord{}
	var err error
	if ret.OwnerFee, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	if ret.ValidatorFee, err = mu.ReadUint64(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (p *ContractFeesRecord) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteUint64(p.OwnerFee)
	mu.WriteUint64(p.ValidatorFee)
	return mu.Bytes()
}
