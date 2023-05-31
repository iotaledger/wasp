@echo off
cd ..
go install ../../tools/schema
del /s /q cargo.lock
schema -go -rs -ts -clean
cd scripts
