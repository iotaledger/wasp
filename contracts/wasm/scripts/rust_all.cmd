@echo off
cd ..
go install ../../tools/schema
schema -rs
schema -rs -build
cd scripts
