@echo off
if not exist ..\wasm md ..\wasm
for /d %%f in (*.) do if not "%%f"=="wasmlib" call build_go.bat %%f
