@echo off
for /d %%f in (*.) do if exist %%f\pkg\%%f*.wasm copy /y %%f\pkg\%%f*.wasm %%f\test\*.*
if exist testcore\pkg\testcore_bg.wasm copy /y testcore\pkg\testcore_bg.wasm ..\..\packages\vm\core\testcore\sandbox_tests\test_sandbox_sc\*.*
if exist inccounter\pkg\inccounter_bg.wasm copy /y inccounter\pkg\inccounter_bg.wasm ..\..\tools\cluster\tests\wasm\*.*

