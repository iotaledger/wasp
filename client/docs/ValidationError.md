# ValidationError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Error** | Pointer to **string** |  | [optional] 
**MissingPermission** | Pointer to **string** |  | [optional] 

## Methods

### NewValidationError

`func NewValidationError() *ValidationError`

NewValidationError instantiates a new ValidationError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewValidationErrorWithDefaults

`func NewValidationErrorWithDefaults() *ValidationError`

NewValidationErrorWithDefaults instantiates a new ValidationError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetError

`func (o *ValidationError) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *ValidationError) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *ValidationError) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *ValidationError) HasError() bool`

HasError returns a boolean if a field has been set.

### GetMissingPermission

`func (o *ValidationError) GetMissingPermission() string`

GetMissingPermission returns the MissingPermission field if non-nil, zero value otherwise.

### GetMissingPermissionOk

`func (o *ValidationError) GetMissingPermissionOk() (*string, bool)`

GetMissingPermissionOk returns a tuple with the MissingPermission field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMissingPermission

`func (o *ValidationError) SetMissingPermission(v string)`

SetMissingPermission sets MissingPermission field to given value.

### HasMissingPermission

`func (o *ValidationError) HasMissingPermission() bool`

HasMissingPermission returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


