# OnLedgerRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The request ID | 
**Output** | [**Output**](Output.md) |  | 
**OutputId** | **string** | The output ID | 
**Raw** | **string** | The raw data of the request (Hex) | 

## Methods

### NewOnLedgerRequest

`func NewOnLedgerRequest(id string, output Output, outputId string, raw string, ) *OnLedgerRequest`

NewOnLedgerRequest instantiates a new OnLedgerRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOnLedgerRequestWithDefaults

`func NewOnLedgerRequestWithDefaults() *OnLedgerRequest`

NewOnLedgerRequestWithDefaults instantiates a new OnLedgerRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *OnLedgerRequest) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *OnLedgerRequest) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *OnLedgerRequest) SetId(v string)`

SetId sets Id field to given value.


### GetOutput

`func (o *OnLedgerRequest) GetOutput() Output`

GetOutput returns the Output field if non-nil, zero value otherwise.

### GetOutputOk

`func (o *OnLedgerRequest) GetOutputOk() (*Output, bool)`

GetOutputOk returns a tuple with the Output field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutput

`func (o *OnLedgerRequest) SetOutput(v Output)`

SetOutput sets Output field to given value.


### GetOutputId

`func (o *OnLedgerRequest) GetOutputId() string`

GetOutputId returns the OutputId field if non-nil, zero value otherwise.

### GetOutputIdOk

`func (o *OnLedgerRequest) GetOutputIdOk() (*string, bool)`

GetOutputIdOk returns a tuple with the OutputId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutputId

`func (o *OnLedgerRequest) SetOutputId(v string)`

SetOutputId sets OutputId field to given value.


### GetRaw

`func (o *OnLedgerRequest) GetRaw() string`

GetRaw returns the Raw field if non-nil, zero value otherwise.

### GetRawOk

`func (o *OnLedgerRequest) GetRawOk() (*string, bool)`

GetRawOk returns a tuple with the Raw field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRaw

`func (o *OnLedgerRequest) SetRaw(v string)`

SetRaw sets Raw field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


