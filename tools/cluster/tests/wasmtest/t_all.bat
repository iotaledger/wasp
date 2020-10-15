go test -buildmode=exe -run TestDeployment
pause
go test -buildmode=exe -run TestIncNothing
pause
go test -buildmode=exe -run Test5xIncNothing
pause
go test -buildmode=exe -run TestIncrement
pause
go test -buildmode=exe -run Test5xIncrement
pause
go test -buildmode=exe -run TestRepeatIncrement
pause
go test -buildmode=exe -run TestRepeatManyIncrement
pause
go test -buildmode=exe -run TestFrNothing
pause
go test -buildmode=exe -run Test5xFrNothing
pause
go test -buildmode=exe -run TestPlaceBet
pause
go test -buildmode=exe -run TestPlace5BetsAndPlay
