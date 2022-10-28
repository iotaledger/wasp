@echo off
if "%1"=="" call cleanup_all.cmd
if "%1"=="" goto :xit2
cd %1
if not exist schema.yaml goto :xit
schema -go -rust -ts -clean
if exist ts\%1\tsconfig.json del ts\%1\tsconfig.json
if exist rs\%1\Cargo.lock del rs\%1\Cargo.lock
if exist rs\%1\Cargo.toml del rs\%1\Cargo.toml
if exist rs\%1\README.md del rs\%1\README.md
if exist rs\%1\LICENSE del rs\%1\LICENSE
:xit
cd ..
:xit2
