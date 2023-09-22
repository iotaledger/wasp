package blob

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/kv/dict"
)

func TestMustGetBlobHash(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		fields := dict.Dict{
			"key0": []byte("val0"),
			"key1": []byte("val1"),
		}

		h, keys, values := mustGetBlobHash(fields)
		for i, k := range keys {
			require.Equal(t, fields[k], values[i])
		}

		resHash, err := hex.DecodeString("54cb8e9c45ca6d368dba92da34cfa47ce617f04807af19f67de333fad0039e6b")
		require.NoError(t, err)
		require.Equal(t, resHash, h.Bytes())
	})
	t.Run("potential collision", func(t *testing.T) {
		fields := dict.Dict{
			"123":  []byte("ab"),
			"123a": []byte("b"),
		}

		h, keys, values := mustGetBlobHash(fields)
		for i, k := range keys {
			require.Equal(t, fields[k], values[i])
		}

		resHash, err := hex.DecodeString("67551f56072748d3b3453808f87bc27098eeee814684896b118a936a6226e023")
		require.NoError(t, err)
		require.Equal(t, resHash, h.Bytes())
	})
}
