@echo off
call core_build.cmd
call go_all.cmd -force
call ts_all.cmd -force
call rust_all.cmd -force
call update_hardcoded.cmd
