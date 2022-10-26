@echo off
if "%1"=="" call cleanup_all.cmd
if "%1"=="" goto :xit2
cd %1
if not exist schema.yaml goto :xit
schema -go -rust -ts -clean
if exist ts\%1\tsconfig.json del ts\%1\tsconfig.json
del /s cargo.*
del /s license
:xit
cd ..
:xit2
