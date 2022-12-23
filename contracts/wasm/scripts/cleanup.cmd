@echo off
if "%1"=="" call cleanup_all.cmd
if "%1"=="" goto :xit2
cd %1
if not exist schema.yaml goto :xit
schema -go -rs -ts -clean
if exist ts\%1impl\tsconfig.json del ts\%1impl\tsconfig.json
if exist rs\%1impl\Cargo.lock del rs\%1impl\Cargo.lock
if exist rs\%1impl\Cargo.toml del rs\%1impl\Cargo.toml
if exist rs\%1impl\README.md del rs\%1impl\README.md
if exist rs\%1impl\LICENSE del rs\%1impl\LICENSE
:xit
cd ..
:xit2
