## Type Definitions

Since we allow nesting of container types it is a bit difficult to create proper
declarations for such nested types. Especially because in a field definition you can only
indicate either a single type, or an array of single type, or a map of single type.

We devised a simple solution to this problem. You can add a `typedefs` section to
schema.json where you can define a single type name for a container type. That way you can
easily create containers that contain such container types. The schema tool will
automatically generate the in-between proxy types necessary to make all of this work.

To keep it at the `betting` smart contract from before, imagine we would want to keep
track of all betting rounds. Since a betting round contains an array of all bets in a
round you could not easily define it if it weren't for typedefs.

Instead, now you add the following to your schema.json:

```json
{
  "typedefs": {
    "BettingRound": "[]Bet // one round of bets"
  },
  "state": {
    "rounds": "[]BettingRound // keep track of all betting rounds"
  }
}
```

The schema tool will generate the following proxies in `typedefs.go`:

```golang
package betting

import "github.com/iotaledger/wasp/packages/vm/wasmlib"

type ImmutableBettingRound = ArrayOfImmutableBet

type ArrayOfImmutableBet struct {
	objID int32
}

func (a ArrayOfImmutableBet) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfImmutableBet) GetBet(index int32) ImmutableBet {
	return ImmutableBet{objID: a.objID, keyID: wasmlib.Key32(index)}
}

type MutableBettingRound = ArrayOfMutableBet

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
```

Note how ImmutableBettingRound and MutableBettingRound type aliases are created for the
types ArrayOfImmutableBet and ArrayOfMutableBet. These are subsequently used in the state
definition in `state.go`:

```golang
package betting

import "github.com/iotaledger/wasp/packages/vm/wasmlib"

type ArrayOfImmutableBettingRound struct {
	objID int32
}

func (a ArrayOfImmutableBettingRound) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfImmutableBettingRound) GetBettingRound(index int32) ImmutableBettingRound {
	subID := wasmlib.GetObjectID(a.objID, wasmlib.Key32(index), wasmlib.TYPE_ARRAY|wasmlib.TYPE_BYTES)
	return ImmutableBettingRound{objID: subID}
}

type ImmutableBettingState struct {
	id int32
}

func (s ImmutableBettingState) Owner() wasmlib.ScImmutableAgentID {
	return wasmlib.NewScImmutableAgentID(s.id, idxMap[IdxStateOwner])
}

func (s ImmutableBettingState) Rounds() ArrayOfImmutableBettingRound {
	arrID := wasmlib.GetObjectID(s.id, idxMap[IdxStateRounds], wasmlib.TYPE_ARRAY|wasmlib.TYPE_BYTES)
	return ArrayOfImmutableBettingRound{objID: arrID}
}

type ArrayOfMutableBettingRound struct {
	objID int32
}

func (a ArrayOfMutableBettingRound) Clear() {
	wasmlib.Clear(a.objID)
}

func (a ArrayOfMutableBettingRound) Length() int32 {
	return wasmlib.GetLength(a.objID)
}

func (a ArrayOfMutableBettingRound) GetBettingRound(index int32) MutableBettingRound {
	subID := wasmlib.GetObjectID(a.objID, wasmlib.Key32(index), wasmlib.TYPE_ARRAY|wasmlib.TYPE_BYTES)
	return MutableBettingRound{objID: subID}
}

type MutableBettingState struct {
	id int32
}

func (s MutableBettingState) Owner() wasmlib.ScMutableAgentID {
	return wasmlib.NewScMutableAgentID(s.id, idxMap[IdxStateOwner])
}

func (s MutableBettingState) Rounds() ArrayOfMutableBettingRound {
	arrID := wasmlib.GetObjectID(s.id, idxMap[IdxStateRounds], wasmlib.TYPE_ARRAY|wasmlib.TYPE_BYTES)
	return ArrayOfMutableBettingRound{objID: arrID}
}
```

Notice how the Rounds() member function returns a proxy to an array of BettingRound. Which
in turn is an array of Bet. So the desired result has been achieved. And every access step
along the way only allows you to take the path laid out which is checked at compile-time.

In the next section we will explore how the schema tool helps to simplify function
definitions.

Next: [Function Definitions](funcs.md)
