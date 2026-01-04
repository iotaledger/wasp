package graphqltypes

import "github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

type ObjectChangeType string

const (
	ObjectChangeCreated     ObjectChangeType = "created"
	ObjectChangeDeleted     ObjectChangeType = "deleted"
	ObjectChangeModified    ObjectChangeType = "modified"
	ObjectChangePublished   ObjectChangeType = "published"
	ObjectChangeTransferred ObjectChangeType = "transferred"
	ObjectChangeWrapped     ObjectChangeType = "wrapped"
)

type ObjectChange struct {
	ObjectID    iotago.ObjectID
	Type        ObjectChangeType
	Created     *CreatedChange
	Deleted     *DeletedChange
	Modified    *ModifiedChange
	Published   *PublishedChange
	Transferred *TransferredChange
	Wrapped     *WrappedChange
}

type CreatedChange struct {
	ObjectType string
	Version    uint64
	Digest     iotago.ObjectDigest
	Owner      *Owner
}

type DeletedChange struct {
	Version uint64
	Digest  iotago.ObjectDigest
}

type ModifiedChange struct {
	ObjectType      string
	PreviousVersion uint64
	Version         uint64
	Digest          iotago.ObjectDigest
	Owner           *Owner
}

type PublishedChange struct {
	PackageID iotago.PackageID
	Version   uint64
	Digest    iotago.ObjectDigest
	Modules   []string
}

type TransferredChange struct {
	ObjectType string
	Version    uint64
	Digest     iotago.ObjectDigest
	Owner      *Owner
}

type WrappedChange struct {
	ObjectType string
	Version    uint64
	Digest     iotago.ObjectDigest
}
