package legacy

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
)

func TestLegacyAgentIDForFakeTXs(t *testing.T) {
	testID := isc.NewAddressAgentID(&cryptolib.Address{0xff, 0xff, 0xff})
	require.Equal(t, testID.Bytes()[:4], []byte{0x1, 0xff, 0xff, 0xff})

	legacyID := AgentIDToBytes(allmigrations.SchemaVersionMigratedRebased, testID)
	require.Equal(t, legacyID[:5], []byte{0x1, 0x0, 0xff, 0xff, 0xff})
}

func TestLegacyAgentIDForFakeTXsWithNewSchemaVersion(t *testing.T) {
	testID := isc.NewAddressAgentID(&cryptolib.Address{0xff, 0xff, 0xff})
	require.Equal(t, testID.Bytes()[:4], []byte{0x1, 0xff, 0xff, 0xff})
	require.Equal(t, AgentIDToBytes(allmigrations.LatestSchemaVersion, testID), testID.Bytes())
}

func TestLegacyAssetsBaseToken(t *testing.T) {
	// Stardust isc.Assets encoded bytes for NewAssets(123123) (123123 base token is)
	stardustEncodedAssetsForBaseToken := []byte{0x80, 0xf3, 0xc1, 0x7}
	a := isc.NewAssets(123123 * 1000) // * 1000 for 6=>9 decimals conversion
	require.Equal(t, AssetsToBytes(allmigrations.SchemaVersionMigratedRebased, a), stardustEncodedAssetsForBaseToken)
}

func TestLegacyAssetsBaseTokenWithNFT(t *testing.T) {
	// Stardust isc.Assets encoded bytes for NewAssets(123123) (123123 base token is)
	stardustEncodedAssetsForBaseToken := []byte{0xA0, 0xF3, 0xC1, 0x7, 0x1, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	a := isc.NewAssets(123123 * 1000) // * 1000 for 6=>9 decimals conversion
	a.AddObject(isc.NewIotaObject(iotago.ObjectID{0xff, 0xff}, iotago.MustTypeFromString("0x1::a::A")))
	require.Equal(t, AssetsToBytes(allmigrations.SchemaVersionMigratedRebased, a), stardustEncodedAssetsForBaseToken)
}

func TestLegacyAssetsBaseTokenWithNewSchemaVersion(t *testing.T) {
	a := isc.NewAssets(123123 * 1000) // * 1000 for 6=>9 decimals conversion
	a.AddObject(isc.NewIotaObject(iotago.ObjectID{0xff, 0xff}, iotago.MustTypeFromString("0x1::a::A")))
	require.Equal(t, AssetsToBytes(allmigrations.LatestSchemaVersion, a), a.Bytes())
}
