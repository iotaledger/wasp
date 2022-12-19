package model

import "github.com/iotaledger/wasp/packages/hashing"

type BlobData struct {
	Data Bytes `swagger:"desc(Blob content (base64))"`
}

func NewBlobData(data []byte) *BlobData {
	return &BlobData{Data: NewBytes(data)}
}

type BlobInfo struct {
	Exists bool      `json:"exists" swagger:"desc(Whether or not the blob exists in the registry)"`
	Hash   HashValue `json:"hash" swagger:"desc(Hash of the blob)"`
}

func NewBlobInfo(exists bool, hash hashing.HashValue) *BlobInfo {
	return &BlobInfo{Exists: exists, Hash: NewHashValue(hash)}
}
