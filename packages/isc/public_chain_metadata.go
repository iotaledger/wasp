package isc

import (
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

type PublicChainMetadata struct {
	EVMJsonRPCURL   string
	EVMWebSocketURL string

	Name        string
	Description string
	Website     string
}

func readMetadataString(mu *marshalutil.MarshalUtil) (string, error) {
	sz, err := mu.ReadUint16()
	if err != nil {
		return "", err
	}
	ret, err := mu.ReadBytes(int(sz))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func writeMetadataString(mu *marshalutil.MarshalUtil, str string) {
	mu.
		WriteUint16(uint16(len(str))).
		WriteBytes([]byte(str))
}

func PublicChainMetadataFromMarshalUtil(mu *marshalutil.MarshalUtil) (*PublicChainMetadata, error) {
	ret := &PublicChainMetadata{}
	var err error

	if ret.EVMJsonRPCURL, err = readMetadataString(mu); err != nil {
		return nil, err
	}

	if ret.EVMWebSocketURL, err = readMetadataString(mu); err != nil {
		return nil, err
	}

	if ret.Name, err = readMetadataString(mu); err != nil {
		return nil, err
	}

	if ret.Description, err = readMetadataString(mu); err != nil {
		return nil, err
	}

	if ret.Website, err = readMetadataString(mu); err != nil {
		return nil, err
	}

	return ret, nil
}

func PublicChainMetadataFromBytes(metadataBytes []byte) (*PublicChainMetadata, error) {
	mu := marshalutil.New(metadataBytes)
	return PublicChainMetadataFromMarshalUtil(mu)
}

func (m *PublicChainMetadata) Bytes() []byte {
	mu := marshalutil.New()

	writeMetadataString(mu, m.EVMJsonRPCURL)
	writeMetadataString(mu, m.EVMWebSocketURL)
	writeMetadataString(mu, m.Name)
	writeMetadataString(mu, m.Description)
	writeMetadataString(mu, m.Website)

	return mu.Bytes()
}
