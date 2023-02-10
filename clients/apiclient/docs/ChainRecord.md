# ChainRecord

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessNodes** | **[]string** |  | 
**IsActive** | **bool** |  | 

## Methods

### NewChainRecord

`func NewChainRecord(accessNodes []string, isActive bool, ) *ChainRecord`

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


### GetIsActive

`func (o *ChainRecord) GetIsActive() bool`

GetIsActive returns the IsActive field if non-nil, zero value otherwise.

### GetIsActiveOk

`func (o *ChainRecord) GetIsActiveOk() (*bool, bool)`

GetIsActiveOk returns a tuple with the IsActive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsActive

`func (o *ChainRecord) SetIsActive(v bool)`

SetIsActive sets IsActive field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


