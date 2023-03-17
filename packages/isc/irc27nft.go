package isc

import "encoding/json"

// IRC27NFTMetadata represents an NFT metadata according to IRC27.
// See: https://github.com/iotaledger/tips/blob/main/tips/TIP-0027/tip-0027.md
type IRC27NFTMetadata struct {
	Standard string `json:"standard"`
	Version  string `json:"version"`
	MIMEType string `json:"type"`
	URI      string `json:"uri"`
	Name     string `json:"name"`
}

func NewIRC27NFTMetadata(mimeType, uri, name string) *IRC27NFTMetadata {
	return &IRC27NFTMetadata{
		Standard: "IRC27",
		Version:  "v1.0",
		MIMEType: mimeType,
		URI:      uri,
		Name:     name,
	}
}

func (m *IRC27NFTMetadata) Bytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *IRC27NFTMetadata) MustBytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func IRC27NFTMetadataFromBytes(b []byte) (*IRC27NFTMetadata, error) {
	var m IRC27NFTMetadata
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
