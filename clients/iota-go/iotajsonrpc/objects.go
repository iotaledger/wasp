package iotajsonrpc

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
)

type SuiObjectRef struct {
	/** Base64 string representing the object digest */
	Digest iotago.TransactionDigest `json:"digest"`
	/** Hex code as string representing the object id */
	ObjectID *iotago.ObjectID `json:"objectId"`
	/** Object version */
	Version iotago.SequenceNumber `json:"version"`
}

type SuiGasData struct {
	Payment []SuiObjectRef `json:"payment"`
	/** Gas Object's owner */
	Owner  string  `json:"owner"`
	Price  *BigInt `json:"price"`
	Budget *BigInt `json:"budget"`
}

type SuiParsedData struct {
	MoveObject *SuiParsedMoveObject `json:"moveObject,omitempty"`
	Package    *SuiMovePackage      `json:"package,omitempty"`
}

func (p SuiParsedData) Tag() string {
	return "dataType"
}

func (p SuiParsedData) Content() string {
	return ""
}

type SuiMovePackage struct {
	Disassembled map[string]interface{} `json:"disassembled"`
}

type SuiParsedMoveObject struct {
	Type              string          `json:"type"`
	HasPublicTransfer bool            `json:"hasPublicTransfer"`
	Fields            json.RawMessage `json:"fields"`
}

type SuiRawData struct {
	MoveObject *SuiRawMoveObject  `json:"moveObject,omitempty"`
	Package    *SuiRawMovePackage `json:"package,omitempty"`
}

func (r SuiRawData) Tag() string {
	return "dataType"
}

func (r SuiRawData) Content() string {
	return ""
}

type SuiRawMoveObject struct {
	Type              iotago.StructTag      `json:"type"`
	HasPublicTransfer bool                  `json:"hasPublicTransfer"`
	Version           iotago.SequenceNumber `json:"version"`
	BcsBytes          iotago.Base64Data     `json:"bcsBytes"`
}

type SuiRawMovePackage struct {
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
	ModuleName string       `json:"moduleName"`
	StructName string          `json:"structName"`
	Package    iotago.ObjectID `json:"package"`
}

type SuiObjectData struct {
	ObjectID *iotago.ObjectID     `json:"objectId"`
	Version  *BigInt              `json:"version"`
	Digest   *iotago.ObjectDigest `json:"digest"`
	/**
	 * Type of the object, default to be undefined unless SuiObjectDataOptions.showType is set to true
	 */
	Type *string `json:"type,omitempty"`
	/**
	 * Move object content or package content, default to be undefined unless SuiObjectDataOptions.showContent is set to true
	 */
	Content *serialization.TagJson[SuiParsedData] `json:"content,omitempty"`
	/**
	 * Move object content or package content in BCS bytes, default to be undefined unless SuiObjectDataOptions.showBcs is set to true
	 */
	Bcs *serialization.TagJson[SuiRawData] `json:"bcs,omitempty"`
	/**
	 * The owner of this object. Default to be undefined unless SuiObjectDataOptions.showOwner is set to true
	 */
	Owner *ObjectOwner `json:"owner,omitempty"`
	/**
	 * The digest of the transaction that created or last mutated this object.
	 * Default to be undefined unless SuiObjectDataOptions.showPreviousTransaction is set to true
	 */
	PreviousTransaction *iotago.TransactionDigest `json:"previousTransaction,omitempty"`
	/**
	 * The amount of SUI we would rebate if this object gets deleted.
	 * This number is re-calculated each time the object is mutated based on
	 * the present storage gas price.
	 * Default to be undefined unless SuiObjectDataOptions.showStorageRebate is set to true
	 */
	StorageRebate *BigInt `json:"storageRebate,omitempty"`
	/**
	 * Display metadata for this object, default to be undefined unless SuiObjectDataOptions.showDisplay is set to true
	 * This can also be None if the struct type does not have Display defined
	 * See more details in https://forums.sui.io/t/nft-object-display-proposal/4872
	 */
	Display interface{} `json:"display,omitempty"`
}

func (data *SuiObjectData) Ref() iotago.ObjectRef {
	return iotago.ObjectRef{
		ObjectID: data.ObjectID,
		Version:  data.Version.Uint64(),
		Digest:   data.Digest,
	}
}

type SuiObjectDataOptions struct {
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

type SuiObjectResponseError struct {
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

func (e SuiObjectResponseError) Tag() string {
	return "code"
}

func (e SuiObjectResponseError) Content() string {
	return ""
}

type SuiObjectResponse struct {
	Data  *SuiObjectData                                 `json:"data,omitempty"`
	Error *serialization.TagJson[SuiObjectResponseError] `json:"error,omitempty"`
}

type CheckpointSequenceNumber = uint64

type ObjectsPage = Page[SuiObjectResponse, iotago.ObjectID]

type SuiObjectDataFilter struct {
	MatchAll  []*SuiObjectDataFilter `json:"MatchAll,omitempty"`
	MatchAny  []*SuiObjectDataFilter `json:"MatchAny,omitempty"`
	MatchNone []*SuiObjectDataFilter `json:"MatchNone,omitempty"`
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

type SuiObjectResponseQuery struct {
	Filter  *SuiObjectDataFilter  `json:"filter,omitempty"`
	Options *SuiObjectDataOptions `json:"options,omitempty"`
}

type SuiPastObjectResponse = serialization.TagJson[SuiPastObject]

type SuiPastObject struct {
	// The object exists and is found with this version
	VersionFound *SuiObjectData `json:"VersionFound,omitempty"`
	// The object does not exist
	ObjectNotExists *iotago.ObjectID `json:"ObjectNotExists,omitempty"`
	// The object is found to be deleted with this version
	ObjectDeleted *SuiObjectRef `json:"ObjectDeleted,omitempty"`
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
	var err error
	input := data[1 : len(data)-2]
	elts := strings.Split(string(input), ",")
	c.ObjectID, err = iotago.ObjectIDFromHex(elts[0][1 : len(elts[0])-2])
	if err != nil {
		return err
	}
	seq, err := strconv.ParseUint(elts[1], 10, 64)
	if err != nil {
		return err
	}
	c.SequenceNumber = seq
	return nil
}

func (s SuiPastObject) Tag() string {
	return "status"
}

func (s SuiPastObject) Content() string {
	return "details"
}

type SuiGetPastObjectRequest struct {
	ObjectId *iotago.ObjectID `json:"objectId"`
	Version  *BigInt          `json:"version"`
}

type SuiNamePage = Page[string, iotago.ObjectID]
