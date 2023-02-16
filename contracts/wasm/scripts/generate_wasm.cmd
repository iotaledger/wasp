@echo off
if exist go.mod goto :root
if not exist schema_all.cmd goto :xit

set BUILD_TAGS=rocksdb
for /f %%f in ('git describe --tags') do set BUILD_LD_FLAGS=-X=github.com/iotaledger/wasp/core/app.Version=%%f
go install -ldflags %BUILD_LD_FLAGS% ../../../tools/schema

call schema_all.cmd
cd ..
golangci-lint run --fix
cd scripts
goto :xit

:root
cd contracts\wasm\scripts
call generate_wasm.cmd
cd ..\..\..

:xit
