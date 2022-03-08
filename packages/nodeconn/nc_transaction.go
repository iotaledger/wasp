// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nodeconn

import (
	serializer "github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
)

// nc_transaction maintains L1-connection related info on single message, e.g.
// promotion status, need for reattachment, etc.
type ncTransaction struct {
	tx *iotago.TransactionEssence
}

func newNCTransaction(tx *iotago.TransactionEssence) *ncTransaction {
	return &ncTransaction{
		tx: tx,
	}
}

func (nct *ncTransaction) ID() (hashing.HashValue, error) {
	ser, err := nct.tx.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return hashing.NilHash, xerrors.Errorf("Cannot create TX Hash, serialization failed: %w", err)
	}
	return hashing.HashData(ser), nil
}
