## Function Descriptors

The schema tool provides us with an easy way to access to smart contract functions through
so-called _function descriptors_. These are structures that provide access to the optional
params and results maps through strict compile-time checked interfaces. They will also
allow you to initiate the function by calling it synchronously or posting a request to run
it asynchronously.

The schema tool will generate a specific function descriptor for each func and view. It
will also generate an interface called ScFuncs that can be used to create and initialize
each function descriptor. Here is the code generated for the `dividend`
example:

```golang
package dividend

import "github.com/iotaledger/wasp/packages/vm/wasmlib"

type DivideCall struct {
	Func *wasmlib.ScFunc
}

type InitCall struct {
	Func   *wasmlib.ScInitFunc
	Params MutableInitParams
}

type MemberCall struct {
	Func   *wasmlib.ScFunc
	Params MutableMemberParams
}

type SetOwnerCall struct {
	Func   *wasmlib.ScFunc
	Params MutableSetOwnerParams
}

type GetFactorCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetFactorParams
	Results ImmutableGetFactorResults
}

type Funcs struct{}

var ScFuncs Funcs

func (sc Funcs) Divide(ctx wasmlib.ScFuncCallContext) *DivideCall {
	return &DivideCall{Func: wasmlib.NewScFunc(HScName, HFuncDivide)}
}

func (sc Funcs) Init(ctx wasmlib.ScFuncCallContext) *InitCall {
	f := &InitCall{Func: wasmlib.NewScInitFunc(HScName, HFuncInit, ctx, keyMap[:], idxMap[:])}
	f.Func.SetPtrs(&f.Params.id, nil)
	return f
}

func (sc Funcs) Member(ctx wasmlib.ScFuncCallContext) *MemberCall {
	f := &MemberCall{Func: wasmlib.NewScFunc(HScName, HFuncMember)}
	f.Func.SetPtrs(&f.Params.id, nil)
	return f
}

func (sc Funcs) SetOwner(ctx wasmlib.ScFuncCallContext) *SetOwnerCall {
	f := &SetOwnerCall{Func: wasmlib.NewScFunc(HScName, HFuncSetOwner)}
	f.Func.SetPtrs(&f.Params.id, nil)
	return f
}

func (sc Funcs) GetFactor(ctx wasmlib.ScViewCallContext) *GetFactorCall {
	f := &GetFactorCall{Func: wasmlib.NewScView(HScName, HViewGetFactor)}
	f.Func.SetPtrs(&f.Params.id, &f.Results.id)
	return f
}
```

As you can see a struct has been generated for each of the funcs and views. Note that the
structs only provide access to `Params` or `Results` when these are specified for the
function. Also note that each struct has a `Func` member that can be used to initiate the
function call in certain ways. The `Func` member will be of type ScFunc or ScView,
depending on whether the function is a func or a view.

The ScFuncs struct provides a member function for each func or view that will create their
respective function descriptor, initialize it properly, and returns it.

In the next section we will look at how to use function descriptors to call a smart
contract function synchronously.

Next: [Calling Functions](call.md)

