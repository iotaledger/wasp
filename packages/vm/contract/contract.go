package contract

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type ContractInterface struct {
	Name        string
	hname       coretypes.Hname
	Description string
	ProgramHash hashing.HashValue
	Functions   map[coretypes.Hname]ContractFunctionInterface
}

type ContractFunctionInterface struct {
	Name        string
	Handler     Handler
	ViewHandler ViewHandler
}

func Funcs(init Handler, fns []ContractFunctionInterface) map[coretypes.Hname]ContractFunctionInterface {
	ret := map[coretypes.Hname]ContractFunctionInterface{
		coretypes.EntryPointInit: Func("init", init),
	}
	for _, f := range fns {
		hname := f.Hname()
		if _, ok := ret[hname]; ok {
			panic(fmt.Sprintf("Duplicate function: %s", f.Name))
		}

		handlers := 0
		if f.Handler != nil {
			handlers += 1
		}
		if f.ViewHandler != nil {
			handlers += 1
		}
		if handlers != 1 {
			panic("Exactly one of Handler, ViewHandler must be set")
		}

		ret[hname] = f
	}
	return ret
}

func Func(name string, handler Handler) ContractFunctionInterface {
	return ContractFunctionInterface{
		Name:    name,
		Handler: handler,
	}
}

func ViewFunc(name string, handler ViewHandler) ContractFunctionInterface {
	return ContractFunctionInterface{
		Name:        name,
		ViewHandler: handler,
	}
}

type Handler func(ctx coretypes.Sandbox) (dict.Dict, error)
type ViewHandler func(ctx coretypes.SandboxView) (dict.Dict, error)

func (i *ContractInterface) WithFunctions(init Handler, funcs []ContractFunctionInterface) {
	i.Functions = Funcs(init, funcs)
}

func (i *ContractInterface) GetFunction(name string) (*ContractFunctionInterface, bool) {
	f, ok := i.Functions[coretypes.Hn(name)]
	return &f, ok
}

func (i *ContractInterface) GetEntryPoint(code coretypes.Hname) (coretypes.EntryPoint, bool) {
	f, ok := i.Functions[code]
	return &f, ok
}

func (i *ContractInterface) GetDescription() string {
	return i.Description
}

// Hname caches the value
func (i *ContractInterface) Hname() coretypes.Hname {
	if i.hname == 0 {
		i.hname = coretypes.Hn(i.Name)
	}
	return i.hname
}

func (i *ContractInterface) ContractID(chainID coretypes.ChainID) coretypes.ContractID {
	return coretypes.NewContractID(chainID, i.Hname())
}

func (f *ContractFunctionInterface) Hname() coretypes.Hname {
	return coretypes.Hn(f.Name)
}

func (f *ContractFunctionInterface) WithGasLimit(_ int) coretypes.EntryPoint {
	return f
}

func (f *ContractFunctionInterface) Call(ctx coretypes.Sandbox) (dict.Dict, error) {
	if f.IsView() {
		return nil, coretypes.ErrWrongTypeEntryPoint
	}
	ret, err := f.Handler(ctx)
	if err != nil {
		ctx.Log().Debugf("error occurred: '%v'", err)
	}
	return ret, err
}

func (f *ContractFunctionInterface) CallView(ctx coretypes.SandboxView) (dict.Dict, error) {
	if !f.IsView() {
		return nil, coretypes.ErrWrongTypeEntryPoint
	}
	ret, err := f.ViewHandler(ctx)
	if err != nil {
		ctx.Log().Debugf("error occurred: '%v'", err)
	}
	return ret, err
}

func (f *ContractFunctionInterface) IsView() bool {
	return f.ViewHandler != nil
}
