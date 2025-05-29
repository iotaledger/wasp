# StateAnchor

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Raw** | **string** | The raw data of the anchor (Hex) | 
**StateIndex** | **int32** | The state index | 
**StateMetadata** | **string** | The state metadata | 

## Methods

### NewStateAnchor

`func NewStateAnchor(raw string, stateIndex int32, stateMetadata string, ) *StateAnchor`

NewStateAnchor instantiates a new StateAnchor object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStateAnchorWithDefaults

`func NewStateAnchorWithDefaults() *StateAnchor`

NewStateAnchorWithDefaults instantiates a new StateAnchor object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRaw

`func (o *StateAnchor) GetRaw() string`

GetRaw returns the Raw field if non-nil, zero value otherwise.

### GetRawOk

`func (o *StateAnchor) GetRawOk() (*string, bool)`

GetRawOk returns a tuple with the Raw field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRaw

`func (o *StateAnchor) SetRaw(v string)`

SetRaw sets Raw field to given value.


### GetStateIndex

`func (o *StateAnchor) GetStateIndex() int32`

GetStateIndex returns the StateIndex field if non-nil, zero value otherwise.

### GetStateIndexOk

`func (o *StateAnchor) GetStateIndexOk() (*int32, bool)`

GetStateIndexOk returns a tuple with the StateIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateIndex

`func (o *StateAnchor) SetStateIndex(v int32)`

SetStateIndex sets StateIndex field to given value.


### GetStateMetadata

`func (o *StateAnchor) GetStateMetadata() string`

GetStateMetadata returns the StateMetadata field if non-nil, zero value otherwise.

### GetStateMetadataOk

`func (o *StateAnchor) GetStateMetadataOk() (*string, bool)`

GetStateMetadataOk returns a tuple with the StateMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateMetadata

`func (o *StateAnchor) SetStateMetadata(v string)`

SetStateMetadata sets StateMetadata field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


