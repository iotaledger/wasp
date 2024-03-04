package migrations

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrationScheme_WithTargetSchemaVersion(t *testing.T) {
	scheme := &MigrationScheme{
		BaseSchemaVersion: 3,
		Migrations: []Migration{
			{}, // 4
			{}, // 5
		},
	}

	t.Run("ok", func(t *testing.T) {
		newScheme, err := scheme.WithTargetSchemaVersion(4)
		require.NoError(t, err)
		require.EqualValues(t, 4, newScheme.LatestSchemaVersion())
	})

	t.Run("missing migration code", func(t *testing.T) {
		_, err := scheme.WithTargetSchemaVersion(2)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrMissingMigrationCode))
	})

	t.Run("invalid schema version", func(t *testing.T) {
		_, err := scheme.WithTargetSchemaVersion(6)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrInvalidSchemaVersion))
	})
}
