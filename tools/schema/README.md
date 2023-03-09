# Schema Tool

The Schema tool will generate the entire interface library and implementation skeleton 
for a smart contract from a schema definition file.

## Usage

### Setting up an initial smart contract folder

You can use the Schema tool to generate an initial folder with a schema definition file by
using `schema -init MyContractName` where `MyContractName` is a user-defined camel-case 
name for the smart contract.

The Schema tool will create a sub-folder `mycontractname` (all lower case) for the smart 
contract, with a `schema.yaml` in it that describes a minimal contract with the 
provided name, and can be used to generate a complete working contract later on.

### Working with the schema definition file

The initially generated `schema.yaml` contains all sections necessary to completely 
define the smart contract interface. You will modify this schema definition file to 
include any interface requirements for the smart contract and then use the Schema tool 
to (re-)generate the interface in the desired language(s).

You can check out the demo smart contracts in the contracts/wasm folder of the Wasp repo 
to see how to use the different sections, or check the (outdated) documentation on the 
[IOTA wiki](https://wiki.iota.org/shimmer/smart-contracts/guide/wasm_vm/schema/).

```yaml
name: MyContractName
description: MyContractName description
author: Eric Hop <eric@iota.org>
events: {}
structs: {}
typedefs: {}
state:
    owner: AgentID # current owner of this smart contract
funcs:
    init:
        params:
            owner: AgentID? # optional owner of this smart contract
    setOwner:
        access: owner # current owner of this smart contract
        params:
            owner: AgentID # new owner of this smart contract
views:
    getOwner:
        results:
            owner: AgentID # current owner of this smart contract
```

### Generating the smart contract interface

Once the `schema.yaml` is to your liking you can use the Schema tool to generate the 
interface code in the desired language(s). To do this you navigate into the folder 
with `schema.yaml` and run the Schema tool with one or more of the language flags:

- `schema -go` to generate the Go interface and implementation
- `schema -rs` to generate the Rust interface and implementation
- `schema -ts` to generate the Typescript interface and Assemblyscript implementation

You can provide multiple language flags if you want. For example `schema -go -rs -ts` 
will generate all three language interfaces and implementations.

Normally the Schema tool will only generate new code when the schema definition file 
or the Schema tool itself is newer than the last generated code. If for some reason you 
need to override this behavior you can use the `-force` flag to force code generation.

The Schema tool will generate a sub-folder for each separate language. These sub-folders 
will each have a several sub-folders:

- `mycontractname` This library/crate sub-folder contains the interface code that can be 
  used to 
  invoke the smart contract functions from within a smart contract, from within a test 
  environment, or from within a client application. It is completely under control of the 
  Schema tool and should not be modified.
- `mycontractnameimpl` This library/crate sub-folder contains the implementation code for 
  the smart contract functions. The generated code is specific to the functioning of the 
  smart contract. The only file in this sub-folder that you should modify is `funcs.xx`, 
  which contains the user-defined func and view implementations for the smart contract.
- `mycontractnamewasm` This crate sub-folder will only be generated for the Rust version. 
  The crate contains the Wasm stub that combines the interface code and implementation 
  code when building the Wasm binary file. The other languages achieve the same thing by 
  using a single stub file called `main.xx` that does not need its own sub-folder.

### Building the smart contract

The Schema tool will also build the smart contract for you. To be able to do that it 
requires that the proper compilers are already installed and can be reached through 
your PATH. These are the required compilers for each language:

- [Go](https://go.dev/) version 1.20 or higher with [tinygo](https://tinygo.org/) version
  0.27.0 or higher.
- [Rust](https://www.rust-lang.org/) version 1.67 or higher with
  [wasm-pack](https://github.com/rustwasm/wasm-pack) version 0.10.3 or higher.
- [Typescript](https://www.npmjs.com/package/typescript) version 4.9.5 or higher with 
  [Assemblyscript](https://www.assemblyscript.org/) version 0.27.1 or higher.

The Schema tool will standardize the naming and location of the generated Wasm code so 
that the `Solo` stand-alone test environment will automatically be able to find the Wasm 
code and deploy it for you in your tests.

To build the desired Wasm binary simply add the `-build` flag and the language flags for 
the desired languages to build. For example, `schema -go -rs -ts -build` will build 
the Wasm binaries for all languages.
