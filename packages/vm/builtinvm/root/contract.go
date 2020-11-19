package root

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type ContractInterface struct {
	VMType      string
	Name        string
	Description string
	Version     string
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

func (i *ContractInterface) GetFunction(name string) (*ContractFunctionInterface, bool) {
	f, ok := i.Functions[coretypes.Hn(name)]
	return &f, ok
}

func (i *ContractInterface) GetEntryPoint(code coretypes.Hname) (vmtypes.EntryPoint, bool) {
	f, ok := i.Functions[code]
	return &f, ok
}

func (i *ContractInterface) GetDescription() string {
	return i.Description
}

func (f *ContractFunctionInterface) Hname() coretypes.Hname {
	return coretypes.Hn(f.Name)
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
