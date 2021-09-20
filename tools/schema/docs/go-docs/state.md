## Smart Contract State

The smart contract state storage on the host consists of a single key/value map. Both key
and value are raw data bytes. As long as we access the data in the same way that we used
to store it we will always get valid data back. The schema tool will create a type-safe
access layer to make sure that data storage and retrieval always uses the expected data
type.

The `state` section in schema.json contains a number of field definitions that together
define the variables that are stored in the state storage. Each field definition uses a
JSON key/value pair to define the name and data type of the field. The JSON key defines
the field name. The JSON value (a string) defines the field's data type, and can be
followed by an optional comment that describes the field.

The schema tool will use this information to generate the specific code that accesses the
state variables in a type-safe way. Let's examine the `state` section of the `dividend`
example in more detail:

```json
{
  "state": {
    "memberList": "[]Address // array with all the recipients of this dividend",
    "members": "[Address]Int64 // map with all the recipient factors of this dividend",
    "owner": "AgentID // owner of contract, the only one who can call 'member' func",
    "totalFactor": "Int64 // sum of all recipient factors"
  }
}
```

Let's start with the simplest state variables. `totalFactor` is an Int64, and `owner` is
an AgentID. Both are predefined [WasmLib value types](types.md).

Then we have the `memberList` variable. The empty brackets `[]` indicate that this is an
array. The brackets are immediately followed by the homogenous type of the elements in the
array, which in this case is the predefined Address value type.

Finally, we have the `members` variable. The non-empty brackets `[]` indicate that this is
a map. Between the brackets is the homogenous type of the keys, which in this case are of
the predefined Address type. The brackets are immediately followed by the homogenous type
of the values in the map, which in this case are of the predefined Int64 type.

Here is part of the Go code in `state.go` that the schema tool has generated. The
MutableDividendState struct defines a type-safe interface to access each of the state
variables through mutable proxies:

```golang
type MutableDividendState struct {
    id int32
}

func (s MutableDividendState) MemberList() ArrayOfMutableAddress {
    arrID := wasmlib.GetObjectID(s.id, idxMap[IdxStateMemberList], wasmlib.TYPE_ARRAY|wasmlib.TYPE_ADDRESS)
    return ArrayOfMutableAddress{objID: arrID}
}

func (s MutableDividendState) Members() MapAddressToMutableInt64 {
    mapID := wasmlib.GetObjectID(s.id, idxMap[IdxStateMembers], wasmlib.TYPE_MAP)
    return MapAddressToMutableInt64{objID: mapID}
}

func (s MutableDividendState) Owner() wasmlib.ScMutableAgentID {
    return wasmlib.NewScMutableAgentID(s.id, idxMap[IdxStateOwner])
}

func (s MutableDividendState) TotalFactor() wasmlib.ScMutableInt64 {
    return wasmlib.NewScMutableInt64(s.id, idxMap[IdxStateTotalFactor])
}
```

As you can see the schema tool has generated a proxy interface for the mutable `dividend`
state, called `MutableDividendState`. It has a 1-to-1 correspondence to the `state`
section in schema.json. Each member function accesses a type-safe proxy object for the
corresponding variable. In addition, the schema tool generates any necessary intermediate
map and array proxy types that force the usage of their respective homogenous types. In
the above example both `ArrayOfMutableAddress` and `MapAddressToMutableInt64` are examples
of such automatically generated proxy types. See the full `state.go` for more details.

In the next section we will look at how to define our own structured data types in the
schema definition file.

Next: [Structured Data Types](structs.md)
