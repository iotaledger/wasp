@echo off
cd ..\..\packages\vm\wasmlib
schema -core -go -rust -ts -force
cd ..\..\..\contracts\wasm
