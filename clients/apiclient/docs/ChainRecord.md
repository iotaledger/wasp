# ChainRecord

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessNodes** | Pointer to **[]string** |  | [optional] 
**Active** | Pointer to **bool** | Whether or not the chain is active | [optional] 
**ChainId** | Pointer to **string** | ChainID (bech32) | [optional] 

## Methods

### NewChainRecord

`func NewChainRecord() *ChainRecord`

NewChainRecord instantiates a new ChainRecord object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewChainRecordWithDefaults

`func NewChainRecordWithDefaults() *ChainRecord`

NewChainRecordWithDefaults instantiates a new ChainRecord object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessNodes

`func (o *ChainRecord) GetAccessNodes() []string`

GetAccessNodes returns the AccessNodes field if non-nil, zero value otherwise.

### GetAccessNodesOk

`func (o *ChainRecord) GetAccessNodesOk() (*[]string, bool)`

GetAccessNodesOk returns a tuple with the AccessNodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessNodes

`func (o *ChainRecord) SetAccessNodes(v []string)`

SetAccessNodes sets AccessNodes field to given value.

### HasAccessNodes

`func (o *ChainRecord) HasAccessNodes() bool`

HasAccessNodes returns a boolean if a field has been set.

### GetActive

`func (o *ChainRecord) GetActive() bool`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *ChainRecord) GetActiveOk() (*bool, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *ChainRecord) SetActive(v bool)`

SetActive sets Active field to given value.

### HasActive

`func (o *ChainRecord) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetChainId

`func (o *ChainRecord) GetChainId() string`

GetChainId returns the ChainId field if non-nil, zero value otherwise.

### GetChainIdOk

`func (o *ChainRecord) GetChainIdOk() (*string, bool)`

GetChainIdOk returns a tuple with the ChainId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChainId

`func (o *ChainRecord) SetChainId(v string)`

SetChainId sets ChainId field to given value.

### HasChainId

`func (o *ChainRecord) HasChainId() bool`

HasChainId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


