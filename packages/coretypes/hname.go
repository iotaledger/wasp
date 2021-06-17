// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/pkg/errors"
	"golang.org/x/xerrors"
)

// Hname is 4 bytes of blake2b hash of any string interpreted as little-endian uint32.
// 0 and not ^0 are reserved values and the coretypes.Hn ensures it is not returned
type Hname uint32

const HnameLength = 4

// FuncInit is a name of the init function for any smart contract
const FuncInit = "init"

// well known hnames
var (
	EntryPointInit = Hn(FuncInit)
	HnameRoot      = Hn("root")
	HnameAccounts  = Hn("accounts")
	HnameBlob      = Hn("blob")
	HnameEventlog  = Hn("eventlog")
	HnameDefault   = Hname(0)
)

// HnameFromBytes constructor, unmarshalling
func HnameFromBytes(data []byte) (ret Hname, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// Hn create hname from arbitrary string.
func Hn(funname string) (ret Hname) {
	h := hashing.HashStrings(funname)
	_ = ret.Read(bytes.NewReader(h[:HnameLength]))
	if ret == 0 || ret == Hname(^uint32(0)) {
		// ensure 0 and ^0 are impossible
		_ = ret.Read(bytes.NewReader(h[HnameLength : 2*HnameLength]))
	}
	return ret
}

func (hn Hname) Bytes() []byte {
	ret := make([]byte, HnameLength)
	binary.LittleEndian.PutUint32(ret, (uint32)(hn))
	return ret
}

func (hn Hname) Clone() Hname {
	return hn
}

func (hn Hname) String() string {
	return fmt.Sprintf("%08x", (int)(hn))
}

func HnameFromString(s string) (Hname, error) {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, errors.Wrap(err, "cannot parse hname")
	}
	return Hname(n), nil
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
		return xerrors.New("wrong data length")
	}
	t := binary.LittleEndian.Uint32(b[:])
	*hn = (Hname)(t)
	return nil
}
