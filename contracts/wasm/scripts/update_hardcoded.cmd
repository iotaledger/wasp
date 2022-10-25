@echo off
if exist ..\testcore\rs\main\pkg\main_bg.wasm copy /y ..\testcore\rs\main\pkg\main_bg.wasm ..\..\..\packages\vm\core\testcore\sbtests\sbtestsc\testcore_bg.*
if exist ..\inccounter\rs\main\pkg\main_bg.wasm copy /y ..\inccounter\rs\main\pkg\main_bg.wasm ..\..\..\tools\cluster\tests\wasm\inccounter.*
cd ..\..\..\documentation\tutorial-examples
wasm-pack build
copy /y pkg\solotutorial_bg.wasm test
cd ..\..\contracts\wasm\scripts
