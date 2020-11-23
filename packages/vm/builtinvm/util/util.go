package util

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

func BuiltinFullName(name string, version string) string {
	return name + "-" + version
}

func BuiltinProgramHash(name string, version string) hashing.HashValue {
	return *hashing.HashStrings(BuiltinFullName(name, version))
}

func BuiltinHname(name string, version string) coretypes.Hname {
	return coretypes.Hn(BuiltinFullName(name, version))
}
