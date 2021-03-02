## View Functions

```Rust
// 'getFactor' is a simple example of a View function. It will retrieve the
// factor associated with the (mandatory) address parameter it was provided with.
pub fn view_get_factor(ctx: &ScViewContext) {

    // Log the fact that we have initiated the 'getFactor' View in the log on the host.
    ctx.log("dividend.getFactor");

    // Now it is time to check the parameter.
    // First we create an ScImmutableMap proxy to the params map on the host.
    let p = ctx.params();

    // Create an ScImmutableAddress proxy to the 'address' parameter that is still stored
    // in the map on the host.
    let param_address= p.get_address(PARAM_ADDRESS);

    // Require that the mandatory 'address' parameter actually exists in the map on the host.
    // If it doesn't we panic out with an error message.
    ctx.require(param_address.exists(), "missing mandatory address");

    // Now that we are sure that the 'address' parameter actually exists we can
    // retrieve its actual value into an ScAddress value object
    let address = param_address.value();

    // Now that we have sorted out the parameter we will access the state storage on the host.
    // First we create an ScImmutableMap proxy to the state storage map on the host.
    // Note that this is an *immutable* map, as opposed to the mutable map we get when
    // we call the state() method on an ScFuncContext.
    let state = ctx.state();

    // Create an ScImmutableMap proxy to the 'members' map in the state storage.
    // Note that again, this is an *immutable* map as opposed to the mutable map we
    // get from the mutable state map we get through ScFuncContext.
    let members = state.get_map(VAR_MEMBERS);

    // Retrieve the factor associated with the address parameter through
    // an ScImmutableInt64 proxy to the value stored in the 'members' map.
    let factor = members.get_int64(&address).value();

    // Create an ScMutableMap proxy to the map on the host that will store the
    // key/value pairs that we want to return from this View function
    let results = ctx.results();

    // Set the value associated with the 'factor' key to the factor we got from
    // the members map through an ScMutableInt64 proxy to the results map.
    results.get_int64(VAR_FACTOR).set_value(factor);

    // Finally, we log the fact that we have successfully completed execution
    // of the 'getFactor' View in the log on the host.
    ctx.log("dividend.getFactor ok");
}
```

Next: [Calling Functions](Calls.md)