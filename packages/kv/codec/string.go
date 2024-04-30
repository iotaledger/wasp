package codec

var String = NewCodec(decodeString, encodeString)

func decodeString(b []byte) (string, error) {
	return string(b), nil
}

func encodeString(value string) []byte {
	return []byte(value)
}
