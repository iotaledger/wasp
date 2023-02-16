@echo off

set BUILD_TAGS=rocksdb
for /f %%f in ('git describe --tags') do set BUILD_LD_FLAGS=-X=github.com/iotaledger/wasp/core/app.Version=%%f
go install -ldflags %BUILD_LD_FLAGS% ../../../tools/schema

cd ..\..\..\packages\wasmvm\wasmlib
schema -core -go -rs -ts -force

cd ..\..\..\contracts\wasm
del /s /q d:\work\node_modules\wasmlib\*.* >nul:
xcopy /s /q ..\..\packages\wasmvm\wasmlib\as\wasmlib d:\work\node_modules\wasmlib
del /s /q d:\work\node_modules\wasmvmhost\*.* >nul:
xcopy /s /q ..\..\packages\wasmvm\wasmvmhost\ts\wasmvmhost d:\work\node_modules\wasmvmhost
cd scripts
