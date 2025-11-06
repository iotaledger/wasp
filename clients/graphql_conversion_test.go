package clients

import (
	"testing"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/stretchr/testify/require"
)

func TestConvertGraphQLTryGetPastObjectResponse_VersionFound(t *testing.T) {
	// Create a mock GraphQL response with a found object
	testAddr := *iotago.MustAddressFromHex("0x0000000000000000000000000000000000000000000000000000000000000001")
	testDigest := "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="

	resp := &iotagraphql.TryGetPastObjectResponse{
		Current: iotagraphql.TryGetPastObjectCurrentObject{
			Address: testAddr,
			Version: 100,
		},
		Object: iotagraphql.TryGetPastObjectObject{
			RPC_OBJECT_FIELDS: iotagraphql.RPC_OBJECT_FIELDS{
				ObjectId: testAddr,
				Version:  50,
				Digest:   testDigest,
			},
		},
	}

	result, err := convertGraphQLTryGetPastObjectResponse(resp, 50, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Data.VersionFound)
	require.Nil(t, result.Data.ObjectNotExists)
	require.Nil(t, result.Data.VersionNotFound)
	require.Nil(t, result.Data.VersionTooHigh)
}

func TestConvertGraphQLTryGetPastObjectResponse_VersionTooHigh(t *testing.T) {
	// Create a mock GraphQL response where requested version is higher than current
	testAddr := *iotago.MustAddressFromHex("0x0000000000000000000000000000000000000000000000000000000000000001")

	resp := &iotagraphql.TryGetPastObjectResponse{
		Current: iotagraphql.TryGetPastObjectCurrentObject{
			Address: testAddr,
			Version: 100,
		},
		Object: iotagraphql.TryGetPastObjectObject{
			RPC_OBJECT_FIELDS: iotagraphql.RPC_OBJECT_FIELDS{
				ObjectId: iotago.Address{}, // Empty address indicates not found
			},
		},
	}

	result, err := convertGraphQLTryGetPastObjectResponse(resp, 150, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Nil(t, result.Data.VersionFound)
	require.Nil(t, result.Data.ObjectNotExists)
	require.Nil(t, result.Data.VersionNotFound)
	require.NotNil(t, result.Data.VersionTooHigh)
	require.Equal(t, iotago.SequenceNumber(150), result.Data.VersionTooHigh.AskedVersion)
	require.Equal(t, iotago.SequenceNumber(100), result.Data.VersionTooHigh.LatestVersion)
}

func TestConvertGraphQLTryGetPastObjectResponse_VersionNotFound(t *testing.T) {
	// Create a mock GraphQL response where version is not found (but not too high)
	testAddr := *iotago.MustAddressFromHex("0x0000000000000000000000000000000000000000000000000000000000000001")

	resp := &iotagraphql.TryGetPastObjectResponse{
		Current: iotagraphql.TryGetPastObjectCurrentObject{
			Address: testAddr,
			Version: 100,
		},
		Object: iotagraphql.TryGetPastObjectObject{
			RPC_OBJECT_FIELDS: iotagraphql.RPC_OBJECT_FIELDS{
				ObjectId: iotago.Address{}, // Empty address indicates not found
			},
		},
	}

	result, err := convertGraphQLTryGetPastObjectResponse(resp, 50, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Nil(t, result.Data.VersionFound)
	require.Nil(t, result.Data.ObjectNotExists)
	require.NotNil(t, result.Data.VersionNotFound)
	require.Nil(t, result.Data.VersionTooHigh)
	require.Equal(t, iotago.SequenceNumber(50), result.Data.VersionNotFound.SequenceNumber)
}

func TestConvertGraphQLTryGetPastObjectResponse_ObjectNotExists(t *testing.T) {
	// Create a mock GraphQL response where the object doesn't exist at all
	emptyAddr := iotago.Address{}

	resp := &iotagraphql.TryGetPastObjectResponse{
		Current: iotagraphql.TryGetPastObjectCurrentObject{
			Address: emptyAddr,
			Version: 0,
		},
		Object: iotagraphql.TryGetPastObjectObject{
			RPC_OBJECT_FIELDS: iotagraphql.RPC_OBJECT_FIELDS{
				ObjectId: emptyAddr,
			},
		},
	}

	result, err := convertGraphQLTryGetPastObjectResponse(resp, 50, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Nil(t, result.Data.VersionFound)
	require.NotNil(t, result.Data.ObjectNotExists)
	require.Nil(t, result.Data.VersionNotFound)
	require.Nil(t, result.Data.VersionTooHigh)
}

func TestConvertRPCObjectFieldsToIotaObjectData(t *testing.T) {
	// Test basic conversion with minimal fields
	testAddr := *iotago.MustAddressFromHex("0x0000000000000000000000000000000000000000000000000000000000000001")
	testDigest := "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="

	fields := &iotagraphql.RPC_OBJECT_FIELDS{
		ObjectId: testAddr,
		Version:  42,
		Digest:   testDigest,
	}

	result, err := convertRPCObjectFieldsToIotaObjectData(fields, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.ObjectID)
	require.Equal(t, iotago.ObjectID(testAddr), *result.ObjectID)
	require.Equal(t, uint64(42), result.Version.Uint64())
}

func TestConvertRPCObjectFieldsToIotaObjectData_WithOptions(t *testing.T) {
	// Test conversion with ShowType option
	testAddr := *iotago.MustAddressFromHex("0x0000000000000000000000000000000000000000000000000000000000000001")
	testDigest := "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="

	fields := &iotagraphql.RPC_OBJECT_FIELDS{
		ObjectId: testAddr,
		Version:  42,
		Digest:   testDigest,
		AsMoveObjectType: iotagraphql.RPC_OBJECT_FIELDSAsMoveObjectTypeMoveObject{
			Contents: iotagraphql.RPC_OBJECT_FIELDSAsMoveObjectTypeMoveObjectContentsMoveValue{
				Type: iotagraphql.RPC_OBJECT_FIELDSAsMoveObjectTypeMoveObjectContentsMoveValueTypeMoveType{
					Repr: "0x2::coin::Coin<0x2::iota::IOTA>",
				},
			},
		},
	}

	options := &iotajsonrpc.IotaObjectDataOptions{
		ShowType: true,
	}

	result, err := convertRPCObjectFieldsToIotaObjectData(fields, options)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Type)
	require.Equal(t, "0x2::coin::Coin<0x2::iota::IOTA>", *result.Type)
}
