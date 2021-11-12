@echo off
cd %1
if not exist schema.yaml if not exist schema.json goto :xit
echo Building %1
schema -ts %2
echo compiling %1_ts.wasm
call asc ts/%1/lib.ts --lib d:/work/node_modules --binaryFile ts/pkg/%1_ts.wasm
rem call asc ts/%1/lib.ts --lib d:/work/node_modules --binaryFile ts/pkg/%1_ts.wasm --textFile ts/pkg/%1_ts.wat
:xit
cd ..
