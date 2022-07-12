---
description: The `blobs` contract maintains a registry of _blobs_ (a collection of arbitrary binary data) which are referenced from smart contracts via their hashes.
image: /img/logo/WASP_logo_dark.png
keywords:
- core contracts
- bloc
- binary data
- store
- get
- entry points
- views
- reference
--- 
# The `blob` Contract

The `blob` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts chain.

The function of the `blob` contract is to maintain an on-chain registry of
_blobs_, a collections of arbitrary binary data. Smart contracts reference _blobs_ via their hashes.

A _blob_ is a collection of named pieces of arbitrary binary data:

```
<fieldName1> : <binaryChunk1>
<fieldName2> : <binaryChunk2>
...
<fieldNameN> : <binaryChunkN>
``` 

Here the `fieldNameK` is an arbitrary binary (a string) used as a name for the
binary data `binaryChunkK`. Usually `fieldNameK` is not long. Its interpretation
is use-case specific.

The `binaryChunkK` may be of arbitrary size (practical limits apply, of course).

The order of the field-chunk pairs is essential because the hash of the blob depends on it.

The hash of the _blob_ is equal to the hash of concatenation of all pieces:

```
blobHash = hash( fieldName1 || binaryChunk1 || fieldName2 || binaryChunk2 || ... || fieldNameN || binaryChunkN)
``` 

There are two predefined field names which are interpreted by the VM while
deploying smart contracts from binary:

- _fieldname_ = `"v"` is interpreted as a _VM type_
- _fieldname_ = `"p"` is interpreted as a _smart contract program binary_

If the field `"v"` is equal to the string `"wasmtimevm"`, the binary chunk
of `"p"` is interpreted as WebAssembly binary, loadable into the _Wasmtime_
Wasm VM.

Another use_case for a _blob_ may be a full collection of self-described
immutable data of a smart contract program:

```
"v" : VM type
"p" : smart contract program binary
"d" : data schema for data exchange between smart contract and outside sources and consumers
"s" : program sources in .zip format
```

---

## Entry Points

There is only one full entry point which allows us to submit a _blob_ to the `blob` contract:

### - `storeBlob()`

In the current implementation the data of the _blob_ is passed
as parameters to the call of the entry point. It may be practically impossible
to submit very large _blobs_ to the chain. In the future we plan to implement
a special mechanism which allows for the nodes to download big data chunks as
part of the committee consensus.

---

## Views

### - `getBlobInfo(hash BlobHash)`

Returns information about fields of the blob with specific hash and sizes of its data chunks:

```
<fieldName1>: <size of the dataChunk1>
...
<fieldNameN>: <size of the dataChunkN>
```

### - `getBlobField(hash BlobHash, field BlobField)`

Returns the data of the specified _blob_ field.

### -`listBlobs()`

Returns a list of pairs `blob hash`: `total size of chunks` for all blobs in the registry.
  