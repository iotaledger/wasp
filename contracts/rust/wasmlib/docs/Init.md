## Smart Contract Initialization

Smart contracts start out with a completely blank state. Sometimes you want to
be able to define initial state, for example if your contract is configurable.
You want to be able to pass this configuration to the contract upon deployment,
so that its state reflects that configuration once the first request comes in.
To support this initialization we allow the smart contract creator to provide an
optional function named `init`. This function, when provided, will automatically
be called immediately after the contract has been deployed on the VM.

To show how this all works we will slowly start fleshing out the smart contract
functions of the `dividend` example. Here is the first part of the Rust code
that implements it, which contains the 'init' function:

```rust
// This example implements 'dividend', a simple smart contract that will
// automatically disperse iota tokens which are sent to the contract to a group
// of member addresses according to predefined division factors. The intent is
// to showcase basic functionality of WasmLib through a minimal implementation
// and not to come up with a complete robust real-world solution.

use wasmlib::*;

use crate::*;

// 'init' is used as a way to initialize a smart contract. It is an optional
// function that will automatically be called upon contract deployment. In this
// case we use it to initialize the 'owner' state variable so that we can later
// use this information to prevent non-owners from calling certain functions.
// The 'init' function takes a single optional parameter:
// - 'owner', which is the agent id of the entity owning the contract.
// When this parameter is omitted the owner will default to the contract creator.
pub fn func_init(ctx: &ScFuncContext) {

    // Log the fact that we have initiated the 'init' Func in the host log.
    ctx.log("dividend.init");

    // First we set up a default value for the owner in case the optional
    // 'owner' parameter was omitted.
    let mut owner: ScAgentId = ctx.contract_creator();

    // Now it is time to check if parameters were provided to the function.
    // We create an ScImmutableMap proxy to the params map on the host.
    let p: ScImmutableMap = ctx.params();

    // Then we create an ScImmutableAgentId proxy to the 'owner' parameter.
    let param_owner: ScImmutableAgentId = p.get_agent_id(PARAM_OWNER);

    // Now we check if the 'owner' parameter is present in the params map.
    if param_owner.exists() {
        // Yes, it was present, so now we overwrite the default owner with
        // the one specified by the 'owner' parameter.
        owner = param_owner.value();
    }

    // Now that we have sorted out which agent will be the owner of this contract
    // we will save this value in the state storage on the host. First we create
    // an ScMutableMap proxy that refers to the state storage map on the host.
    let state: ScMutableMap = ctx.state();

    // Then we create an ScMutableAgentId proxy to an 'owner' variable in state storage.
    let state_owner: ScMutableAgentId = state.get_agent_id(VAR_OWNER);

    // And then we save the owner value in the 'owner' variable in state storage.
    state_owner.set_value(&owner);

    // Finally, we log the fact that we have successfully completed execution
    // of the 'init' Func in the host log.
    ctx.log("dividend.init ok");
}
```

In the next sections we will go deeper into explaining how function parameters
are passed to a smart contract function and how to interact with the state
storage.

Next: [Function Parameters](Params.md)
