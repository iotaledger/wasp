package isc_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestPublicChainMetadataSerialization(t *testing.T) {
	metadata := &isc.PublicChainMetadata{
		EVMJsonRPCURL:   "EVMJsonRPCURL",
		EVMWebSocketURL: "EVMWebSocketURL",
		Name:            "Name",
		Description:     "Description",
		Website:         "Website",
	}
	rwutil.ReadWriteTest(t, metadata, new(isc.PublicChainMetadata))
	rwutil.BytesTest(t, metadata, isc.PublicChainMetadataFromBytes)
}
