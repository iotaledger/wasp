package isc

import "encoding/json"

// IRC27NFTMetadata represents an NFT metadata according to IRC27.
// See: https://github.com/iotaledger/tips/blob/main/tips/TIP-0027/tip-0027.md
type IRC27NFTMetadata struct {
	Standard       string             `json:"standard"`
	Version        string             `json:"version"`
	MIMEType       string             `json:"type"`
	URI            string             `json:"uri"`
	Name           string             `json:"name"`
	CollectionName string             `json:"collectionName,omitempty"`
	Royalties      map[string]float32 `json:"royalties,omitempty"`
	IssuerName     string             `json:"issuerName,omitempty"`
	Description    string             `json:"description,omitempty"`
	Attributes     []interface{}      `json:"attributes,omitempty"`
}

func NewIRC27NFTMetadata(mimeType, uri, name string, attributes []interface{}) *IRC27NFTMetadata {
	return &IRC27NFTMetadata{
		Standard:   "IRC27",
		Version:    "v1.0",
		MIMEType:   mimeType,
		URI:        uri,
		Name:       name,
		Attributes: attributes,
	}
}

func (m *IRC27NFTMetadata) Bytes() []byte {
	ret, err := json.Marshal(m)
	if err != nil {
		// only happens when passing unsupported structures, which we don't
		panic(err)
	}
	return ret
}

func IRC27NFTMetadataFromBytes(b []byte) (*IRC27NFTMetadata, error) {
	var m IRC27NFTMetadata
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
