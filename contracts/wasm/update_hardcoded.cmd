@echo off
for /d %%f in (*.) do if exist %%f\pkg\%%f*_bg.wasm copy /y %%f\pkg\%%f*_bg.wasm %%f\test\*.*
if exist testcore\pkg\testcore_bg.wasm copy /y testcore\pkg\testcore_bg.wasm ..\..\packages\vm\core\testcore_stardust\sbtests\sbtestsc\*.*
if exist inccounter\pkg\inccounter_bg.wasm copy /y inccounter\pkg\inccounter_bg.wasm ..\..\tools\cluster\tests\wasm\*.*
cd ..\..\documentation\tutorial-examples
wasm-pack build
copy /y pkg\example_tutorial_bg.wasm test
cd ..\..\contracts\wasm
