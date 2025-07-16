package isc_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestPublicChainMetadataSerialization(t *testing.T) {
	metadata := &isc.PublicChainMetadata{
		EVMJsonRPCURL:   "EVMJsonRPCURL",
		EVMWebSocketURL: "EVMWebSocketURL",
		Name:            "Name",
		Description:     "Description",
		Website:         "Website",
	}
	bcs.TestCodecAndHash(t, metadata, "72d016e59cb1")
}
