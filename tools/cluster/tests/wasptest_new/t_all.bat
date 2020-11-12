go test -buildmode=exe -run TestDeployChain %1
pause
go test -buildmode=exe -run TestDeployContractOnly %1
pause
go test -buildmode=exe -run TestDeployContractAndSpawn %1
pause
go test -buildmode=exe -run TestDeployExternalContractOnly %1
pause
go test -buildmode=exe -run TestIncNothing %1
pause
go test -buildmode=exe -run TestInc5xNothing %1

