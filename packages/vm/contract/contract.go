package contract

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"

	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type ContractInterface struct {
	Name        string
	hname       coret.Hname
	Description string
	ProgramHash hashing.HashValue
	Functions   map[coret.Hname]ContractFunctionInterface
}

type ContractFunctionInterface struct {
	Name        string
	Handler     Handler
	ViewHandler ViewHandler
}

func Funcs(init Handler, fns []ContractFunctionInterface) map[coret.Hname]ContractFunctionInterface {
	ret := map[coret.Hname]ContractFunctionInterface{
		coret.EntryPointInit: Func("init", init),
	}
	for _, f := range fns {
		if _, ok := ret[f.Hname()]; ok {
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

		ret[f.Hname()] = f
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

type Handler func(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error)
type ViewHandler func(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error)

func (i *ContractInterface) WithFunctions(init Handler, funcs []ContractFunctionInterface) {
	i.Functions = Funcs(init, funcs)
}

func (i *ContractInterface) GetFunction(name string) (*ContractFunctionInterface, bool) {
	f, ok := i.Functions[coret.Hn(name)]
	return &f, ok
}

func (i *ContractInterface) GetEntryPoint(code coret.Hname) (vmtypes.EntryPoint, bool) {
	f, ok := i.Functions[code]
	return &f, ok
}

func (i *ContractInterface) GetDescription() string {
	return i.Description
}

func (i *ContractInterface) Hname() coret.Hname {
	if i.hname == 0 {
		i.hname = coret.Hn(i.Name)
	}
	return i.hname
}

func (f *ContractFunctionInterface) Hname() coret.Hname {
	return coret.Hn(f.Name)
}

func (f *ContractFunctionInterface) WithGasLimit(_ int) vmtypes.EntryPoint {
	return f
}

func (f *ContractFunctionInterface) Call(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	if f.IsView() {
		return nil, vmtypes.ErrWrongTypeEntryPoint
	}
	ret, err := f.Handler(ctx)
	if err != nil {
		ctx.Eventf("error occurred: '%v'", err)
	}
	return ret, err
}

func (f *ContractFunctionInterface) CallView(ctx vmtypes.SandboxView) (codec.ImmutableCodec, error) {
	if !f.IsView() {
		return nil, vmtypes.ErrWrongTypeEntryPoint
	}
	ret, err := f.ViewHandler(ctx)
	if err != nil {
		ctx.Eventf("error occurred: '%v'", err)
	}
	return ret, err
}

func (f *ContractFunctionInterface) IsView() bool {
	return f.ViewHandler != nil
}
