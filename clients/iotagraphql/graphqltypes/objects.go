package graphqltypes

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type ObjectData struct {
	ObjectID *iotago.ObjectID
	Version  uint64
	Digest   *iotago.ObjectDigest

	Type                *string
	Content             json.RawMessage
	BcsBytes            []byte
	Owner               *Owner
	PreviousTransaction *iotago.TransactionDigest
	StorageRebate       *uint64
	Display             interface{}
}

func (data *ObjectData) Ref() iotago.ObjectRef {
	return iotago.ObjectRef{
		ObjectID: data.ObjectID,
		Version:  data.Version,
		Digest:   data.Digest,
	}
}

type ObjectResponseError struct {
	NotExists    *iotago.ObjectID
	Deleted      *DeletedObjectInfo
	DisplayError *string
}

type DeletedObjectInfo struct {
	ObjectID iotago.ObjectID
	Version  uint64
	Digest   iotago.ObjectDigest
}

func (e ObjectResponseError) String() string {
	if e.NotExists != nil {
		return fmt.Sprintf("object not exists: %s", e.NotExists.String())
	}
	if e.Deleted != nil {
		return fmt.Sprintf("deleted obj{id=%s, version=%v, digest=%s}",
			e.Deleted.ObjectID.String(), e.Deleted.Version, e.Deleted.Digest.String())
	}
	if e.DisplayError != nil {
		return fmt.Sprintf("display err: %s", *e.DisplayError)
	}
	return "unknown error"
}

type PastObjectStatus string

const (
	PastObjectVersionFound    PastObjectStatus = "version_found"
	PastObjectNotExists       PastObjectStatus = "not_exists"
	PastObjectDeleted         PastObjectStatus = "deleted"
	PastObjectVersionNotFound PastObjectStatus = "version_not_found"
	PastObjectVersionTooHigh  PastObjectStatus = "version_too_high"
)

type PastObject struct {
	Status PastObjectStatus

	VersionFound    *ObjectData
	ObjectNotExists *iotago.ObjectID
	ObjectDeleted   *DeletedObjectInfo
	VersionNotFound *VersionNotFoundInfo
	VersionTooHigh  *VersionTooHighInfo
}

type VersionNotFoundInfo struct {
	ObjectID *iotago.ObjectID
	Version  uint64
}

type VersionTooHighInfo struct {
	ObjectID      iotago.ObjectID
	AskedVersion  uint64
	LatestVersion uint64
}
