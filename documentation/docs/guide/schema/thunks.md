# Thunk Functions

In computer programming, a thunk is a function used to inject a calculation into another
function. Thunks are used to insert operations at the beginning or end of the other
function to adapt it to changing requirements. If you remember from
the [function call context](context.md) section, the `on_load` function and skeleton
function signatures looked like this:

```go
func OnLoad() {
    exports := wasmlib.NewScExports()
    exports.AddFunc("divide", funcDivide)
    exports.AddFunc("init", funcInit)
    exports.AddFunc("member", funcMember)
    exports.AddFunc("setOwner", funcSetOwner)
    exports.AddView("getFactor", viewGetFactor)
    exports.AddView("getOwner", viewGetOwner)
}

func funcDivide(ctx wasmlib.ScFuncContext) {}
func funcInit(ctx wasmlib.ScFuncContext) {}
func funcMember(ctx wasmlib.ScFuncContext) {}
func funcSetOwner(ctx wasmlib.ScFuncContext) {}
func viewGetFactor(ctx wasmlib.ScViewContext) {}
func viewGetOwner(ctx wasmlib.ScViewContext) {}
```

```rust
fn on_load() {
    let exports = ScExports::new();
    exports.add_func("divide", func_divide);
    exports.add_func("init", func_init);
    exports.add_func("member", func_member);
    exports.add_func("setOwner", func_set_owner);
    exports.add_view("getFactor", view_get_factor);
    exports.add_view("getOwner", view_get_owner);
}

fn func_divide(ctx: &ScFuncContext) {}
fn func_init(ctx: &ScFuncContext) {}
fn func_member(ctx: &ScFuncContext) {}
fn func_set_owner(ctx: &ScFuncContext) {}
fn view_get_factor(ctx: &ScViewContext) {}
fn view_get_owner(ctx: &ScViewContext) {}
```

Now that the schema tool introduces a bunch of automatically generated features, that is
no longer sufficient. Luckily, the schema tool also generates thunks to inject these
features, before calling the function implementations that are maintained by the user.
Here is the new `on_load` function for the `dividend` contract:

```go
func OnLoad() {
    exports := wasmlib.NewScExports()
    exports.AddFunc(FuncDivide, funcDivideThunk)
    exports.AddFunc(FuncInit, funcInitThunk)
    exports.AddFunc(FuncMember, funcMemberThunk)
    exports.AddFunc(FuncSetOwner, funcSetOwnerThunk)
    exports.AddView(ViewGetFactor, viewGetFactorThunk)
    exports.AddView(ViewGetOwner, viewGetOwnerThunk)
    
    for i, key := range keyMap {
        idxMap[i] = key.KeyID()
    }
}
```

```rust
fn on_load() {
    let exports = ScExports::new();
    exports.add_func(FUNC_DIVIDE, func_divide_thunk);
    exports.add_func(FUNC_INIT, func_init_thunk);
    exports.add_func(FUNC_MEMBER, func_member_thunk);
    exports.add_func(FUNC_SET_OWNER, func_set_owner_thunk);
    exports.add_view(VIEW_GET_FACTOR, view_get_factor_thunk);
    exports.add_view(VIEW_GET_OWNER, view_get_owner_thunk);

    unsafe {
        for i in 0..KEY_MAP_LEN {
            IDX_MAP[i] = get_key_id_from_string(KEY_MAP[i]);
        }
    }
}
```

As you can see instead of calling the user functions directly, we now call thunk versions
of these functions. We also added initialization of a local array that holds all key IDs
negotiated with the host, so that we can simply use (generated) indexes into this array
instead of having to negotiate these IDs each time we need them. The rest of the generated
code will use those indexes whenever a known key is used.

Here is an example of a thunk function for the `setOwner` contract function. You can
examine the other thunks that all follow the same pattern in the generated `lib.rs`:

```go
type SetOwnerContext struct {
    Params ImmutableSetOwnerParams
    State  MutableDividendState
}

func funcSetOwnerThunk(ctx wasmlib.ScFuncContext) {
    ctx.Log("dividend.funcSetOwner")
    // only defined owner of contract can change owner
    access := ctx.State().GetAgentID(wasmlib.Key("owner"))
    ctx.Require(access.Exists(), "access not set: owner")
    ctx.Require(ctx.Caller() == access.Value(), "no permission")
    
    f := &SetOwnerContext{
        Params: ImmutableSetOwnerParams{
           id: wasmlib.OBJ_ID_PARAMS,
        },
        State: MutableDividendState{
            id: wasmlib.OBJ_ID_STATE,
        },
    }
    ctx.Require(f.Params.Owner().Exists(), "missing mandatory owner")
    funcSetOwner(ctx, f)
    ctx.Log("dividend.funcSetOwner ok")
}
```

```rust
pub struct SetOwnerContext {
    params: ImmutableSetOwnerParams,
    state:  MutableDividendState,
}

fn func_set_owner_thunk(ctx: &ScFuncContext) {
    ctx.log("dividend.funcSetOwner");
    // only defined owner of contract can change owner
    let access = ctx.state().get_agent_id("owner");
    ctx.require(access.exists(), "access not set: owner");
    ctx.require(ctx.caller() == access.value(), "no permission");

    let f = SetOwnerContext {
        params: ImmutableSetOwnerParams {
            id: OBJ_ID_PARAMS,
        },
        state: MutableDividendState {
            id: OBJ_ID_STATE,
        },
    };
    ctx.require(f.params.owner().exists(), "missing mandatory owner");
    func_set_owner(ctx, &f);
    ctx.log("dividend.funcSetOwner ok");
}
```

The thunk first logs the contract and function name to show the call has started. Then it
sets up the access control for the function according to schema.json. In this case it
retrieves the `owner` state variable, requires that it exists, and then requires that the
caller() of the function equals that value. Any failing requirement will panic out of the
function with an error message. So this code makes sure only the owner of the contract can
call this function.

Next we set up a strongly typed function-specific context structure. First we add the
function-specific immutable `params` interface structure, which is only present when the
function can have parameters. Then we add the contract-specific `state` interface
structure. In this case it is mutable because setOwner is a Func. For Views this will be
an immutable state interface. Finally, we add the function-specific mutable `results`
interface structure, which is only present when the function returns results. Obviously,
this is not the case for this setOwner function.

Now we get to the point where we can use the function-specific `params` interface to check
for mandatory parameters. Each mandatory parameter is required to exist, or else we will
panic out of the function with an error message.

With the automated checks and setup completed we now call the function implementation that
is maintained by the user. After the user function has completed we log that the contract
function has completed successfully. Remember that any error within the user function will
cause a panic, so this logging will never happen in that case.

In the next section we will look at the specifics of view functions.
