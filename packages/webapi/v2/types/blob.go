package types

import "github.com/iotaledger/wasp/packages/hashing"

type BlobData struct {
	Data Base64 `swagger:"desc(Blob content (base64))"`
}

func NewBlobData(data []byte) *BlobData {
	return &BlobData{Data: NewBase64(data)}
}

type BlobInfo struct {
	Exists bool      `swagger:"desc(Whether or not the blob exists in the registry)"`
	Hash   HashValue `swagger:"desc(Hash of the blob)"`
}

func NewBlobInfo(exists bool, hash hashing.HashValue) *BlobInfo {
	return &BlobInfo{Exists: exists, Hash: NewHashValue(hash)}
}
