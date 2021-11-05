@echo off
for /d %%f in (*.) do call rust_build.cmd %%f %1

