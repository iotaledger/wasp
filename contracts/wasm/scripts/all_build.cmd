@echo off
call core_build.cmd
call schema_all.cmd -force
call go_all.cmd
call ts_all.cmd
call rust_all.cmd
call update_hardcoded.cmd
