wasp-cli init
wasp-cli request-funds
wasp-cli chain deploy --committee=0 --quorum=1 --chain=mychain --description="My chain"
wasp-cli chain deposit base:1000000
wasp-cli chain deploy-contract wasmtime testwasmlib "Test WasmLib" ../pkg/testwasmlib_bg.wasm
