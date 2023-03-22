@echo off
call core_build.cmd
cd ..
schema -go -rs -ts -force
schema -go -rs -ts -build
golangci-lint run --fix
cd scripts
call update_hardcoded.cmd
