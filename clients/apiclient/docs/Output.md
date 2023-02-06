# Output

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OutputType** | **int32** | The output type | 
**Raw** | **string** | The raw data of the output (Hex) | 

## Methods

### NewOutput

`func NewOutput(outputType int32, raw string, ) *Output`

NewOutput instantiates a new Output object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOutputWithDefaults

`func NewOutputWithDefaults() *Output`

NewOutputWithDefaults instantiates a new Output object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOutputType

`func (o *Output) GetOutputType() int32`

GetOutputType returns the OutputType field if non-nil, zero value otherwise.

### GetOutputTypeOk

`func (o *Output) GetOutputTypeOk() (*int32, bool)`

GetOutputTypeOk returns a tuple with the OutputType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutputType

`func (o *Output) SetOutputType(v int32)`

SetOutputType sets OutputType field to given value.


### GetRaw

`func (o *Output) GetRaw() string`

GetRaw returns the Raw field if non-nil, zero value otherwise.

### GetRawOk

`func (o *Output) GetRawOk() (*string, bool)`

GetRawOk returns a tuple with the Raw field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRaw

`func (o *Output) SetRaw(v string)`

SetRaw sets Raw field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


