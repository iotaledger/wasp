@echo off
if exist ..\testcore\rs\testcore_main\pkg\testcore_main_bg.wasm copy /y ..\testcore\rs\testcore_main\pkg\testcore_main_bg.wasm ..\..\..\packages\vm\core\testcore\sbtests\sbtestsc\testcore_bg.*
if exist ..\inccounter\rs\inccounter_main\pkg\inccounter_main_bg.wasm copy /y ..\inccounter\rs\inccounter_main\pkg\inccounter_main_bg.wasm ..\..\..\tools\cluster\tests\wasm\inccounter_bg.*
cd ..\..\..\documentation\tutorial-examples
wasm-pack build
copy /y pkg\solotutorial_bg.wasm test
cd ..\..\contracts\wasm\scripts
