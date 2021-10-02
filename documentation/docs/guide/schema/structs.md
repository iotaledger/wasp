# Structured Data Types

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

The schema tool will generate `types.rs` which contains the following code for the Bet
struct:

```go
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

```rust
// @formatter:off

#![allow(dead_code)]

use wasmlib::*;
use wasmlib::host::*;

pub struct Bet {
    pub amount: i64,       // bet amount
    pub better: ScAgentID, // Who placed this bet
    pub number: i32,       // number of item we bet on
    pub time:   i64,       // timestamp of this bet
}

impl Bet {
    pub fn from_bytes(bytes: &[u8]) -> Bet {
        let mut decode = BytesDecoder::new(bytes);
        Bet {
            amount: decode.int64(),
            better: decode.agent_id(),
            number: decode.int32(),
            time: decode.int64(),
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut encode = BytesEncoder::new();
        encode.int64(self.amount);
        encode.agent_id(&self.better);
        encode.int32(self.number);
        encode.int64(self.time);
        return encode.data();
    }
}

pub struct ImmutableBet {
    pub(crate) obj_id: i32,
    pub(crate) key_id: Key32,
}

impl ImmutableBet {
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BYTES)
    }

    pub fn value(&self) -> Bet {
        Bet::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_BYTES))
    }
}

pub struct MutableBet {
    pub(crate) obj_id: i32,
    pub(crate) key_id: Key32,
}

impl MutableBet {
    pub fn exists(&self) -> bool {
        exists(self.obj_id, self.key_id, TYPE_BYTES)
    }

    pub fn set_value(&self, value: &Bet) {
        set_bytes(self.obj_id, self.key_id, TYPE_BYTES, &value.to_bytes());
    }

    pub fn value(&self) -> Bet {
        Bet::from_bytes(&get_bytes(self.obj_id, self.key_id, TYPE_BYTES))
    }
}

// @formatter:on
```

Notice how the generated ImmutableBet and MutableBet proxies use the from_bytes() and
to_bytes() (de)serialization code to automatically transform byte arrays into Bet structs.

The generated code in `state.rs` that implements the state interface is shown here:

```go
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

```rust
#![allow(dead_code)]
#![allow(unused_imports)]

use wasmlib::*;
use wasmlib::host::*;

use crate::*;
use crate::keys::*;
use crate::types::*;

pub struct ArrayOfImmutableBet {
    pub(crate) obj_id: i32,
}

impl ArrayOfImmutableBet {
    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }

    pub fn get_bet(&self, index: i32) -> ImmutableBet {
        ImmutableBet { obj_id: self.obj_id, key_id: Key32(index) }
    }
}

#[derive(Clone, Copy)]
pub struct ImmutableBettingState {
    pub(crate) id: i32,
}

impl ImmutableBettingState {
    pub fn bets(&self) -> ArrayOfImmutableBet {
        let arr_id = get_object_id(self.id, idx_map(IDX_STATE_BETS), TYPE_ARRAY | TYPE_BYTES);
        ArrayOfImmutableBet { obj_id: arr_id }
    }

    pub fn owner(&self) -> ScImmutableAgentID {
        ScImmutableAgentID::new(self.id, idx_map(IDX_STATE_OWNER))
    }
}

pub struct ArrayOfMutableBet {
    pub(crate) obj_id: i32,
}

impl ArrayOfMutableBet {
    pub fn clear(&self) {
        clear(self.obj_id);
    }

    pub fn length(&self) -> i32 {
        get_length(self.obj_id)
    }

    pub fn get_bet(&self, index: i32) -> MutableBet {
        MutableBet { obj_id: self.obj_id, key_id: Key32(index) }
    }
}

#[derive(Clone, Copy)]
pub struct MutableBettingState {
    pub(crate) id: i32,
}

impl MutableBettingState {
    pub fn bets(&self) -> ArrayOfMutableBet {
        let arr_id = get_object_id(self.id, idx_map(IDX_STATE_BETS), TYPE_ARRAY | TYPE_BYTES);
        ArrayOfMutableBet { obj_id: arr_id }
    }

    pub fn owner(&self) -> ScMutableAgentID {
        ScMutableAgentID::new(self.id, idx_map(IDX_STATE_OWNER))
    }
}
```

The end result is an ImmutableBettingState and MutableBettingState structure that can
directly interface to the state of the betting contract.

In the next section we will look at how to make even more complex type definitions.
