# Item

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | Pointer to **string** | key (hex-encoded) | [optional] 
**Value** | Pointer to **string** | value (hex-encoded) | [optional] 

## Methods

### NewItem

`func NewItem() *Item`

NewItem instantiates a new Item object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewItemWithDefaults

`func NewItemWithDefaults() *Item`

NewItemWithDefaults instantiates a new Item object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *Item) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *Item) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *Item) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *Item) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetValue

`func (o *Item) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *Item) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *Item) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *Item) HasValue() bool`

HasValue returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


