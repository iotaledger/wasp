@echo off
cd %1
if not exist schema.yaml goto :xit
echo Building %1
schema -rust %2
echo Compiling %1_bg.wasm
wasm-pack build
:xit
cd ..
