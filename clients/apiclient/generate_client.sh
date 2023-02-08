#!/bin/sh

function finish {
  rm -f "$SCRIPTPATH/wasp_swagger_schema.json"
}
trap finish EXIT


APIGEN_FOLDER="../../tools/api-gen/"
APIGEN_SCRIPT="apigen.sh"

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")

(cd "$SCRIPTPATH/$APIGEN_FOLDER"; sh -c "./$APIGEN_SCRIPT >| $SCRIPTPATH/wasp_swagger_schema.json")

openapi-generator-cli generate -i "$SCRIPTPATH/wasp_swagger_schema.json" \
  --global-property=models,supportingFiles,apis,modelTests=false,apiTests=false \
  -g go \
  -o "$SCRIPTPATH" \
  --package-name=apiclient \
  --additional-properties preferUnsignedInt=TRUE

## This is a temporary fix for the blob info response.
## The Schema generator does not properly handle the uint32 type and this is adjusted manually for now.

sed -i 's/int32/uint32/' "$SCRIPTPATH/model_blob_info_response.go"
sed -i 's/int32/uint32/' "$SCRIPTPATH/docs/BlobInfoResponse.md"