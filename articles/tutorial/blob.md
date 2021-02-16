## The `blob` contract

The `blob` contract is one of 4 [core contracts](coresc.md) on each ISCP chain.
 
Function of the `blob` contract is to maintain on-chain registry of _blobs_, the binary data. 
The _blobs_ are referenced from smart contracts via their hashes. 

A _blob_ is a collection of named pieces of arbitrary binary data:
```
    <fieldName1> : <binaryChunk1>
    <fieldName2> : <binaryChunk2>
    ...
    <fieldNameN> : <binaryChunkN>
``` 

Here the `fieldNameK` is arbitrary binary array used as a name of the `binaryChunkK`. 
Usually `fieldNameK` is not long. Its interpretation is use-case specific.

The `binaryChunkK` may be of arbitrary size (practical limits apply, of course).

The order of field-chunk pairs is essential.

The hash of the _blob_ is equal to the hash of concatenation of all pieces:
```
    blobHash = hash( fieldName1 || binaryChunk1 || fieldName2 || binaryChunk2 || ... || fieldNameN || binaryChunkN)
``` 

There are two predefined field name which are interpreted by the VM while deploying smart contracts from binary:

- _fieldname_ = `"v"` is interpreted as _VM type_
- _fieldname_ = `"p"` is interpreted as _smart contract program binary_
  
If the field `"v"` is equal the string `"wasmtimevm"`, the binary chunk of `"p""` is interpreted as WebAssembly binary,
loadable to the _Wasmtime_ wasm interpreter.
    
Another use_case for the _blob_ may be full collection of self described immutable data of the smart contract program:
```
    "v" : VM type
    "p" : smart contract program binary
    "d" : data schema for data exchange between smart contract and outside sources and consumers
    "s" : program sources in .zip format
```
 
### Entry points

There's only on entry point, which allows to submit _blob_ to the smart contract:

* **storeBlob**. In the current implementation the data of the _blob_ is passed as parameters of the call 
to the entry point. It may be practically impossible to submit very large _blobs_ to the chain. In the future
we plan to implement special mechanism which allows for the nodes to download big data chunks as part of the 
committee consensus.     

### Views 

* **getBlobInfo** view returns information about fields of the blob with specific hash and sizes of data chunks:
```
    <fieldName1>: <size of the dataChunk1>
    ...
    <fieldNameN>: <size of the dataChunkN>
```
  
* **getBlobField** view allows to download data chunk of the field of particular _blob_

* **listBlobs** view returns list of pairs `blob hash`: `total size of chunks` for all blobs in the registry
 
 