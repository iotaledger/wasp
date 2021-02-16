@echo off
cd %1
if not exist src\lib.rs goto :xit
echo compiling %1_bg.wasm
wasm-pack build
copy /y pkg\%1_bg.wasm test
:xit
cd ..
