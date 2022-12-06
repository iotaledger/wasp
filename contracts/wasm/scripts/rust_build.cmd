@echo off
cd %1
if not exist schema.yaml goto :xit
echo Building %1
schema -rs %2
echo Compiling %1wasm_bg.wasm
wasm-pack build rs\%1wasm
:xit
cd ..
