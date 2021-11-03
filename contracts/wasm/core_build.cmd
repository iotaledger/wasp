@echo off
cd ..\..\packages\vm\wasmlib
schema -core -go -rust -ts -force
del /s /q d:\work\node_modules\wasmlib\*.* >nul:
xcopy /s /q d:\Work\go\github.com\iotaledger\wasp\packages\vm\wasmlib\ts\wasmlib d:\work\node_modules\wasmlib
cd ..\..\..\contracts\wasm
