@echo off
cd %1
if not exist schema.yaml goto :xit
echo Generating %1
schema -go -rs -ts %2
if exist ..\..\..\..\debug.txt call ..\scripts\debug.cmd
if exist ..\..\..\..\..\debug.txt call ..\..\scripts\debug.cmd
:xit
cd ..
