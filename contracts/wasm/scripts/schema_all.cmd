@echo off
cd ..
go install ../../tools/schema
schema -go -rs -ts
golangci-lint run --fix
cd ..\scripts
