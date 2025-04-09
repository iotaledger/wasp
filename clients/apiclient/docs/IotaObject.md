# IotaObject

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **[]int32** |  | 
**Type** | [**ObjectType**](ObjectType.md) |  | 

## Methods

### NewIotaObject

`func NewIotaObject(id []int32, type_ ObjectType, ) *IotaObject`

NewIotaObject instantiates a new IotaObject object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIotaObjectWithDefaults

`func NewIotaObjectWithDefaults() *IotaObject`

NewIotaObjectWithDefaults instantiates a new IotaObject object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *IotaObject) GetId() []int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *IotaObject) GetIdOk() (*[]int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *IotaObject) SetId(v []int32)`

SetId sets Id field to given value.


### GetType

`func (o *IotaObject) GetType() ObjectType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *IotaObject) GetTypeOk() (*ObjectType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *IotaObject) SetType(v ObjectType)`

SetType sets Type field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


