package iotajsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
)

type IotaObjectRef struct {
	/** Base64 string representing the object digest */
	Digest iotago.TransactionDigest `json:"digest"`
	/** Hex code as string representing the object id */
	ObjectID *iotago.ObjectID `json:"objectId"`
	/** Object version */
	Version iotago.SequenceNumber `json:"version"`
}

type IotaGasData struct {
	Payment []IotaObjectRef `json:"payment"`
	/** Gas Object's owner */
	Owner  string  `json:"owner"`
	Price  *BigInt `json:"price"`
	Budget *BigInt `json:"budget"`
}

type IotaParsedData struct {
	MoveObject *IotaParsedMoveObject `json:"moveObject,omitempty"`
	Package    *IotaMovePackage      `json:"package,omitempty"`
}

func (p IotaParsedData) Tag() string {
	return "dataType"
}

func (p IotaParsedData) Content() string {
	return ""
}

type IotaMovePackage struct {
	Disassembled map[string]interface{} `json:"disassembled"`
}

type IotaParsedMoveObject struct {
	Type              string          `json:"type"`
	HasPublicTransfer bool            `json:"hasPublicTransfer"`
	Fields            json.RawMessage `json:"fields"`
}

type IotaRawData struct {
	MoveObject *IotaRawMoveObject  `json:"moveObject,omitempty"`
	Package    *IotaRawMovePackage `json:"package,omitempty"`
}

func (r IotaRawData) Tag() string {
	return "dataType"
}

func (r IotaRawData) Content() string {
	return ""
}

type IotaRawMoveObject struct {
	Type              iotago.StructTag      `json:"type"`
	HasPublicTransfer bool                  `json:"hasPublicTransfer"`
	Version           iotago.SequenceNumber `json:"version"`
	BcsBytes          iotago.Base64Data     `json:"bcsBytes"`
}

type IotaRawMovePackage struct {
	Id              *iotago.ObjectID             `json:"id"`
	Version         iotago.SequenceNumber        `json:"version"`
	ModuleMap       map[string]iotago.Base64Data `json:"moduleMap"`
	TypeOriginTable []TypeOrigin                 `json:"typeOriginTable"`
	LinkageTable    map[iotago.ObjectID]UpgradeInfo
}

type UpgradeInfo struct {
	UpgradedId      iotago.ObjectID
	UpgradedVersion iotago.SequenceNumber
}

type TypeOrigin struct {
	ModuleName string          `json:"moduleName"`
	StructName string          `json:"structName"`
	Package    iotago.ObjectID `json:"package"`
}

type IotaObjectData struct {
	ObjectID *iotago.ObjectID     `json:"objectId"`
	Version  *BigInt              `json:"version"`
	Digest   *iotago.ObjectDigest `json:"digest"`
	/**
	 * Type of the object, default to be undefined unless IotaObjectDataOptions.showType is set to true
	 */
	Type *string `json:"type,omitempty"`
	/**
	 * Move object content or package content, default to be undefined unless IotaObjectDataOptions.showContent is set to true
	 */
	Content *serialization.TagJson[IotaParsedData] `json:"content,omitempty"`
	/**
	 * Move object content or package content in BCS bytes, default to be undefined unless IotaObjectDataOptions.showBcs is set to true
	 */
	Bcs *serialization.TagJson[IotaRawData] `json:"bcs,omitempty"`
	/**
	 * The owner of this object. Default to be undefined unless IotaObjectDataOptions.showOwner is set to true
	 */
	Owner *ObjectOwner `json:"owner,omitempty"`
	/**
	 * The digest of the transaction that created or last mutated this object.
	 * Default to be undefined unless IotaObjectDataOptions.showPreviousTransaction is set to true
	 */
	PreviousTransaction *iotago.TransactionDigest `json:"previousTransaction,omitempty"`
	/**
	 * The amount of IOTA we would rebate if this object gets deleted.
	 * This number is re-calculated each time the object is mutated based on
	 * the present storage gas price.
	 * Default to be undefined unless IotaObjectDataOptions.showStorageRebate is set to true
	 */
	StorageRebate *BigInt `json:"storageRebate,omitempty"`
	/**
	 * Display metadata for this object, default to be undefined unless IotaObjectDataOptions.showDisplay is set to true
	 * This can also be None if the struct type does not have Display defined
	 * See more details in https://forums.sui.io/t/nft-object-display-proposal/4872
	 */
	Display interface{} `json:"display,omitempty"`
}

func (data *IotaObjectData) Ref() iotago.ObjectRef {
	return iotago.ObjectRef{
		ObjectID: data.ObjectID,
		Version:  data.Version.Uint64(),
		Digest:   data.Digest,
	}
}

type IotaObjectDataOptions struct {
	/* Whether to fetch the object type, default to be false */
	ShowType bool `json:"showType,omitempty"`
	/* Whether to fetch the object content, default to be false */
	ShowContent bool `json:"showContent,omitempty"`
	/* Whether to fetch the object content in BCS bytes, default to be false */
	ShowBcs bool `json:"showBcs,omitempty"`
	/* Whether to fetch the object owner, default to be false */
	ShowOwner bool `json:"showOwner,omitempty"`
	/* Whether to fetch the previous transaction digest, default to be false */
	ShowPreviousTransaction bool `json:"showPreviousTransaction,omitempty"`
	/* Whether to fetch the storage rebate, default to be false */
	ShowStorageRebate bool `json:"showStorageRebate,omitempty"`
	/* Whether to fetch the display metadata, default to be false */
	ShowDisplay bool `json:"showDisplay,omitempty"`
}

type IotaObjectResponseError struct {
	NotExists *struct {
		ObjectID iotago.ObjectID `json:"object_id"`
	} `json:"notExists,omitempty"`
	Deleted *struct {
		ObjectID iotago.ObjectID       `json:"object_id"`
		Version  iotago.SequenceNumber `json:"version"`
		Digest   iotago.ObjectDigest   `json:"digest"`
	} `json:"deleted,omitempty"`
	UnKnown      *struct{} `json:"unKnown"`
	DisplayError *struct {
		Error string `json:"error"`
	} `json:"displayError"`
}

func (e IotaObjectResponseError) String() string {
	if e.NotExists != nil {
		return fmt.Sprintf("object not exists: %s", e.NotExists.ObjectID.String())
	}
	if e.Deleted != nil {
		return fmt.Sprintf("deleted obj{id=%s, version=%v, digest=%s}", e.Deleted.ObjectID.String(), e.Deleted.Version, e.Deleted.Digest.String())
	}
	if e.UnKnown != nil {
		return "unknown object"
	}
	if e.DisplayError != nil {
		return fmt.Sprintf("display err: %s", e.DisplayError.Error)
	}
	return ""
}

func (e IotaObjectResponseError) Tag() string {
	return "code"
}

func (e IotaObjectResponseError) Content() string {
	return ""
}

type IotaObjectResponse struct {
	Data  *IotaObjectData                                 `json:"data,omitempty"`
	Error *serialization.TagJson[IotaObjectResponseError] `json:"error,omitempty"`
}

func (r IotaObjectResponse) ResponseError() error {
	if r.Error != nil {
		return fmt.Errorf("%s", r.Error.Data.String())
	}
	return nil
}

type CheckpointSequenceNumber = uint64

type ObjectsPage = Page[IotaObjectResponse, iotago.ObjectID]

type IotaObjectDataFilter struct {
	MatchAll  []*IotaObjectDataFilter `json:"MatchAll,omitempty"`
	MatchAny  []*IotaObjectDataFilter `json:"MatchAny,omitempty"`
	MatchNone []*IotaObjectDataFilter `json:"MatchNone,omitempty"`
	// Query by type a specified Package.
	Package *iotago.ObjectID `json:"Package,omitempty"`
	// Query by type a specified Move module.
	MoveModule *MoveModule `json:"MoveModule,omitempty"`
	// Query by type
	StructType   *iotago.StructTag `json:"StructType,omitempty"`
	AddressOwner *iotago.Address   `json:"AddressOwner,omitempty"`
	ObjectOwner  *iotago.ObjectID  `json:"ObjectOwner,omitempty"`
	ObjectId     *iotago.ObjectID  `json:"ObjectId,omitempty"`
	// allow querying for multiple object ids
	ObjectIds []*iotago.ObjectID `json:"ObjectIds,omitempty"`
	Version   *BigInt            `json:"Version,omitempty"`
}

type IotaObjectResponseQuery struct {
	Filter  *IotaObjectDataFilter  `json:"filter,omitempty"`
	Options *IotaObjectDataOptions `json:"options,omitempty"`
}

type IotaPastObjectResponse = serialization.TagJson[IotaPastObject]

type IotaPastObject struct {
	// The object exists and is found with this version
	VersionFound *IotaObjectData `json:"VersionFound,omitempty"`
	// The object does not exist
	ObjectNotExists *iotago.ObjectID `json:"ObjectNotExists,omitempty"`
	// The object is found to be deleted with this version
	ObjectDeleted *IotaObjectRef `json:"ObjectDeleted,omitempty"`
	// The object exists but not found with this version
	VersionNotFound *VersionNotFoundData `json:"VersionNotFound,omitempty"`
	// The asked object version is higher than the latest
	VersionTooHigh *struct {
		ObjectID      iotago.ObjectID       `json:"object_id"`
		AskedVersion  iotago.SequenceNumber `json:"asked_version"`
		LatestVersion iotago.SequenceNumber `json:"latest_version"`
	} `json:"VersionTooHigh,omitempty"`
}

type VersionNotFoundData struct {
	ObjectID       *iotago.ObjectID
	SequenceNumber iotago.SequenceNumber
}

func (c *VersionNotFoundData) UnmarshalJSON(data []byte) error {
	var vals []any
	err := json.Unmarshal(data, &vals)
	if err != nil {
		return fmt.Errorf("failed to parse VersionNotFound content: %w", err)
	}
	if len(vals) != 2 {
		return fmt.Errorf("failed to parse VersionNotFound content: expected 2 elements, got %d", len(vals))
	}
	objIDHex, ok := vals[0].(string)
	if !ok {
		return fmt.Errorf("failed to parse VersionNotFound content: expected string, got %T", vals[0])
	}
	c.ObjectID, err = iotago.ObjectIDFromHex(objIDHex)
	seq, ok := vals[1].(float64)
	if !ok {
		return fmt.Errorf("failed to parse VersionNotFound content: expected number, got %T", vals[1])
	}
	c.SequenceNumber = uint64(seq)
	return nil
}

func (s IotaPastObject) Tag() string {
	return "status"
}

func (s IotaPastObject) Content() string {
	return "details"
}

type IotaGetPastObjectRequest struct {
	ObjectId *iotago.ObjectID `json:"objectId"`
	Version  *BigInt          `json:"version"`
}

type IotaNamePage = Page[string, iotago.ObjectID]
