wasp-cli request-funds
timeout 5
wasp-cli chain deploy --committee=0 --quorum=1 --chain=mychain --description="My chain"
timeout 2
wasp-cli chain deposit :10000
timeout 2
wasp-cli chain deploy-contract wasmtime testwasmlib "Test WasmLib" testwasmlib_bg.wasm
