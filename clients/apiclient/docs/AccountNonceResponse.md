# AccountNonceResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Nonce** | **string** | The nonce (uint64 as string) | 

## Methods

### NewAccountNonceResponse

`func NewAccountNonceResponse(nonce string, ) *AccountNonceResponse`

NewAccountNonceResponse instantiates a new AccountNonceResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccountNonceResponseWithDefaults

`func NewAccountNonceResponseWithDefaults() *AccountNonceResponse`

NewAccountNonceResponseWithDefaults instantiates a new AccountNonceResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNonce

`func (o *AccountNonceResponse) GetNonce() string`

GetNonce returns the Nonce field if non-nil, zero value otherwise.

### GetNonceOk

`func (o *AccountNonceResponse) GetNonceOk() (*string, bool)`

GetNonceOk returns a tuple with the Nonce field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNonce

`func (o *AccountNonceResponse) SetNonce(v string)`

SetNonce sets Nonce field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


