@echo off

cd ..\..\..\packages\wasmvm\wasmlib\ts\wasmlib
call npm i
call npm link

cd ..\..\..\..\..\contracts\wasm\testwasmlib\ts\testwasmlib
call npm link
call npm link wasmlib

cd ..\..\..\..\..\packages\wasmvm\wasmclient\ts\wasmclient
call npm i
call npm link wasmlib testwasmlib

cd ..\..\..\..\..\contracts\wasm\scripts
