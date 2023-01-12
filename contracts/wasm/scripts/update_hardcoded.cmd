@echo off
if exist ..\testcore\rs\testcorewasm\pkg\testcorewasm_bg.wasm copy /y ..\testcore\rs\testcorewasm\pkg\testcorewasm_bg.wasm ..\..\..\packages\vm\core\testcore\sbtests\sbtestsc\testcore_bg.*
if exist ..\inccounter\rs\inccounterwasm\pkg\inccounterwasm_bg.wasm copy /y ..\inccounter\rs\inccounterwasm\pkg\inccounterwasm_bg.wasm ..\..\..\tools\cluster\tests\wasm\inccounter_bg.*
cd ..\..\..\documentation\tutorial-examples
schema -go -rs
wasm-pack build rs\solotutorialwasm
copy /y rs\solotutorialwasm\pkg\solotutorialwasm_bg.wasm test\solotutorial_bg.wasm
cd ..\..\contracts\wasm\scripts
