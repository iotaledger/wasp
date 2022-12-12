package trie

import (
	"encoding/hex"
)

// unpack16 src places each 4 bit nibble into separate byte
func unpack16(dst, src []byte) []byte {
	for _, c := range src {
		dst = append(dst, c>>4, c&0x0F)
	}
	return dst
}

// pack16 places to 4 bit nibbles into one byte. I case of odd number of nibbles,
// the 4 lower bit in the last byte remains 0
func pack16(dst, src []byte) ([]byte, error) {
	for i := 0; i < len(src); i += 2 {
		if src[i] > 0x0F {
			return nil, ErrWrongNibble
		}
		c := src[i] << 4
		if i+1 < len(src) {
			c |= src[i+1]
		}
		dst = append(dst, c)
	}
	return dst, nil
}

// encode16 packs nibbles and prefixes it with number of excess bytes (0 or 1)
func encode16(k16 []byte) ([]byte, error) {
	ret := append(make([]byte, 0, len(k16)/2+1), byte(len(k16)%2))
	return pack16(ret, k16)
}

func decode16(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmpty
	}
	if data[0] > 1 {
		return nil, ErrWrongFormat
	}
	ret := make([]byte, 0, len(data)*2)
	ret = unpack16(ret, data[1:])
	if data[0] == 1 && ret[len(ret)-1] != 0 {
		// enforce padding with 0
		return nil, ErrWrongFormat
	}
	// cut the excess byte
	ret = ret[:len(ret)-int(data[0])]
	return ret, nil
}

func unpackBytes(src []byte) []byte {
	return unpack16(make([]byte, 0, 2*len(src)), src)
}

func encodeUnpackedBytes(unpacked []byte) ([]byte, error) {
	if len(unpacked) == 0 {
		return nil, nil
	}
	return encode16(unpacked)
}

func packUnpackedBytes(unpacked []byte) ([]byte, error) {
	if len(unpacked) == 0 {
		return nil, nil
	}
	ret, err := encode16(unpacked)
	if err != nil {
		return nil, err
	}
	return ret[1:], nil
}

func mustEncodeUnpackedBytes(unpacked []byte) []byte {
	ret, err := encodeUnpackedBytes(unpacked)
	assertf(err == nil, "trie::MustEncodeUnpackedBytes: err: %v, unpacked: %s", err, hex.EncodeToString(unpacked))
	return ret
}

func decodeToUnpackedBytes(encoded []byte) ([]byte, error) {
	if len(encoded) == 0 {
		return nil, nil
	}
	return decode16(encoded)
}
