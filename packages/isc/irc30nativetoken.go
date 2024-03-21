package isc

import "encoding/json"

// IRC30NativeTokenMetadata represents the Native Token metadata according to IRC30.
// See: https://github.com/iotaledger/tips/blob/main/tips/TIP-0030/tip-0030.md
// Right now, only required properties are included.
// Optional parameters such as description or logo/Url can be added later
type IRC30NativeTokenMetadata struct {
	Standard string `json:"standard"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

func NewIRC30NativeTokenMetadata(name, symbol string, decimals uint8) *IRC30NativeTokenMetadata {
	return &IRC30NativeTokenMetadata{
		Standard: "IRC30",
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}
}

func (m *IRC30NativeTokenMetadata) Bytes() []byte {
	ret, err := json.Marshal(m)
	if err != nil {
		// only happens when passing unsupported structures, which we don't
		panic(err)
	}
	return ret
}

func IRC30NativeTokenMetadataFromBytes(b []byte) (*IRC30NativeTokenMetadata, error) {
	var m IRC30NativeTokenMetadata
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
