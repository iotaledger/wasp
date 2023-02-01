# Allowance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FungibleTokens** | [**FungibleTokens**](FungibleTokens.md) |  | 
**Nfts** | **[]string** |  | 

## Methods

### NewAllowance

`func NewAllowance(fungibleTokens FungibleTokens, nfts []string, ) *Allowance`

NewAllowance instantiates a new Allowance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAllowanceWithDefaults

`func NewAllowanceWithDefaults() *Allowance`

NewAllowanceWithDefaults instantiates a new Allowance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFungibleTokens

`func (o *Allowance) GetFungibleTokens() FungibleTokens`

GetFungibleTokens returns the FungibleTokens field if non-nil, zero value otherwise.

### GetFungibleTokensOk

`func (o *Allowance) GetFungibleTokensOk() (*FungibleTokens, bool)`

GetFungibleTokensOk returns a tuple with the FungibleTokens field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFungibleTokens

`func (o *Allowance) SetFungibleTokens(v FungibleTokens)`

SetFungibleTokens sets FungibleTokens field to given value.


### GetNfts

`func (o *Allowance) GetNfts() []string`

GetNfts returns the Nfts field if non-nil, zero value otherwise.

### GetNftsOk

`func (o *Allowance) GetNftsOk() (*[]string, bool)`

GetNftsOk returns a tuple with the Nfts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfts

`func (o *Allowance) SetNfts(v []string)`

SetNfts sets Nfts field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


