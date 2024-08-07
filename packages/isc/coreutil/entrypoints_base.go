package coreutil

import "github.com/iotaledger/wasp/packages/isc"

//go:generate go run generator/generate_entrypoints.go

// EP0 is a utility type for entry points that receive 0 parameters
type EP0[S isc.SandboxBase] struct{ EntryPointInfo[S] }

func (e EP0[S]) Message() isc.Message { return e.EntryPointInfo.Message(isc.CallArguments{}) }

func NewEP0(contract *ContractInfo, name string) EP0[isc.Sandbox] {
	return EP0[isc.Sandbox]{EntryPointInfo: contract.Func(name)}
}

func NewViewEP0(contract *ContractInfo, name string) EP0[isc.SandboxView] {
	return EP0[isc.SandboxView]{EntryPointInfo: contract.ViewFunc(name)}
}
