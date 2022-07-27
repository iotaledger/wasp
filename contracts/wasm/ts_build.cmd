@echo off
cd %1
if not exist schema.yaml goto :xit
echo Building %1
schema -ts %2
echo compiling %1_ts.wasm
call npx asc ts/%1/lib.ts --lib d:/work/node_modules -O --outFile ts/pkg/%1_ts.wasm
rem call npx asc ts/%1/lib.ts --lib d:/work/node_modules -O --outFile ts/pkg/%1_ts.wasm --textFile ts/pkg/%1_ts.wat
:xit
cd ..
