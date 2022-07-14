---
description: The errors contract keeps a map of error codes to error messages templates. These error codes are used in request receipts.
image: /img/logo/WASP_logo_dark.png
keywords:
- smart contracts
- core
- root
- initialization
- entry points
- fees
- ownership
- views
- reference
--- 
# The `errors` Contract

The `errors` contract is one of the [core contracts](overview.md) on each IOTA Smart Contracts
chain.

The `errors` contract keeps a map of error codes to error messages templates. These error codes are used in request receipts.

---

## Entry Points

### `registerError(m ErrorMessageFormat) c ErrorCode`

Registers an error message template. These templates support standard [go verbs](https://pkg.go.dev/fmt#hdr-Printing) for variable printing.

The error code returned can then be used in contract panicsi, this is a way to save gas for the users of a given contract, as lengthy error strings can be stored only once and be reused just by providing the error code.

---

## Views

### `getErrorMessageFormat(c ErrorCode) m ErrorMessageFormat`

Returns the message template stored for a given error code.
