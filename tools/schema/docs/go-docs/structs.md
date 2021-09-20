## Structured Data Types

The schema tool allows you to define your own structured data types that are composed of
the predefined WasmLib value data types. The tool will generate a struct with named fields
according to the definition in schema.json, and also generates code to serialize and
deserialize the structure to a byte array, so that it can be saved as a single unit of
data, for example in state storage.

You can use such structs directly as a type in state storage definitions and the schema
tool will automatically generate the proxy code to access it properly.

For example, let's say you are creating a `betting` smart contract. Then you would want to
store information for each bet. The Bet structure could consist of the bet amount and time
of the bet, the number of the item that was bet on, and the agent ID of the one who placed
the bet. And you would keep track of all bets in state storage in an array of Bet structs.
You would insert the following into schema.json:

```json
{
  "structs": {
    "Bet": {
      "amount": "Int64 // bet amount",
      "time": "Int64 // timestamp of this bet",
      "number": "Int32 // number of item we bet on",
      "better": "AgentID // Who placed this bet"
    }
  },
  "state": {
    "bets": "[]Bet // all bets made in this round"
  }
}
```

The schema tool will generate `types.go` which contains the following code for the Bet
struct:

```golang
package betting

import "github.com/iotaledger/wasp/packages/vm/wasmlib"

type Bet struct {
	Amount int64             // bet amount
	Better wasmlib.ScAgentID // Who placed this bet
	Number int32             // number of item we bet on
	Time   int64             // timestamp of this bet
}

func NewBetFromBytes(bytes []byte) *Bet {
	decode := wasmlib.NewBytesDecoder(bytes)
	data := &Bet{}
	data.Amount = decode.Int64()
	data.Better = decode.AgentID()
	data.Number = decode.Int32()
	data.Time = decode.Int64()
	decode.Close()
	return data
}

func (o *Bet) Bytes() []byte {
	return wasmlib.NewBytesEncoder().
		Int64(o.Amount).
		AgentID(o.Better).
		Int32(o.Number).
		Int64(o.Time).
		Data()
}

type ImmutableBet struct {
	objID int32
	keyID wasmlib.Key32
}

func (o ImmutableBet) Exists() bool {
	return wasmlib.Exists(o.objID, o.keyID, wasmlib.TYPE_BYTES)
}

func (o ImmutableBet) Value() *Bet {
	return NewBetFromBytes(wasmlib.GetBytes(o.objID, o.keyID, wasmlib.TYPE_BYTES))
}

type MutableBet struct {
	objID int32
	keyID wasmlib.Key32
}

func (o MutableBet) Exists() bool {
	return wasmlib.Exists(o.objID, o.keyID, wasmlib.TYPE_BYTES)
}

func (o MutableBet) SetValue(value *Bet) {
	wasmlib.SetBytes(o.objID, o.keyID, wasmlib.TYPE_BYTES, value.Bytes())
}

func (o MutableBet) Value() *Bet {
	return NewBetFromBytes(wasmlib.GetBytes(o.objID, o.keyID, wasmlib.TYPE_BYTES))
}
```

Notice how the generated ImmutableBet and MutableBet proxies use the Bytes() and
NewBetFromBytes() (de)serialization code to automatically transform byte arrays into Bet structs.

The generated code in `state.go` that implements the state interface is shown here:

```golang
package betting

import "github.com/iotaledger/wasp/packages/vm/wasmlib"

type ArrayOfImmutableBet struct {
	objID int32
}

func (a ArrayOfImmutableBet) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfImmutableBet) GetBet(index int32) ImmutableBet {
	return ImmutableBet{objID: a.objID, keyID: wasmlib.Key32(index)}
}

type ImmutableBettingState struct {
	id int32
}

func (s ImmutableBettingState) Bets() ArrayOfImmutableBet {
	arrID := wasmlib.GetObjectID(s.id, idxMap[IdxStateBets], wasmlib.TYPE_ARRAY|wasmlib.TYPE_BYTES)
	return ArrayOfImmutableBet{objID: arrID}
}

func (s ImmutableBettingState) Owner() wasmlib.ScImmutableAgentID {
	return wasmlib.NewScImmutableAgentID(s.id, idxMap[IdxStateOwner])
}

type ArrayOfMutableBet struct {
	objID int32
}

func (a ArrayOfMutableBet) Clear() {
	wasmlib.Clear(a.objID)
}

func (a ArrayOfMutableBet) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfMutableBet) GetBet(index int32) MutableBet {
	return MutableBet{objID: a.objID, keyID: wasmlib.Key32(index)}
}

type MutableBettingState struct {
	id int32
}

func (s MutableBettingState) Bets() ArrayOfMutableBet {
	arrID := wasmlib.GetObjectID(s.id, idxMap[IdxStateBets], wasmlib.TYPE_ARRAY|wasmlib.TYPE_BYTES)
	return ArrayOfMutableBet{objID: arrID}
}

func (s MutableBettingState) Owner() wasmlib.ScMutableAgentID {
	return wasmlib.NewScMutableAgentID(s.id, idxMap[IdxStateOwner])
}
```

The end result is an ImmutableBettingState and MutableBettingState structure that can
directly interface to the state of the betting contract.

In the next section we will look at how to make even more complex type definitions.

Next: [Type Definitions](typedefs.md)
