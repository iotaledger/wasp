go test -buildmode=exe -run TestBlobDeployChain %1
pause
go test -buildmode=exe -run TestBlobStoreSmallBlob %1
pause
go test -buildmode=exe -run TestBlobStoreManyBlobs %1
