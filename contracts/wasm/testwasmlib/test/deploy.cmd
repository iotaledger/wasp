rem wasp-cli init
wasp-cli request-funds
wasp-cli chain deploy --committee=0 --quorum=1 --chain=mychain --description="My chain"
wasp-cli chain deposit base:500000000
wasp-cli balance
wasp-cli chain balance
wasp-cli chain deploy-contract wasmtime testwasmlib "Test WasmLib" ..\rs\testwasmlibwasm\pkg\testwasmlibwasm_bg.wasm
wasp-cli chain post-request -s testwasmlib random
wasp-cli chain call-view testwasmlib getRandom | wasp-cli decode string random uint64
wasp-cli chain balance
wasp-cli chain list-accounts
wasp-cli check-versions
