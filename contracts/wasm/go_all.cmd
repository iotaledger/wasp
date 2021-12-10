@echo off
for /d %%f in (*.) do call go_build.cmd %%f %1
