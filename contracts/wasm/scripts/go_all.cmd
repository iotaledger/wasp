@echo off
cd ..
go install ../../tools/schema
schema -go -build
cd scripts
