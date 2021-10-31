@echo off
cd %1
if not exist go\main.go goto :xit
echo Building %1
schema -go %2
echo compiling %1_go.wasm
if not exist go\pkg mkdir go\pkg
tinygo build -o go/pkg/%1_go.wasm -target wasm go/main.go
if exist go/pkg/%1_go.wasm wasm2wat go/pkg/%1_go.wasm >go\pkg\%1_go.wat
if exist go/pkg/%1_go.wasm wasm-decompile go/pkg/%1_go.wasm >go\pkg\%1_go.txt
:xit
cd ..
