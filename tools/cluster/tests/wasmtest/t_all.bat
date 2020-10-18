go test -buildmode=exe -run TestDeployment %1
pause
go test -buildmode=exe -run TestIncNothing %1
pause
go test -buildmode=exe -run Test5xIncNothing %1
pause
go test -buildmode=exe -run TestIncrement %1
pause
go test -buildmode=exe -run Test5xIncrement %1
pause
go test -buildmode=exe -run TestRepeatIncrement %1
pause
go test -buildmode=exe -run TestRepeatManyIncrement %1
pause
go test -buildmode=exe -run TestFrNothing %1
pause
go test -buildmode=exe -run Test5xFrNothing %1
pause
go test -buildmode=exe -run TestPlaceBet %1
pause
go test -buildmode=exe -run TestPlace5BetsAndPlay %1
pause
go test -buildmode=exe -run TestMintSupply %1
