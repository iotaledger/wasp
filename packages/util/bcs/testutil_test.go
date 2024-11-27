package bcs_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestEmpty(t *testing.T) {
	require.Equal(t, 0, bcs.Empty(123))
	require.Equal(t, 0.0, bcs.Empty(123.4))
	require.Equal(t, "", bcs.Empty("hello"))
	require.Equal(t, structWithField[bool]{}, bcs.Empty(structWithField[bool]{true}))
	require.Equal(t, (*int)(nil), bcs.Empty(lo.ToPtr(42)))

	v := bcs.Empty(any(123))
	require.Equal(t, 0, v.(int))

	v = bcs.Empty(any(lo.ToPtr(123)))
	require.Equal(t, (*int)(nil), v.(*int))
}
