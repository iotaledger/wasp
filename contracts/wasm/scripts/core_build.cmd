@echo off

go install ../../../tools/schema

cd ..\..\..\packages\wasmvm\wasmlib
schema -go -rs -ts -force

cd ..\..\..\contracts\wasm
call npm install
xcopy /s /q ..\..\packages\wasmvm\wasmlib\as\wasmlib\*.* node_modules\wasmlib\*.*
xcopy /s /q ..\..\packages\wasmvm\wasmvmhost\ts\wasmvmhost\*.* node_modules\wasmvmhost\*.*
cd scripts
