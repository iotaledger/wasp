# AuthInfoModel

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AuthURL** | **string** | JWT only | 
**Scheme** | **string** |  | 

## Methods

### NewAuthInfoModel

`func NewAuthInfoModel(authURL string, scheme string, ) *AuthInfoModel`

NewAuthInfoModel instantiates a new AuthInfoModel object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthInfoModelWithDefaults

`func NewAuthInfoModelWithDefaults() *AuthInfoModel`

NewAuthInfoModelWithDefaults instantiates a new AuthInfoModel object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthURL

`func (o *AuthInfoModel) GetAuthURL() string`

GetAuthURL returns the AuthURL field if non-nil, zero value otherwise.

### GetAuthURLOk

`func (o *AuthInfoModel) GetAuthURLOk() (*string, bool)`

GetAuthURLOk returns a tuple with the AuthURL field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthURL

`func (o *AuthInfoModel) SetAuthURL(v string)`

SetAuthURL sets AuthURL field to given value.


### GetScheme

`func (o *AuthInfoModel) GetScheme() string`

GetScheme returns the Scheme field if non-nil, zero value otherwise.

### GetSchemeOk

`func (o *AuthInfoModel) GetSchemeOk() (*string, bool)`

GetSchemeOk returns a tuple with the Scheme field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScheme

`func (o *AuthInfoModel) SetScheme(v string)`

SetScheme sets Scheme field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


