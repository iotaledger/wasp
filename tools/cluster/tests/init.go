//go:build !tests_build_only
// +build !tests_build_only

package tests

import (
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func init() {
	l1 = l1starter.ClusterStart(parseConfig())
}
