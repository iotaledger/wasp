# RequestProcessedResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChainId** | **string** |  | 
**IsProcessed** | **bool** |  | 
**RequestId** | **string** |  | 

## Methods

### NewRequestProcessedResponse

`func NewRequestProcessedResponse(chainId string, isProcessed bool, requestId string, ) *RequestProcessedResponse`

NewRequestProcessedResponse instantiates a new RequestProcessedResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRequestProcessedResponseWithDefaults

`func NewRequestProcessedResponseWithDefaults() *RequestProcessedResponse`

NewRequestProcessedResponseWithDefaults instantiates a new RequestProcessedResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChainId

`func (o *RequestProcessedResponse) GetChainId() string`

GetChainId returns the ChainId field if non-nil, zero value otherwise.

### GetChainIdOk

`func (o *RequestProcessedResponse) GetChainIdOk() (*string, bool)`

GetChainIdOk returns a tuple with the ChainId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChainId

`func (o *RequestProcessedResponse) SetChainId(v string)`

SetChainId sets ChainId field to given value.


### GetIsProcessed

`func (o *RequestProcessedResponse) GetIsProcessed() bool`

GetIsProcessed returns the IsProcessed field if non-nil, zero value otherwise.

### GetIsProcessedOk

`func (o *RequestProcessedResponse) GetIsProcessedOk() (*bool, bool)`

GetIsProcessedOk returns a tuple with the IsProcessed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsProcessed

`func (o *RequestProcessedResponse) SetIsProcessed(v bool)`

SetIsProcessed sets IsProcessed field to given value.


### GetRequestId

`func (o *RequestProcessedResponse) GetRequestId() string`

GetRequestId returns the RequestId field if non-nil, zero value otherwise.

### GetRequestIdOk

`func (o *RequestProcessedResponse) GetRequestIdOk() (*string, bool)`

GetRequestIdOk returns a tuple with the RequestId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestId

`func (o *RequestProcessedResponse) SetRequestId(v string)`

SetRequestId sets RequestId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


