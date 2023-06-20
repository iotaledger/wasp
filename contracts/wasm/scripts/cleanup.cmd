@echo off
cd ..
go install ../../tools/schema
del /s /q cargo.lock
del /s /q target\*.*
schema -go -rs -ts -clean
cd scripts
