# Schema Tool

The Schema tool will (re-)generate the entire interface library and implementation 
skeleton for a smart contract from a schema definition file. The idea behind the Schema 
tool is that the programmer will be relieved from as many burdens as possible. This means 
that the Schema tool will generate code a lot of boilerplate code within a standardized 
folder structure. It will also provide a strongly typed interface to the smart contract 
function parameters, return values, and to the smart contract's state storage.

The code generator of the Schema tool has been thoroughly tested, which means that you can
be sure that the generated code is bug-free, and you can focus on the implementation of
the smart contract without distractions. Any time changes are made to the smart contract 
interface by the programmer in the schema definition file, the Schema tool will 
automatically re-generate any affected code to reflect these changes.

## Usage

```text
Usage of schema:
  -init string
        generate new folder with schema file for smart contract named <string>
  -go
        generate Go code
  -rs
        generate Rust code
  -ts
        generate TypScript code
  -force
        force code generation
  -build
        build wasm target for specified languages
  -clean
        clean up files that can be re-generated for specified languages
  -version
        show schema tool version
```

### Setting up an initial smart contract folder

You can use the Schema tool to generate an initial folder with a default schema 
definition file by using `schema -init MyContractName` where `MyContractName` is a 
user-defined camel-case name for the smart contract.

The Schema tool will create a sub-folder `mycontractname` (all lower case) for the smart 
contract, with a default `schema.yaml` in it that describes a minimal contract with the 
provided name, and can be used to generate a complete working contract later on.

The initial `schema.yaml` contains all sections necessary to completely define the smart 
contract interface. You will modify this schema definition file to include additional 
interface requirements for your smart contract and then use the Schema tool to 
(re-)generate the interface in the desired language(s).

You can check out our demo smart contracts in the contracts/wasm folder of the
[Wasp repository](https://github.com/iotaledger/wasp) to see how to use the different 
sections, or check the (somewhat outdated, coming soon) documentation on the 
[IOTA wiki](https://wiki.iota.org/shimmer/smart-contracts/guide/wasm_vm/schema/).

The initial `schema.yaml` looks like this:

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

Once `schema.yaml` has been modified to your liking you can use the Schema tool to
(re-)generate the interface code in the desired language(s). To do this you navigate into 
the folder that contains `schema.yaml` and run the Schema tool with one or more of the 
language flags:

- `schema -go` to generate the Go interface and implementation
- `schema -rs` to generate the Rust interface and implementation
- `schema -ts` to generate the Typescript interface and implementation

You can provide multiple language flags if you want. For example `schema -go -rs -ts` 
will generate all three language interfaces and implementations.

Normally the Schema tool will only generate new code when the schema definition file 
or the Schema tool itself is newer than the last generated code. If for some reason you 
need to override this behavior you can use the `-force` flag to force code generation.

The Schema tool will generate a sub-folder for each separate language. These sub-folders 
will each have a several sub-folders:

- `mycontractname` This library/crate sub-folder contains the interface code that can be 
  used to invoke the smart contract functions from within a smart contract, from within a 
  test environment, or from within a client application. This code is completely under 
  control of the Schema tool and should never be modified by the user.
- `mycontractnameimpl` This library/crate sub-folder contains the implementation code for 
  the smart contract functions. The generated code is specific to the functioning of the 
  smart contract. The only file in this sub-folder that you should modify is `funcs.xx`, 
  which contains the user-defined function implementations for the smart contract.
- `mycontractnamewasm` This crate sub-folder will only be generated for Rust code. The 
  crate contains the Wasm stub that combines the interface code and implementation code 
  when building the Wasm binary file. The other languages achieve the same thing by using 
  a single stub source file called `main.xx`, which does not need its own sub-folder.

### Building the smart contract

The Schema tool will also build the smart contract for you. To be able to do that it 
requires that the proper compilers have already been installed and can be reached through 
your execution PATH. These are the required compilers for each language:

- [Go](https://go.dev/) version 1.20 or higher with [tinygo](https://tinygo.org/) version
  0.27 or higher.
- [Rust](https://www.rust-lang.org/) version 1.67 or higher with
  [wasm-pack](https://github.com/rustwasm/wasm-pack) version 0.10.3 or higher.
- [Typescript](https://www.npmjs.com/package/typescript) version 4.9.5 or higher with 
  [Assemblyscript](https://www.assemblyscript.org/) version 0.27.1 or higher.

The Schema tool will standardize the naming and location of the generated Wasm code so 
that the `Solo` stand-alone test environment will automatically be able to find the Wasm 
code and deploy it for you in your tests.

To build the desired Wasm binaries simply add the `-build` flag to the desired language 
flags. For example, `schema -go -rs -ts -build` will build the Wasm binaries for all 
supported languages.

### Cleaning up after yourself

The Schema tool will normally generate quite a lot of code. Sometimes you will want to 
be able to clean up all generated artifacts. To that end the Schema tool provides the 
`-clean` flag. Add this flag to the desired language flags to clean up the artifacts 
for the specified languages. 

You may want to make sure that the artifact files do not end up in your repository, 
since they can be re-generated at any time. There is a `.gitignore` file in the 
`contracts/wasm` folder of the Wasp repository that shows how to achieve this.
