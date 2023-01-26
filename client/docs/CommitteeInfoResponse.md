# CommitteeInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessNodes** | Pointer to [**[]CommitteeNode**](CommitteeNode.md) | A list of all access nodes and their peering info. | [optional] 
**Active** | Pointer to **bool** | Whether or not the chain is active. | [optional] 
**CandidateNodes** | Pointer to [**[]CommitteeNode**](CommitteeNode.md) | A list of all candidate nodes and their peering info. | [optional] 
**ChainId** | Pointer to **string** | ChainID (Bech32-encoded). | [optional] 
**CommitteeNodes** | Pointer to [**[]CommitteeNode**](CommitteeNode.md) | A list of all committee nodes and their peering info. | [optional] 
**StateAddress** | Pointer to **string** |  | [optional] 

## Methods

### NewCommitteeInfoResponse

`func NewCommitteeInfoResponse() *CommitteeInfoResponse`

NewCommitteeInfoResponse instantiates a new CommitteeInfoResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommitteeInfoResponseWithDefaults

`func NewCommitteeInfoResponseWithDefaults() *CommitteeInfoResponse`

NewCommitteeInfoResponseWithDefaults instantiates a new CommitteeInfoResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessNodes

`func (o *CommitteeInfoResponse) GetAccessNodes() []CommitteeNode`

GetAccessNodes returns the AccessNodes field if non-nil, zero value otherwise.

### GetAccessNodesOk

`func (o *CommitteeInfoResponse) GetAccessNodesOk() (*[]CommitteeNode, bool)`

GetAccessNodesOk returns a tuple with the AccessNodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessNodes

`func (o *CommitteeInfoResponse) SetAccessNodes(v []CommitteeNode)`

SetAccessNodes sets AccessNodes field to given value.

### HasAccessNodes

`func (o *CommitteeInfoResponse) HasAccessNodes() bool`

HasAccessNodes returns a boolean if a field has been set.

### GetActive

`func (o *CommitteeInfoResponse) GetActive() bool`

GetActive returns the Active field if non-nil, zero value otherwise.

### GetActiveOk

`func (o *CommitteeInfoResponse) GetActiveOk() (*bool, bool)`

GetActiveOk returns a tuple with the Active field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActive

`func (o *CommitteeInfoResponse) SetActive(v bool)`

SetActive sets Active field to given value.

### HasActive

`func (o *CommitteeInfoResponse) HasActive() bool`

HasActive returns a boolean if a field has been set.

### GetCandidateNodes

`func (o *CommitteeInfoResponse) GetCandidateNodes() []CommitteeNode`

GetCandidateNodes returns the CandidateNodes field if non-nil, zero value otherwise.

### GetCandidateNodesOk

`func (o *CommitteeInfoResponse) GetCandidateNodesOk() (*[]CommitteeNode, bool)`

GetCandidateNodesOk returns a tuple with the CandidateNodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCandidateNodes

`func (o *CommitteeInfoResponse) SetCandidateNodes(v []CommitteeNode)`

SetCandidateNodes sets CandidateNodes field to given value.

### HasCandidateNodes

`func (o *CommitteeInfoResponse) HasCandidateNodes() bool`

HasCandidateNodes returns a boolean if a field has been set.

### GetChainId

`func (o *CommitteeInfoResponse) GetChainId() string`

GetChainId returns the ChainId field if non-nil, zero value otherwise.

### GetChainIdOk

`func (o *CommitteeInfoResponse) GetChainIdOk() (*string, bool)`

GetChainIdOk returns a tuple with the ChainId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChainId

`func (o *CommitteeInfoResponse) SetChainId(v string)`

SetChainId sets ChainId field to given value.

### HasChainId

`func (o *CommitteeInfoResponse) HasChainId() bool`

HasChainId returns a boolean if a field has been set.

### GetCommitteeNodes

`func (o *CommitteeInfoResponse) GetCommitteeNodes() []CommitteeNode`

GetCommitteeNodes returns the CommitteeNodes field if non-nil, zero value otherwise.

### GetCommitteeNodesOk

`func (o *CommitteeInfoResponse) GetCommitteeNodesOk() (*[]CommitteeNode, bool)`

GetCommitteeNodesOk returns a tuple with the CommitteeNodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommitteeNodes

`func (o *CommitteeInfoResponse) SetCommitteeNodes(v []CommitteeNode)`

SetCommitteeNodes sets CommitteeNodes field to given value.

### HasCommitteeNodes

`func (o *CommitteeInfoResponse) HasCommitteeNodes() bool`

HasCommitteeNodes returns a boolean if a field has been set.

### GetStateAddress

`func (o *CommitteeInfoResponse) GetStateAddress() string`

GetStateAddress returns the StateAddress field if non-nil, zero value otherwise.

### GetStateAddressOk

`func (o *CommitteeInfoResponse) GetStateAddressOk() (*string, bool)`

GetStateAddressOk returns a tuple with the StateAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateAddress

`func (o *CommitteeInfoResponse) SetStateAddress(v string)`

SetStateAddress sets StateAddress field to given value.

### HasStateAddress

`func (o *CommitteeInfoResponse) HasStateAddress() bool`

HasStateAddress returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


