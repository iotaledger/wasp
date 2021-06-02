package util

import (
	"encoding"
	"encoding/binary"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/pkg/errors"
)

func WriteByte(w io.Writer, val byte) error {
	b := []byte{val}
	_, err := w.Write(b[:])
	return err
}

func ReadByte(r io.Reader) (byte, error) {
	var b [1]byte
	_, err := r.Read(b[:])
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func Uint16To2Bytes(val uint16) []byte {
	var tmp2 [2]byte
	binary.LittleEndian.PutUint16(tmp2[:], val)
	return tmp2[:]
}

func MustUint16From2Bytes(b []byte) uint16 {
	if len(b) != 2 {
		panic("len(b) != 2")
	}
	return binary.LittleEndian.Uint16(b[:])
}

func Uint32To4Bytes(val uint32) []byte {
	var tmp4 [4]byte
	binary.LittleEndian.PutUint32(tmp4[:], val)
	return tmp4[:]
}

func Uint32From4Bytes(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, errors.New("len(b) != 4")
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func MustUint32From4Bytes(b []byte) uint32 {
	n, err := Uint32From4Bytes(b)
	if err != nil {
		panic(err)
	}
	return n
}

func MustUint64From8Bytes(b []byte) uint64 {
	if len(b) != 8 {
		panic("len(b) != 8")
	}
	return binary.LittleEndian.Uint64(b[:])
}

func Uint64From8Bytes(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, errors.New("len(b) != 8")
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

func Int64From8Bytes(b []byte) (int64, error) {
	ret, err := Uint64From8Bytes(b)
	if err != nil {
		return 0, err
	}
	return int64(ret), nil
}

func Uint64To8Bytes(val uint64) []byte {
	var tmp8 [8]byte
	binary.LittleEndian.PutUint64(tmp8[:], val)
	return tmp8[:]
}

func Int64To8Bytes(val int64) []byte {
	return Uint64To8Bytes(uint64(val))
}

func WriteUint16(w io.Writer, val uint16) error {
	_, err := w.Write(Uint16To2Bytes(val))
	return err
}

func ReadUint16(r io.Reader, pval *uint16) error {
	var tmp2 [2]byte
	_, err := r.Read(tmp2[:])
	if err != nil {
		return err
	}
	*pval = binary.LittleEndian.Uint16(tmp2[:])
	return nil
}

func WriteUint32(w io.Writer, val uint32) error {
	_, err := w.Write(Uint32To4Bytes(val))
	return err
}

func ReadUint32(r io.Reader, pval *uint32) error {
	var tmp4 [4]byte
	_, err := r.Read(tmp4[:])
	if err != nil {
		return err
	}
	*pval = MustUint32From4Bytes(tmp4[:])
	return nil
}

func WriteUint64(w io.Writer, val uint64) error {
	_, err := w.Write(Uint64To8Bytes(val))
	return err
}

func ReadUint64(r io.Reader, pval *uint64) error {
	var tmp8 [8]byte
	_, err := r.Read(tmp8[:])
	if err != nil {
		return err
	}
	*pval = binary.LittleEndian.Uint64(tmp8[:])
	return nil
}

func WriteInt64(w io.Writer, val int64) error {
	_, err := w.Write(Uint64To8Bytes(uint64(val)))
	return err
}

func ReadInt64(r io.Reader, pval *int64) error {
	var tmp8 [8]byte
	_, err := r.Read(tmp8[:])
	if err != nil {
		return err
	}
	uval := binary.LittleEndian.Uint64(tmp8[:])
	*pval = int64(uval)
	return nil
}

const MaxUint16 = int(^uint16(0))

func WriteBytes16(w io.Writer, data []byte) error {
	if len(data) > MaxUint16 {
		panic("WriteBytes16: too long data")
	}
	err := WriteUint16(w, uint16(len(data)))
	if err != nil {
		return err
	}
	if len(data) != 0 {
		_, err = w.Write(data)
	}
	return err
}

func ReadBytes16(r io.Reader) ([]byte, error) {
	var length uint16
	err := ReadUint16(r, &length)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return []byte{}, nil
	}
	ret := make([]byte, length)
	_, err = r.Read(ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

const MaxUint32 = int(^uint32(0))

func WriteBytes32(w io.Writer, data []byte) error {
	if len(data) > MaxUint32 {
		panic("WriteBytes32: too long data")
	}
	err := WriteUint32(w, uint32(len(data)))
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func ReadBytes32(r io.Reader) ([]byte, error) {
	var length uint32
	err := ReadUint32(r, &length)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return []byte{}, nil
	}
	ret := make([]byte, length)
	_, err = r.Read(ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func WriteBoolByte(w io.Writer, cond bool) error {
	var b [1]byte
	if cond {
		b[0] = 0xFF
	}
	_, err := w.Write(b[:])
	return err
}

func ReadBoolByte(r io.Reader, cond *bool) error {
	var b [1]byte
	_, err := r.Read(b[:])
	if err != nil {
		return err
	}
	*cond = b[0] == 0xFF
	if !*cond && b[0] != 0x00 {
		return errors.New("ReadBoolByte: unexpected value")
	}
	return nil
}

func WriteTime(w io.Writer, ts time.Time) error {
	return WriteUint64(w, uint64(ts.UnixNano()))
}

func ReadTime(r io.Reader, ts *time.Time) error {
	var nano uint64
	err := ReadUint64(r, &nano)
	if err != nil {
		return err
	}
	*ts = time.Unix(0, int64(nano))
	return nil
}

func WriteString16(w io.Writer, str string) error {
	return WriteBytes16(w, []byte(str))
}

func ReadString16(r io.Reader) (string, error) {
	ret, err := ReadBytes16(r)
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func WriteStrings16(w io.Writer, strs []string) error {
	if len(strs) > MaxUint16 {
		panic("WriteStrings16: too long array")
	}
	if err := WriteUint16(w, uint16(len(strs))); err != nil {
		return err
	}
	for _, s := range strs {
		if err := WriteString16(w, s); err != nil {
			return err
		}
	}
	return nil
}

func ReadStrings16(r io.Reader) ([]string, error) {
	var size uint16
	if err := ReadUint16(r, &size); err != nil {
		return nil, nil
	}
	ret := make([]string, size)
	var err error
	for i := range ret {
		if ret[i], err = ReadString16(r); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func ReadColor(r io.Reader, color *ledgerstate.Color) error {
	n, err := r.Read(color[:])
	if err != nil {
		return err
	}
	if n != ledgerstate.ColorLength {
		return errors.New("error while reading color code")
	}
	return nil
}

func ReadHashValue(r io.Reader, h *hashing.HashValue) error {
	n, err := r.Read(h[:])
	if err != nil {
		return err
	}
	if n != hashing.HashSize {
		return errors.New("error while reading hash")
	}
	return nil
}

// WriteMarshaled supports kyber.Point, kyber.Scalar and similar.
func WriteMarshaled(w io.Writer, val encoding.BinaryMarshaler) error {
	var err error
	var bin []byte
	if bin, err = val.MarshalBinary(); err != nil {
		return err
	}
	if err = WriteBytes16(w, bin); err != nil {
		return err
	}
	return nil
}

// ReadMarshaled supports kyber.Point, kyber.Scalar and similar.
func ReadMarshaled(r io.Reader, val encoding.BinaryUnmarshaler) error {
	var err error
	var bin []byte
	if bin, err = ReadBytes16(r); err != nil {
		return err
	}
	if err = val.UnmarshalBinary(bin); err != nil {
		return err
	}
	return nil
}
