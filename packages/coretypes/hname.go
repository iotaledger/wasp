package coretypes

import (
	"bytes"
	"encoding/binary"
	"github.com/iotaledger/wasp/packages/hashing"
	"io"
	"strconv"
)

type Hname uint32

const HnameLength = 4

const FuncInit = "init"

var EntryPointInit = Hn(FuncInit)

func NewHnameFromBytes(data []byte) (ret Hname, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// Hn beware collisions: hash is only 4 bytes!
// must always be checked against the whole table for collisions and adjusted
func Hn(funname string) (ret Hname) {
	_ = ret.Read(bytes.NewReader(hashing.HashStrings(funname)[:HnameLength]))
	return ret
}

func (hn Hname) Bytes() []byte {
	ret := make([]byte, HnameLength)
	binary.LittleEndian.PutUint32(ret, (uint32)(hn))
	return ret
}

func (hn Hname) String() string {
	return strconv.Itoa((int)(hn))
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
		return ErrWrongDataLength
	}
	t := binary.LittleEndian.Uint32(b[:])
	*hn = (Hname)(t)
	return nil
}
