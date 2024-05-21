// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"fmt"
	"io"
	"strconv"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// Hname is calculated as the first 4 bytes of the BLAKE2b hash of a string,
// interpreted as a little-endian uint32.
type Hname uint32

const HnameLength = 4

// FuncInit is a name of the init function for any smart contract
const FuncInit = "init"

// well known hnames
var (
	EntryPointInit = Hn(FuncInit)
)

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
	_, err = rwutil.ReadFromBytes(data, &ret)
	return ret, err
}

func HnameFromString(s string) (Hname, error) {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return HnameNil, fmt.Errorf("cannot parse hname: %w", err)
	}
	return Hname(n), nil
}

func (hn Hname) Bytes() []byte {
	return rwutil.WriteToBytes(&hn)
}

func (hn Hname) Clone() Hname {
	return hn
}

func (hn Hname) IsNil() bool {
	return hn == HnameNil
}

func (hn Hname) String() string {
	return fmt.Sprintf("%08x", int(hn))
}

func (hn *Hname) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	*hn = Hname(rr.ReadUint32())
	return rr.Err
}

func (hn *Hname) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteUint32(uint32(*hn))
	return ww.Err
}

func ContractStateSubrealm(chainState kv.KVStore, contract Hname) kv.KVStore {
	return subrealm.New(chainState, kv.Key(contract.Bytes()))
}

func ContractStateSubrealmR(chainState kv.KVStoreReader, contract Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(contract.Bytes()))
}
