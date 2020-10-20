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
go test -buildmode=exe -run TestIncRepeatIncrement %1
pause
go test -buildmode=exe -run TestIncRepeatManyIncrement %1
pause
go test -buildmode=exe -run TestFrNothing %1
pause
go test -buildmode=exe -run TestFr5xNothing %1
pause
go test -buildmode=exe -run TestFrPlaceBet %1
pause
go test -buildmode=exe -run TestFrPlace5BetsAndPlay %1
pause
go test -buildmode=exe -run TestTrMintSupply %1
pause
go test -buildmode=exe -run TestDwfDeploy %1
pause
go test -buildmode=exe -run TestDwfDonateNTimes %1
pause
go test -buildmode=exe -run TestDwfDonateWithdrawAuthorised %1
pause
go test -buildmode=exe -run TestDwfDonateWithdrawNotAuthorised %1
pause
go test -buildmode=exe -run TestLoadTrAndFaAndThenRunTrMint %1
pause
go test -buildmode=exe -run TestTrMintAndFaAuctionWith2Bids %1
