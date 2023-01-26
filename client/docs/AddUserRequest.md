# AddUserRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Password** | Pointer to **string** |  | [optional] 
**Permissions** | Pointer to **[]string** |  | [optional] 
**Username** | Pointer to **string** |  | [optional] 

## Methods

### NewAddUserRequest

`func NewAddUserRequest() *AddUserRequest`

NewAddUserRequest instantiates a new AddUserRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAddUserRequestWithDefaults

`func NewAddUserRequestWithDefaults() *AddUserRequest`

NewAddUserRequestWithDefaults instantiates a new AddUserRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPassword

`func (o *AddUserRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *AddUserRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *AddUserRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *AddUserRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetPermissions

`func (o *AddUserRequest) GetPermissions() []string`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *AddUserRequest) GetPermissionsOk() (*[]string, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *AddUserRequest) SetPermissions(v []string)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *AddUserRequest) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetUsername

`func (o *AddUserRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *AddUserRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *AddUserRequest) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *AddUserRequest) HasUsername() bool`

HasUsername returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


