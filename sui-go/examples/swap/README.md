# swap

An example shows how to implement an extremely basic defi swap. 
The swap contract refers the [official example](https://github.com/MystenLabs/sui/blob/main/sui_programmability/examples/defi/sources/pool.move)

Sui toolchain is necessary. 

## How to Run

1. run a sui test validator
```bash
$ sui-test-validator
```
1. run go main program in `swap-go/`
```bash
$ go run main.go
```
The swap contract will be automatically built
