package codec

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

func DecodeBool(b []byte, def ...bool) (bool, error) {
	if b == nil {
		if len(def) == 0 {
			return false, xerrors.Errorf("cannot decode nil bytes")
		}
		return def[0], nil
	}
	var ret bool
	err := util.ReadBoolByte(bytes.NewReader(b), &ret)
	return ret, err
}

func EncodeBool(value bool) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	err := util.WriteBoolByte(buf, value)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}
