package solo

import "github.com/stretchr/testify/require"

type Context interface {
	require.TestingT
	Name() string
	Cleanup(func())
	Helper()
	Logf(string, ...any)
	Fatalf(string, ...any)
}
