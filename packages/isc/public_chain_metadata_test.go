package isc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicMetadataSerialization(t *testing.T) {
	metadata := PublicChainMetadata{
		EVMJsonRPCURL:   "EVMJsonRPCURL",
		EVMWebSocketURL: "EVMWebSocketURL",
		Name:            "Name",
		Description:     "Description",
		Website:         "Website",
	}

	metadataDeserialized, err := PublicChainMetadataFromBytes(metadata.Bytes())

	require.NoError(t, err)
	require.EqualValues(t, metadata, *metadataDeserialized)
	require.EqualValues(t, metadata.Bytes(), metadataDeserialized.Bytes())
}
