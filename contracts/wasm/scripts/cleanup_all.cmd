rem @echo off
cd ..
for /d %%f in (*.) do call scripts\cleanup.cmd %%f
cd gascalibration
for /d %%f in (*.) do call ..\scripts\cleanup.cmd %%f
cd ..\scripts
