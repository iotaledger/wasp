@echo off
cd %1
if not exist schema.yaml goto :xit
echo Generating %1
schema -go -rust -ts %2
:xit
cd ..
