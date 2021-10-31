@echo off
for /d %%f in (*.) do call ts_build.cmd %%f %1
