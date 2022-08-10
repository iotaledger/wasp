@echo off
cd ..
for /d %%f in (*.) do call scripts\rust_build.cmd %%f %1
cd gascalibration
for /d %%f in (*.) do call ..\scripts\rust_build.cmd %%f %1
cd ..\scripts
