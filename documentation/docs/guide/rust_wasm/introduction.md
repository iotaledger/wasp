---
keywords:
- ISCP
- Smart Contracts
- Rust
- Wasm
description: Rust/Wasm based smart contracts
image: /img/logo/WASP_logo_dark.png
---

# Rust/Wasm Based Smart Contracts




<!-- 
// TODO 

The `example1` program has three entry points:

- `storeString` a full entry point. It first checks if parameter
  called `paramString` exist. If so, it stores the string value of the parameter
  into the state variable `storedString`. If parameter `paramString` is missing,
  the program panics.

- `getString` is a view entry point that returns the value of the
  variable `storedString`.

- `withdrawIota` is a full entry point that checks if the caller is and address
  and if the caller is equal to the creator of smart contract. If not, it
  panics. If it passes the validation, the program sends all the iotas contained
  in the smart contract's account to the caller.

Note that in the `example1` the Rust functions associated with full entry points
take a parameter of type `ScFuncContext`. It gives full (read-write) access to
the state. In contrast, `getString` is a view entry point and its associated
function parameter has type `ScViewContext`. A view is not allowed to mutate 
the state. -->