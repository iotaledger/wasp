@echo off
cd ..
go install ../../tools/schema
schema -ts -build
cd scripts
