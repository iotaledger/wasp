rem go test -buildmode=exe -run %1
go test -v -tags noevm -buildmode=exe -cpu 1 -count 10 -run %1 -failfast