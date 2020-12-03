package codec

func DecodeString(b []byte) (string, bool, error) {
	if b == nil {
		return "", false, nil
	}
	return string(b), true, nil
}

func EncodeString(value string) []byte {
	return []byte(value)
}
