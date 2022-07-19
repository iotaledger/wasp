@echo off
for /d %%f in (*.) do call ts_build.cmd %%f %1
cd gascalibration
for /d %%f in (*.) do call ..\ts_build.cmd %%f %1
cd ..
