go test -buildmode=exe -run TestDeployChain %1
pause
go test -buildmode=exe -run TestDeployContractOnly %1
pause
go test -buildmode=exe -run TestDeployContractAndSpawn %1
pause
go test -buildmode=exe -run TestIncDeployment %1
pause
go test -buildmode=exe -run TestIncNothing %1
pause
go test -buildmode=exe -run TestInc5xNothing %1
pause
go test -buildmode=exe -run TestIncIncrement %1
pause
go test -buildmode=exe -run TestInc5xIncrement %1

