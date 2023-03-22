// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

// Hname is calculated as the first 4 bytes of the blake2b hash of a string, interpreted as
// a little-endian uint32.
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

// HnameFromBytes constructor, unmarshalling
func HnameFromMarshalUtil(mu *marshalutil.MarshalUtil) (ret Hname, err error) {
	err = ret.ReadFromMarshalUtil(mu)
	return
}

func HnameFromBytes(data []byte) (ret Hname, err error) {
	ret, err = HnameFromMarshalUtil(marshalutil.New(data))
	return
}

// Hn calculates the hname for the given string.
// For any given string s, it is guaranteed that Hn(s) != HnaneNil.
func Hn(name string) (ret Hname) {
	h := hashing.HashStrings(name)
	for i := byte(0); i < hashing.HashSize; i += HnameLength {
		_ = ret.Read(bytes.NewReader(h[i : i+HnameLength]))
		if ret != HnameNil {
			return ret
		}
	}
	// astronomically unlikely to end up here
	return 1
}

func (hn Hname) IsNil() bool {
	return hn == HnameNil
}

func (hn Hname) Bytes() []byte {
	ret := make([]byte, HnameLength)
	binary.LittleEndian.PutUint32(ret, uint32(hn))
	return ret
}

func (hn Hname) Clone() Hname {
	return hn
}

func (hn Hname) String() string {
	return fmt.Sprintf("%08x", int(hn))
}

func HnameFromHexString(s string) (Hname, error) {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("cannot parse hname: %w", err)
	}
	return Hname(n), nil
}

func (hn *Hname) WriteToMarshalUtil(mu *marshalutil.MarshalUtil) {
	mu.Write(hn)
}

func (hn *Hname) ReadFromMarshalUtil(mu *marshalutil.MarshalUtil) error {
	b, err := mu.ReadBytes(HnameLength)
	if err != nil {
		return err
	}
	*hn = Hname(binary.LittleEndian.Uint32(b))
	return nil
}

func (hn *Hname) Write(w io.Writer) error {
	_, err := w.Write(hn.Bytes())
	return err
}

func (hn *Hname) Read(r io.Reader) error {
	var b [HnameLength]byte
	n, err := r.Read(b[:])
	if err != nil {
		return err
	}
	if n != HnameLength {
		return errors.New("wrong data length")
	}
	t := binary.LittleEndian.Uint32(b[:])
	*hn = Hname(t)
	return nil
}
