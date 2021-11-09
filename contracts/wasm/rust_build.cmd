@echo off
cd %1
if not exist schema.yaml if not exist schema.json goto :xit
echo Building %1
schema -rust %2
echo compiling %1_bg.wasm
wasm-pack build
:xit
cd ..
