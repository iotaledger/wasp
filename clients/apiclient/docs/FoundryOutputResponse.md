# FoundryOutputResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Assets** | [**AssetsResponse**](AssetsResponse.md) |  | 
**FoundryId** | **string** |  | 

## Methods

### NewFoundryOutputResponse

`func NewFoundryOutputResponse(assets AssetsResponse, foundryId string, ) *FoundryOutputResponse`

NewFoundryOutputResponse instantiates a new FoundryOutputResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFoundryOutputResponseWithDefaults

`func NewFoundryOutputResponseWithDefaults() *FoundryOutputResponse`

NewFoundryOutputResponseWithDefaults instantiates a new FoundryOutputResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAssets

`func (o *FoundryOutputResponse) GetAssets() AssetsResponse`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *FoundryOutputResponse) GetAssetsOk() (*AssetsResponse, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *FoundryOutputResponse) SetAssets(v AssetsResponse)`

SetAssets sets Assets field to given value.


### GetFoundryId

`func (o *FoundryOutputResponse) GetFoundryId() string`

GetFoundryId returns the FoundryId field if non-nil, zero value otherwise.

### GetFoundryIdOk

`func (o *FoundryOutputResponse) GetFoundryIdOk() (*string, bool)`

GetFoundryIdOk returns a tuple with the FoundryId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFoundryId

`func (o *FoundryOutputResponse) SetFoundryId(v string)`

SetFoundryId sets FoundryId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


