package isc

import (
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type PublicChainMetadata struct {
	EVMJsonRPCURL   string
	EVMWebSocketURL string
	Name            string
	Description     string
	Website         string
}

func PublicChainMetadataFromBytes(data []byte) (*PublicChainMetadata, error) {
	return bcs.Unmarshal[*PublicChainMetadata](data)
}

func (m *PublicChainMetadata) Bytes() []byte {
	return bcs.MustMarshal(m)
}
