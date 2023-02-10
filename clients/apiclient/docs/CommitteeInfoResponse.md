# CommitteeInfoResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessNodes** | [**[]CommitteeNode**](CommitteeNode.md) | A list of all access nodes and their peering info. | 
**Active** | **bool** | Whether or not the chain is active. | 
**CandidateNodes** | [**[]CommitteeNode**](CommitteeNode.md) | A list of all candidate nodes and their peering info. | 
**ChainId** | **string** | ChainID (Bech32-encoded). | 
**CommitteeNodes** | [**[]CommitteeNode**](CommitteeNode.md) | A list of all committee nodes and their peering info. | 
**StateAddress** | **string** |  | 

## Methods

### NewCommitteeInfoResponse

`func NewCommitteeInfoResponse(accessNodes []CommitteeNode, active bool, candidateNodes []CommitteeNode, chainId string, committeeNodes []CommitteeNode, stateAddress string, ) *CommitteeInfoResponse`

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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


