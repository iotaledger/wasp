#!/bin/sh

openapi-generator-cli generate -i http://localhost:9090/doc/swagger.json \
  --global-property=models,supportingFiles,apis,modelTests=false,apiTests=false \
  -g go \
  -o /mnt/Dev/Coding/iota/wasp/clients/apiclient \
  --package-name=apiclient \
  --additional-properties preferUnsignedInt=TRUE