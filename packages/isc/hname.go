// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"fmt"
	"strconv"
	"strings"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/hashing"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/subrealm"
)

// Hname is calculated as the first 4 bytes of the BLAKE2b hash of a string,
// interpreted as a little-endian uint32.
type Hname uint32

const HnameLength = 4

// HnameNil is the value used to represent a non-existent Hname.
const HnameNil = Hname(0)

// Hn calculates the hname for the given string.
// For any given string s, it is guaranteed that Hn(s) != HnameNil.
func Hn(name string) (ret Hname) {
	h := hashing.HashStrings(name)
	for i := 0; i < hashing.HashSize; i += HnameLength {
		ret, _ = HnameFromBytes(h[i : i+HnameLength])
		if ret != HnameNil {
			return ret
		}
	}
	// astronomically unlikely to end up here
	return 1
}

func HnameFromBytes(data []byte) (ret Hname, err error) {
	return bcs.Unmarshal[Hname](data)
}

func HnameFromString(s string) (Hname, error) {
	s = strings.TrimPrefix(s, "0x")
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return HnameNil, fmt.Errorf("cannot parse hname: %w", err)
	}
	return Hname(n), nil
}

func (hn Hname) Bytes() []byte {
	return bcs.MustMarshal(&hn)
}

func (hn Hname) Clone() Hname {
	return hn
}

func (hn Hname) IsNil() bool {
	return hn == HnameNil
}

func (hn Hname) String() string {
	return fmt.Sprintf("0x%08x", int(hn))
}

func ContractStateSubrealm(chainState kv.KVStore, contract Hname) kv.KVStore {
	return subrealm.New(chainState, kv.Key(contract.Bytes()))
}

func ContractStateSubrealmR(chainState kv.KVStoreReader, contract Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(contract.Bytes()))
}
