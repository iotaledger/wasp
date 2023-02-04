# API Gen

Prints a Swagger schema without the requirement of running a configured node or http server.

## Usage

`go run main.go > schema.json`

This can be used to generate clients afterwords:

`openapi-generator generate -i schema.json -g go -o testclient`