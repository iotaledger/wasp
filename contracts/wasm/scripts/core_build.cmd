@echo off
go install ../../../tools/schema
cd ..\..\..\packages\wasmvm\wasmlib
schema -core -go -rust -ts -force
cd ..\..\..\contracts\wasm
del /s /q d:\work\node_modules\wasmlib\*.* >nul:
xcopy /s /q ..\..\packages\wasmvm\wasmlib\ts\wasmlib d:\work\node_modules\wasmlib
cd scripts
