# EVM Deposit Frontend

A minimalistic frontend PoC which allows the user to fill up a specific EVM address on Wasp L2 chains with funds.

Funds will be requested from a faucet and be sent to the selected address.

> Note: Currently only nodes with an active Proof of Work component are supported. Once iota.js contains a faster Pow function, all nodes are supported.

## Installation
```
npm install
```

## Development

This starts a development server:

```
npm run dev
```

## Building

```
npm run build
```

## Validation / Linting

```
npm run check
tsc --noEmit
```