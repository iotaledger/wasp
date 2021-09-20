## Function Parameters

The optional `params` subsection contains field definitions for each of the parameters
that a function takes. The layout of the field definitions is identical to that of
the [state](state.md) field definitions, with one addition. The field type can be prefixed
with a question mark, which indicates that that parameter is optional.

The schema tool will automatically generate an immutable structure with member variables
for proxies to each parameter variable in the `params` map. It will also generate code to
check the presence of each non-optional parameter, and it will also verify the parameter's
data type. This checking is done before the function is called. The user will be able to
immediately start using the parameter proxy through the structure that is passed to the
function.

When this subsection is empty or completely omitted, no structure will be generated or
passed to the function.

For example, here is the structure generated for the immutable params for the `member`
function:

```golang
type ImmutableMemberParams struct {
    id int32
}

func (s ImmutableMemberParams) Address() wasmlib.ScImmutableAddress {
    return wasmlib.NewScImmutableAddress(s.id, idxMap[IdxParamAddress])
}

func (s ImmutableMemberParams) Factor() wasmlib.ScImmutableInt64 {
    return wasmlib.NewScImmutableInt64(s.id, idxMap[IdxParamFactor])
}
```

Note that the schema tool will also generate a mutable version of the structure, suitable
for providing the parameters when calling this smart contract function.

In the next section we will look at the `results` subsection.

Next: [Function Results](results.md)

