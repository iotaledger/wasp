# .UsersApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**addUser**](UsersApi.md#addUser) | **POST** /v1/users | Add a user
[**changeUserPassword**](UsersApi.md#changeUserPassword) | **PUT** /v1/users/{username}/password | Change user password
[**changeUserPermissions**](UsersApi.md#changeUserPermissions) | **PUT** /v1/users/{username}/permissions | Change user permissions
[**deleteUser**](UsersApi.md#deleteUser) | **DELETE** /v1/users/{username} | Deletes a user
[**getUser**](UsersApi.md#getUser) | **GET** /v1/users/{username} | Get a user
[**getUsers**](UsersApi.md#getUsers) | **GET** /v1/users | Get a list of all users


# **addUser**
> void addUser(addUserRequest)


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .UsersApi(configuration);

let body:.UsersApiAddUserRequest = {
  // AddUserRequest | The user data
  addUserRequest: {
    password: "password_example",
    permissions: [
      "permissions_example",
    ],
    username: "username_example",
  },
};

apiInstance.addUser(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **addUserRequest** | **AddUserRequest**| The user data |


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | User successfully added |  -  |
**400** | Invalid request |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **changeUserPassword**
> void changeUserPassword(updateUserPasswordRequest)


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .UsersApi(configuration);

let body:.UsersApiChangeUserPasswordRequest = {
  // string | The username
  username: "username_example",
  // UpdateUserPasswordRequest | The users new password
  updateUserPasswordRequest: {
    password: "password_example",
  },
};

apiInstance.changeUserPassword(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updateUserPasswordRequest** | **UpdateUserPasswordRequest**| The users new password |
 **username** | [**string**] | The username | defaults to undefined


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | User successfully updated |  -  |
**400** | Invalid request |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | User not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **changeUserPermissions**
> void changeUserPermissions(updateUserPermissionsRequest)


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .UsersApi(configuration);

let body:.UsersApiChangeUserPermissionsRequest = {
  // string | The username
  username: "username_example",
  // UpdateUserPermissionsRequest | The users new permissions
  updateUserPermissionsRequest: {
    permissions: [
      "permissions_example",
    ],
  },
};

apiInstance.changeUserPermissions(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updateUserPermissionsRequest** | **UpdateUserPermissionsRequest**| The users new permissions |
 **username** | [**string**] | The username | defaults to undefined


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | User successfully updated |  -  |
**400** | Invalid request |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | User not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **deleteUser**
> void deleteUser()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .UsersApi(configuration);

let body:.UsersApiDeleteUserRequest = {
  // string | The username
  username: "username_example",
};

apiInstance.deleteUser(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **username** | [**string**] | The username | defaults to undefined


### Return type

**void**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Deletes a specific user |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | User not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getUser**
> User getUser()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .UsersApi(configuration);

let body:.UsersApiGetUserRequest = {
  // string | The username
  username: "username_example",
};

apiInstance.getUser(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **username** | [**string**] | The username | defaults to undefined


### Return type

**User**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Returns a specific user |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |
**404** | User not found |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)

# **getUsers**
> Array<User> getUsers()


### Example


```typescript
import {  } from '';
import * as fs from 'fs';

const configuration = .createConfiguration();
const apiInstance = new .UsersApi(configuration);

let body:any = {};

apiInstance.getUsers(body).then((data:any) => {
  console.log('API called successfully. Returned data: ' + data);
}).catch((error:any) => console.error(error));
```


### Parameters
This endpoint does not need any parameter.


### Return type

**Array<User>**

### Authorization

[Authorization](README.md#Authorization)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | A list of all users |  -  |
**401** | Unauthorized (Wrong permissions, missing token) |  -  |

[[Back to top]](#) [[Back to API list]](README.md#documentation-for-api-endpoints) [[Back to Model list]](README.md#documentation-for-models) [[Back to README]](README.md)


