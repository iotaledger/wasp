#!/bin/sh

openapi-generator-cli generate -i http://localhost:9090/doc/swagger.json \
	-g go \
	--package-name=apiclient \
	--additional-properties preferUnsignedInt=TRUE
