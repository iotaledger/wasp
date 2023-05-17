package routes

/**
As we generate the client by the generated API schema, it is generally not needed to put all routes for each handler in a separate file. (Like we did in the previous WebAPI).

For the metadata feature we need to combine the chains publicAPI path with the EVM suffixes below.
To make sure we only need to change them once, they are placed here.
*/

const (
	EVMJsonRPCPathSuffix       = "evm"
	EVMJsonWebSocketPathSuffix = "evm/ws"
)
