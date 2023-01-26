# CommitteeNode

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessAPI** | Pointer to **string** |  | [optional] 
**Node** | Pointer to [**PeeringNodeStatusResponse**](PeeringNodeStatusResponse.md) |  | [optional] 

## Methods

### NewCommitteeNode

`func NewCommitteeNode() *CommitteeNode`

NewCommitteeNode instantiates a new CommitteeNode object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommitteeNodeWithDefaults

`func NewCommitteeNodeWithDefaults() *CommitteeNode`

NewCommitteeNodeWithDefaults instantiates a new CommitteeNode object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessAPI

`func (o *CommitteeNode) GetAccessAPI() string`

GetAccessAPI returns the AccessAPI field if non-nil, zero value otherwise.

### GetAccessAPIOk

`func (o *CommitteeNode) GetAccessAPIOk() (*string, bool)`

GetAccessAPIOk returns a tuple with the AccessAPI field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessAPI

`func (o *CommitteeNode) SetAccessAPI(v string)`

SetAccessAPI sets AccessAPI field to given value.

### HasAccessAPI

`func (o *CommitteeNode) HasAccessAPI() bool`

HasAccessAPI returns a boolean if a field has been set.

### GetNode

`func (o *CommitteeNode) GetNode() PeeringNodeStatusResponse`

GetNode returns the Node field if non-nil, zero value otherwise.

### GetNodeOk

`func (o *CommitteeNode) GetNodeOk() (*PeeringNodeStatusResponse, bool)`

GetNodeOk returns a tuple with the Node field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNode

`func (o *CommitteeNode) SetNode(v PeeringNodeStatusResponse)`

SetNode sets Node field to given value.

### HasNode

`func (o *CommitteeNode) HasNode() bool`

HasNode returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


