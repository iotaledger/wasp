@echo off
cd %1
if not exist schema.yaml goto :xit
echo Building %1
schema -go %2
echo compiling %1_go.wasm
if not exist go\pkg mkdir go\pkg
tinygo build -o go/pkg/%1_go.wasm -target wasm -gc=leaking -opt 2 -no-debug go/main.go
:xit
cd ..
