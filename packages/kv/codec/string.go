package codec

var String = NewCodec(decodeString, encodeString)

func decodeString(b []byte) (string, error) {
	return string(b), nil
}

func encodeString(value string) []byte {
	return []byte(value)
}

var Bytes = NewCodec(decodeBytes, encodeBytes)

func decodeBytes(b []byte) ([]byte, error) {
	return b, nil
}

func encodeBytes(value []byte) []byte {
	return value
}
