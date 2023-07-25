// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package legacymigration

import (
	"errors"

	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/encoding/t5b1"
	"github.com/iotaledger/iota.go/math"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func addressesFromBundle(bndl bundle.Bundle) (migratedAddresses [][]byte, targetAddress iotago.Address, err error) {
	migratedAddresses = make([][]byte, 0)
	for _, tx := range bndl {
		// collect a list of ALL the legacy addresses "spending" in the bundle
		if tx.Value < 0 { // check only txs that spend (tx.Value is negative)
			legacyAddrBytes := legacyAddressBytesFromTrytes(tx.Address)
			migratedAddresses = append(migratedAddresses, legacyAddrBytes)
		}

		// there must be 1 single ed25519 address (the migration target)
		if tx.Value > 0 { // check only txs that receive (tx.Value is positive)
			ed25519AddrBytes, err := address.ParseMigrationAddress(tx.Address)
			if err != nil {
				return nil, nil, err
			}
			if targetAddress != nil {
				return nil, nil, errors.New("more than 1 target address")
			}
			var tmp iotago.Ed25519Address = ed25519AddrBytes
			targetAddress = &tmp
		}
	}
	return migratedAddresses, targetAddress, nil
}

func validBundleFromBytes(data []byte) (bundle.Bundle, error) {
	reader := rwutil.NewBytesReader(data)
	numTrytes := reader.ReadUint8()

	rawTrytes := make([]string, numTrytes)

	for i := 0; i < int(numTrytes); i++ {
		txTrytesBytesLen := reader.ReadUint16()
		txTrytesBytes := make([]byte, txTrytesBytesLen)
		reader.ReadN(txTrytesBytes)
		txTrytes := string(txTrytesBytes) // trytes are encoded as UTF-8
		if err := trinary.ValidTrytes(txTrytes); err != nil {
			return nil, err
		}
		rawTrytes[i] = txTrytes
	}
	if reader.Err != nil {
		return nil, reader.Err
	}

	txs, err := transaction.AsTransactionObjects(rawTrytes, nil)
	if err != nil {
		return nil, err
	}
	// validate the bundle, do migration checks
	if err := bundle.ValidBundle(txs, true); err != nil {
		return nil, err
	}

	// extra check from hornet legacy: https://github.com/iotaledger/hornet/blob/legacy/pkg/compressed/tx.go#L103-L113
	for _, tx := range txs {
		tx := tx // avoid implicit memory aliasing in for loop
		if tx.Value != 0 {
			trits, err := transaction.TransactionToTrits(&tx)
			if err != nil {
				return nil, err
			}
			if trits[consts.AddressTrinaryOffset+consts.AddressTrinarySize-1] != 0 {
				// The last trit is always zero because of KERL/keccak
				return nil, consts.ErrInvalidAddress
			}

			if math.AbsInt64(tx.Value) > consts.TotalSupply {
				return nil, consts.ErrInsufficientBalance
			}
		}
	}

	return txs, nil
}

// ------- taken from legacy hornet
const (
	hashTrytesSize = consts.HashTrytesSize
	tagTrytesSize  = consts.TagTrinarySize / consts.TritsPerTryte
)

// legacyAddressBytesFromTrytes returns the binary representation of the given address trytes.
// It panics when trytes hash invalid length.
func legacyAddressBytesFromTrytes(trytes trinary.Trytes) []byte {
	if len(trytes) != hashTrytesSize && len(trytes) != consts.AddressWithChecksumTrytesSize {
		panic("invalid address length")
	}
	return t5b1.EncodeTrytes(trytes[:hashTrytesSize])
}
