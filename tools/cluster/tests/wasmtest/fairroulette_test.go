package wasmtest

import "testing"

const fr_wasmPath = "fairroulette_bg.wasm"
const fr_description = "Fair roulette, a PoC smart contract"

func TestFrNothing(t *testing.T) {
	testNothing(t, "TestFrNothing", fr_wasmPath, fr_description, 1)
}

func Test5xFrNothing(t *testing.T) {
	testNothing(t, "Test5xFrNothing", fr_wasmPath, fr_description, 5)
}
