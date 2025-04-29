package util

import (
	"encoding"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/packages/cryptolib"
)

func EncodeHexBinaryMarshaled(value encoding.BinaryMarshaler) (string, error) {
	data, err := value.MarshalBinary()
	if err != nil {
		return "", err
	}

	return hexutil.Encode(data), nil
}

func DecodeHexBinaryMarshaled(dataHex string, value encoding.BinaryUnmarshaler) error {
	data, err := cryptolib.DecodeHex(dataHex)
	if err != nil {
		return err
	}

	return value.UnmarshalBinary(data)
}

func EncodeSliceHexBinaryMarshaled[M encoding.BinaryMarshaler](values []M) ([]string, error) {
	results := make([]string, 0)
	for _, value := range values {
		valueHex, err := EncodeHexBinaryMarshaled(value)
		if err != nil {
			return nil, err
		}
		results = append(results, valueHex)
	}
	return results, nil
}

func DecodeSliceHexBinaryMarshaled[M encoding.BinaryUnmarshaler](dataHex []string, values []M) error {
	for i, hex := range dataHex {
		data, err := cryptolib.DecodeHex(hex)
		if err != nil {
			return err
		}

		if err := values[i].UnmarshalBinary(data); err != nil {
			return err
		}
	}

	return nil
}

// PrefixHex formats byte data with a hex prefix, mostly for logging.
func PrefixHex(data []byte, prefixLen int) string {
	if data == nil {
		return "<nil>"
	}
	if len(data) <= prefixLen {
		return hexutil.Encode(data)
	}
	return hexutil.Encode(data[0:prefixLen]) + "..."
}
