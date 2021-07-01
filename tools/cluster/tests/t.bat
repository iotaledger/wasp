rem go test -buildmode=exe -run %1
go test -tags rocksdb -buildmode=exe -count 10 -run TestIncRepeatManyIncrement