package isc_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestPublicChainMetadataSerialization(t *testing.T) {
	metadata := &isc.PublicChainMetadata{
		EVMJsonRPCURL:   "EVMJsonRPCURL",
		EVMWebSocketURL: "EVMWebSocketURL",
		Name:            "Name",
		Description:     "Description",
		Website:         "Website",
	}
	bcs.TestCodec(t, metadata)
}
