@echo off
go install ../../tools/schema
cd ..\..\packages\wasmvm\wasmlib
schema -core -go -rust -ts -force
del /s /q d:\work\node_modules\wasmlib\*.* >nul:
del /s /q d:\work\node_modules\wasmclient\*.* >nul:
xcopy /s /q d:\Work\go\github.com\iotaledger\wasp\packages\wasmvm\wasmlib\ts\wasmlib d:\work\node_modules\wasmlib
rem xcopy /s /q d:\Work\go\github.com\iotaledger\wasp\packages\wasmvm\wasmclient\ts\wasmclient d:\work\node_modules\wasmclient
cd ..\..\..\contracts\wasm
