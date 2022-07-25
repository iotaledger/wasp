@echo off
for /d %%f in (*.) do if exist %%f\pkg\%%f*_bg.wasm copy /y %%f\pkg\%%f*_bg.wasm %%f\test\*.*
cd gascalibration
for /d %%f in (*.) do if exist %%f\pkg\%%f*_bg.wasm copy /y %%f\pkg\%%f*_bg.wasm %%f\test\*.*
for /d %%f in (*.) do if exist %%f\go\pkg\%%f*_go.wasm copy /y %%f\go\pkg\%%f*_go.wasm %%f\test\*.*
for /d %%f in (*.) do if exist %%f\ts\pkg\%%f*_ts.wasm copy /y %%f\ts\pkg\%%f*_ts.wasm %%f\test\*.*
cd ..
if exist testcore\pkg\testcore_bg.wasm copy /y testcore\pkg\testcore_bg.wasm ..\..\packages\vm\core\testcore\sbtests\sbtestsc\*.*
if exist inccounter\pkg\inccounter_bg.wasm copy /y inccounter\pkg\inccounter_bg.wasm ..\..\tools\cluster\tests\wasm\*.*
cd ..\..\documentation\tutorial-examples
wasm-pack build
copy /y pkg\solotutorial_bg.wasm test
cd ..\..\contracts\wasm
