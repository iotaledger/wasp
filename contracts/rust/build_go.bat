@echo off
cd %1
if not exist wasmmain\main.go goto :xit
echo Building %1
schema -go %2
echo compiling %1_go.wasm
if not exist wasmmain\pkg md wasmmain\pkg
tinygo build -o wasmmain/pkg/%1_go.wasm -target wasm wasmmain/main.go
:xit
cd ..
