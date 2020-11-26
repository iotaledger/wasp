go test -buildmode=exe -run TestIncDeployment %1
pause
go test -buildmode=exe -run TestIncNothing %1
pause
go test -buildmode=exe -run TestInc5xNothing %1
pause
go test -buildmode=exe -run TestIncIncrement %1
pause
go test -buildmode=exe -run TestInc5xIncrement %1
pause
go test -buildmode=exe -run TestIncrementWithTransfer %1
pause
go test -buildmode=exe -run TestIncPostIncrement %1
pause
go test -buildmode=exe -run TestIncRepeatManyIncrement %1
pause
go test -buildmode=exe -run TestIncLocalStateInternalCall %1
pause
go test -buildmode=exe -run TestIncLocalStateSandboxCall %1
pause
go test -buildmode=exe -run TestIncLocalStatePost %1
