@echo off
cd ..
go install ../../tools/schema
schema -go -rs -ts -clean
cd scripts
