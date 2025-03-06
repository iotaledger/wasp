# BCS serialization

## What can this library do

* Serialize basic types: **bool**, **int8**, **int16**, **int32**, **int64**, **int**, **uint8**, **uint16**, **uint32**, **uint64**, **uint**, **string**.
* Serialize extended types: **time.Time**, **big.Int**, **ULEB128** (variable-length int).
* Recursively serialize complex types: **structures**, **arrays**, **slices** and **maps**.
* Define **enumerations** in form of **interface** types or **struct** types.
* Define **custom encoders/decoders** through functors or methods.
* Define **custom initializer** to be executed after decoding.
* Define **type parameters** using type's method or structure field tag.

## Usage

#### Marshaling/Unmarshaling

###### Bytes:

```
v := "hello"
vEncoded := bcs.MustMarshal(&v)
vDecoded := bcs.MustUnmarshal(vEncoded)
```

###### Stream:

```
v := "hello"
var wBuff bytes.Buffer
bcs.MustMarshalStream(&v, &wBuff)
vEncoded := wBuff.Bytes()

rBuff := bytes.NewReader(vEncoded)
vDecoded := bcs.MustUnmarshal(vEncoded)
```

###### Into existing value:

```
...
var vDecodedFromBytes string
bcs.MustUnmarshalInto(vEncoded, &vDecodedFromBytes)

...
var vDecodedFromStream string
bcs.MustUnmarshalStreamInto(ioStreamReader, &vDecodedFromStream)
```

#### Using encoder/decoder

```
v1 := "hello"
v2 := 123
var wBuff bytes.Buffer
enc := bcs.NewEncoder(wBuff)
enc.MustEncode(v1)
enc.MustEncode(&v2) // note both value and pointer supported
encoded := wBuff.Bytes()

dec := bcs.NewDecoder(bytes.NewReader(encoded))
dec.MustDecode(&v1)
dec.MustDecode(&v2)
```

**NOTE:** Althouth `Encode`() supports both value and pointer as argument, try prefering passing a pointer (see perf section)

#### Using BytesEncoder/BytesDecoder

```
v1 := "hello"
v2 := 123
enc := bcs.NewBytesEncoder()
enc.MustEncode(v1)
enc.MustEncode(&v2) // note both value and pointer supported
encoded := enc.Bytes()

dec := bcs.NewBytesDecoder(encoded)
dec.MustDecode(&v1)
dec.MustDecode(&v2)
```

#### Decoder with helper functions

```
dec := bcs.NewDecoder(bytes.NewReader(encoded))
v1 := bcs.MustDecode[string](dec)
v2 := bcs.MustDecode[int](dec)
```

#### Using specialized functions of encoder/decoder

```
var wBuff bytes.Buffer
enc := bcs.NewEncoder(wBuff)
enc.WriteString("hello") // will have better performance than enc.Encode()
enc.WriteInt(123)
if enc.Err() != nil {
   return enc.Err()
}
encoded := wBuff.Bytes()

dec := bcs.NewDecoder(bytes.NewReader(encoded))
v1 := dec.ReadString()  // will have better performance than dec.Decode()
v2 := dec.ReadInt()
if dec.Err() != nil {
   return enc.Err()
}

```

#### Error handling

###### Classic:

```
if err := bcs.Marshal(&v); err != nil {
   return err
}
```

###### Deffered:

```
dec := bcs.NewDecoder(bytes.NewReader(encoded))
v1 := bcs.Decode[string](dec) // will return empty value if error happen
v2 := bcs.Decode[int](dec)    // if there was on error on previous line, this line will just do nothing and return empty value

if err := dec.Err(); err != nil {
   return err
}
```

###### Panic:

```
v := bcs.MustUnmarshal[string](encoded)
```

###### Error checked automatically after MarshalBCS/UnmarshalBCS and custom serialization functions

```
func (p *MyStruct) MarshalBCS(e *bcs.Encoder) error {
    e.Encode(&p.Field1)
    e.Encode(&p.Field2) // if an error happened on previous line, this line will just do nothing
    return nil // the encoder will automatically check if there was an error after call to MarshalBCS
}
```

## Complex types

#### Structures

Structures are encoded/decoded same way as basic types. Fields are encoded in the order of their definition. Fields may have any encodable type, including nested structs and collections.

Only **public** **fields** of structure are serialized **by default**. To include **private fields** use  `bcs:"export"`.

To **exclude field** from serialization, use \`bcs:"-"\`.

```
type TestStruct struct {
   A int
   B string
   C []byte `bcs:"-"` // excluded
   d bool   `bcs:"export"` // notice fiels is unexported, but we force it to be exported through BCS
}

v := TestStruct{10, "hello", true}

vEncoded := bcs.MustMarshal(&v)

// This in equivalent to:
// var wBuff bytes.Buffer
// enc := bcs.NewEncoder(&wBuff)
// enc.WriteInt(v.A)
// enc.WriteString(v.B)
// enc.WriteString(v.d)
```

#### Arrays of constant length

Elements of array are written one by one. They could be of any encodable type (e.g. structs or collections):

```
var v [3]int8 = {1, 2, 3}
vEncoded := bcs.MustMarshal(&v)
// vEncoded = []byte{1, 2, 3}
```

**WARNING:** Avoid using construct `v[:]`, because it will result in array being treated as slice, so length will be encoded in addition to elements.

#### Slices

```
// Slices are encoded same way as arrays except that they also have length encoded at start as variable-length int
var v []int8 = {1, 2, 3}
vEncoded := bcs.MustMarshal(&v)
// vEncoded = []byte{3, 1, 2, 3}
```

#### Maps

Maps are encoded as slice of key-value entries, ordered by byte representation of keys:

```
v := map[int8]int8{2: 10, 3: 5, 1: 20}
vEncoded := bcs.MustMarshal(&v)
// vEncoded = []byte{3, 1, 20, 2, 10, 3, 5}
```

#### Strings

Strings encoded as byte slices:

```
v := "abc"
vEncoded := bcs.MustMarshal(&v)
// vEncoded = []byte{0x3, 0x61, 0x62, 0x63}
```

## Pointers

Upon encoding pointers are **dereferenced**. Pointer value **must** be either **non-nil** or marked as **optional** (see other sections). Otherwise encoding will fail with error.

Upon decoding, for each **nil** pointer new value is **allocated** and assigned. If pointer field was preset before decoding with non-nil value, the **preset** value is **kept**.

```
type testStruct struct {
	A **bool
}

a := true
pa := &a
v := &testStruct{A: &pa}
pv := &v

vEnc := bcs.MustMarshal(&pv)
vDec := bcs.MustUnmarshal[****testStruct](vEnc)
require.Equal(t, v, ***vDec)
```

## Interfaces

Interface can be serialized in three different ways:

* **By default**, interface is encoded just as plain value. For decoding it **must** contain some value to specify an actual type.
* If interface is registered as **enumeration** (see sections below), the library uses its registered enum specification for serialization. In that case, no need to preset value on decoding.
* If structure's interface field is marked as **"not_enum"** (see sections below), the library ignores enum registration for that field.
* If interface has custom serialization **functor**, that functor is used for serialization and everything else is ignored.

If interface does not satisfy any of those criterias, serialization will return an **error**.

###### Encoding enum interface's value using Marshal/Encode

If you need to encode the value wrapped by the interface, which is registered as enum, you can pass interface **by value** to `Encode()` of `bcs.Encoder`:

```
type Message interface{}

var m Message = "hello"
e := bcs.NewBytesEncoder()
e.Encode(m)
```

This works, because `Encode()` function has argument of type `any`, so information about `Message` interface is lost - value in unpacked from `Message` and packed as `any`.

With `bcs.Marshal` this won't work, because it enforces pointer argument, so passing interface by value is not an option.
For that purpose, you can either cast interface to actual value (`bcs.Marshal(m.(string)`). But if the type is unknown, you can cast to `any` and then take its pointer:

```
...
var eI any = m
bcs.Marshal(&eI)

// Or:

bcs.Marshal(lo.ToPtr[any](m))
```

## Enumerations

BCS supports enumerations - values of variable type from the fixed list of types. It is implemented as enumeration variant index being encoded as variabl-length int before encoding the value.
Current BCS library supports two different styles of defining enumerations: struct enums and interface enums.

#### Struct enumerations

Struct enumeration is defined by defining a structure.
The structure:

* Must have all fields of nullable (e.g. pointer) and encodable types.
* Must define method **IsBcsEnum** with **value receiver**.

**One and only one** **fields** of that structure **must** be set when encoding.
Index of the field is used as index of enum variant.

```
type TestEnum struct {
   A *int16
   B *string
   C *bool
}

func (TestEnum) IsBcsEnum() {}

var a int16 = 10
bcs.Marshal(&TestEnum{A: &a}) // []byte{0, 10, 0}
b := "abc"
bcs.Marshal(&TestEnum{B: &b}) // []byte{1, 0x3, 0x61, 0x62, 0x63}
c := true
bcs.Marshal(&TestEnum{C: &c}) // []byte{2, 1}
```

#### Interface enumerations

Having separate structure type for enumeration is sometimes not an elegant solution. So there is another way to define an enum - based on interface.

* Enum type **must be an interface**. It can have methods.
* All of the variant types must **implement** **the interface** (except for `bcs.None` - see below). If variant implements interface with **pointer receiver**, pointer to a type must be used as a variant type.
* Variant type **cannot be an interface** type.
* Enum must be **registered** using one of `bcs.RegisterEnum*(...)` functions.

```
type TestEnum interface{}

var _ = bcs.RegisterEnumType3[TestEnum, int16, string, bool]

var v TestEnum = int16(10)
bcs.Marshal(&v) // []byte{0, 10, 0}
v = "abc"
bcs.Marshal(&v) // []byte{1, 0x3, 0x61, 0x62, 0x63}
v = true
bcs.Marshal(&v) // []byte{2, 1}
```

###### bcs.None

For absense of the value special type `bcs.None` can be used as variant.
Encoding interface enum type with nil value will result in an error if enum does not have bcs.None registered as one of variants.

```
...
var _ = bcs.RegisterEnumType3[TestEnum, bcs.None, int16, string, bool]

v = nil
bcs.Marshal(&v) // []byte{0}
```

###### Passing an interface to Encode

**WARNING**: Methods `Encode()` of types `bcs.Encoder` expects encoded value to be passed through argument of type `any`. That type is an interface, which means that when passing enum interface to it the value will be unwrapped and wrapped again into `any` thus loosing the information about initial interface type.
To avoid that, pass it as pointer: `e.Encode(&v)`. That way the information about original interface type is preserved.
There is no such issue with `bcs.Marshal` because its argument enforces pointer value.

## Customization

#### Customizing struct field

Each struct field encoding/decoding can be customized using struct field tag metainformation.
There are three tags available:

* **bcs** - customizes serialization of field value
* **bcs_elem** - customizes serialization of array/slice elememt if field is array/slice.
* **bcs_key / bcs_value** - customizes serialization of map key/value if field is a map.

Example:

```
type TestStruct struct {
   I *int        `bcs:"optional"`
   hidden bool   `bcs:"export"`
   Excluded bool `bcs:"-"`
   S []*int64    `bcs:"len_bytes=4" bcs_elem:"compact"`
   M map[int]int `bcs:"len_bytes=2" bcs_key:"type=int32" bcs_value:"bytearr"`
   F interface{} `bcs:"not_enum"`
}
```

For the list of available customizations see next sections.

#### Customizing type

Method **BCSOptions()** can be defined to customize type serialization.
NOTE: The method **must** have value receiver.

```
type CompactInt int

func (CompactInt) BCSOptions() bcs.TypeOptions {
   return bcs.TypeOptions{IsCompactInt: true}
}
```

It can also be used to customize element/key/value serialization of arrays/slices/maps:

```
type MapOfCompactInts map[string]int64

func (MapOfCompactInts) BCSOptions() bcs.TypeOptions {
   return bcs.TypeOptions{
      MapValue: &bcs.TypeOptions{
         IsCompactInt: true,
      },
   }
}
```

#### Manual serialization

Serizalization of a type can fully customized by implementing custom encoder/decoder.
Custom serialization can be provided **for** **any type:** basic types, structs, collections, pointers, third-part types etc.

There are multiple methods available in Encoder and Decoder to help implement manual serialization, including serialization of optionals, collections and enumerations (see sections below).

NOTE: No need to manually check an **error** inside of custom serialization methods and functions - it will be checked automatically by the encoder/decoder after the call.

###### MarshalBCS/UnmarshalBCS methods:

One or both methods of the following methods can be defined to implement manual serialization.

* Encoding - **MarshalBCS**. Supports both **value receiver** and **pointer receiver**.
* Decoding - **UnmarshalBCS**. It **must** have **pointer receiver**.

```
type TestStruct struct {
   A int
   ...
}

func (s *TestStruct) MarshalBCS(e *bcs.Encoder) ([]byte, error) {
   e.WriteInt(s.A)
   ...   
}

func (s *TestStruct) UnmarshalBCS(d *bcs.Decoder) error {
   s.A = d.ReadInt()
   ...
}
```

###### Read/Write methods:

In addition to MarshalBCS/UnmarshalBCS, another pair of methods is supported: **Read()** and **Write()**.
This is done to ensure compatibility with some others libraries.

```
type TestStruct struct {
   A int
   ...
}

func (s *TestStruct) Write(w io.Writer) error {
   e := bcs.NewEncoder(w)  
   e.WriteInt(s.A)
   ...
}

func (s *TestStruct) Read(r io.Reader) error {
   d := bcs.NewDecoder(r)
   e.A = d.ReadInt()
   ...
}

```

###### Custom serialization functors:

Special functions can be registered to customize serialization of a type. This has two **advantages** over using methods:

* Allows to implement custom serialization for **third-party types** (like it is implemented for **time.Time** and **big.Int**).
* Allows to implement custom serialization for **interfaces**.

It is convenient (but not required) to run register them upon program initialization using Golang's package **init()** function.
It is permitted to register **separate functors** for the **type itself** and its **pointer type**.

```
func init() {
   AddCustomEncoder(func(e *Encoder, v time.Time) error {
      e.w.WriteInt64(v.UnixNano())
      return e.w.Err
   })

   AddCustomDecoder(func(d *Decoder, v *time.Time) error {
      *v = time.Unix(0, d.r.ReadInt64())
      return d.r.Err
   })
}

```

#### Custom initialization

Type can have custom initializer function to be executed after the value is decoded.
It is implemented by defining **BCSInit** method. It **must** have **pointer receiver**.

Serialization will **fail** if unexported field has BCS tag, but is not marked as "export". Reason: such case signals mixed intention.
Serialization will also **fail** if already exported field has "export" tag. Reason: engineers might have a standard expectation, that when field is renamed it becomes hidden. But they may forget about BCS serialization.

```
type TestStruct struct {
   A int
   B int `bcs:"-"`
}

func (s *TestStruct) BCSInit() error {
	s.B = s.A * 2
}
```

#### Available field tags

###### "export"

Marks unexported field to be exported through BCS.
Appicable to: **unexported fields of structrures**.

```
type TestStruct struct {
   unexportedField         int                   // This field won't be serialized because it is unexported
   unexportedButSerialized string `bcs:"export"` // Although this field is unexported too, it will be serialized 
}
```

###### "optional"

Marks field as optional.
Applicable to: **pointers**, **interfaces, maps, slices**.

When encoding such field, first presense flag is encoded as 1 byte, and then value itself is encoded if present.

```
type TestStruct struct {
   A *int16 `bcs:"optional"`
}

var a int16 = 10
withValue    := bcs.MustMarshal(&TestStruct{A: &a})  // []byte{1, 10, 0}
withoutValue := bcs.MustMarshal(&TestStruct{A: nil}) // []byte{0}
```

###### "compact"

Mark integer field to be written as **ULEB128** - variable-length integer which enhances space usage but decreases serialization performance. This is the same format used to serialize collection length or enumeration variant index.
Applicable to: **integers**.

**WARNING:** BCS specification does not have mentions about ULEB128 being used for anything other than length of collections and enumeration variant indexes. So this logic is a custom extension mostly designed for usage with types, which are not used for interaction with other actors.

###### "type=T"

Allows to **override type** of integer value. For example, if field has type int64, using this tag you can store it as int16.
Applicable to: **integers**.
Possible values of **T**: i8, i16, i32, i64, u8, u16, u32, u64 (and corresponding aliases int8, uint16, ...).

```
type TestStruct struct {
   A int64 `bcs:"type=i16"`
}

bcs.Marshal(&TestStruct{A: 10}) // []byte{10, 0}
```

**WARNING:** In case of overflow an **error** is returned. This is done to ensure, that field is not accidentally missused when type from definition is bigger than serialized type.

###### "len_bytes=N"

Sets size limitation for length of a collection.
Applicable to: **slices**, **maps**.
Possible values of **N**: 2, 4.

###### "nil_if_empty"

Deserialize empty slice into `nil` instead of `[]ElemType{}`.
Applicable to: **slices.**

**NOTE:** you can also achive same effect by implementing that logic in `BCSInit()` method.

###### "bytearr"

Marks value to be written as slice of bytes.
Appicable to: **any type**.

The only difference this flag introduces is that value is prepended by the count of its bytes. This might be useful to be able to skip value without knowing its actual structure.

```
type TestStruct struct {
   A int32
   B int32 `bcs:"bytearr"
}

bcs.Marshal(&TestStruct{A: 10, B: 10}) // []byte{
                                       //       10, 0, 0, 0,   <-- A
                                       //    4, 10, 0, 0, 0,   <-- B (notice 4 added the bytes of a value)
                                       // }
```

###### "not_enum"

Forces interface field to be encoded/decoded as plain value and not as enumeration.
Applicable to: **interfaces, that are registered as enums.**

## Performance considerations

#### Prefer passing pointer into Encode()

Unlike from function `Marshal`(), method `Encode`() accepts both value and pointer. But passing by value will force encoder to copy value to make it addressable to support `MarshalBCS`() method with pointer receiver. It will also not work properly when interface is passed by value, because value is unpacked and packed again as `any` thus information about type of initial interface will be lost.
So it is better to pass a pointer always when it is easy to do.

#### Serialization of byte arrays is optimized

Arrays of bytes, whose elements does not have any customizations, are directly copied into/from the stream.

#### Serialization of arrays of intergers is NOT yet optimized

If elements of array are of integer type, and they dont have any customization specified for them (except of `"type=T"`), serialization of such array could be optimized to avoid redudant calls for each array element. But this not done specifically to keep code simple while it is maturing.

#### Type parsing is cached

Upon serialization the types are checked for the presense of customizations. This make take significant time.
To improve that, the type information is stored in cache.
To avoid global mutex locks when reading/writing cache, atomic swapping is used instead.
It works like that:

* When coder is created, it get pointer to the current version of cache.
* Current version of cache is only read, never written, so multiple encoders can read it at the same time.
* When coder is done with coding, it creates new cache with updated information.
* Then it atomically swaps the pointer to the new cache with the pointer to the current cache.
* Multiple coders may update cache, thus overwritting modifications of each other. But it is not a problem, because
  the info is only extended by them, so eventually cache will have information about all types.
