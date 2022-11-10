@echo off
if exist go.mod goto :root
if not exist schema_all.cmd goto :xit

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
