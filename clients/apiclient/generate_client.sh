#!/bin/bash

function finish {
  rm -f "$SCRIPTPATH/wasp_swagger_schema.json"
  echo "Done"
}
trap finish EXIT


APIGEN_FOLDER="../../tools/api-gen/"
APIGEN_SCRIPT="apigen.sh"

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")

GENERATE_MODE=${1:-cli}

GENERATE_ARGS="\
    --global-property=models,supportingFiles,apis,modelTests=false,apiTests=false \
    -g go \
    --package-name=apiclient \
    --additional-properties preferUnsignedInt=TRUE
"

(cd "$SCRIPTPATH/$APIGEN_FOLDER"; sh -c "./$APIGEN_SCRIPT >| $SCRIPTPATH/wasp_swagger_schema.json")

if [ $GENERATE_MODE = "docker" ]; then
  echo "Generating client with Docker"

  docker run -v "$SCRIPTPATH"/wasp_swagger_schema.json:/tmp/schema.json:ro \
    -v "$SCRIPTPATH":/tmp/apiclient \
    lukasmoe/openapi-generator \
    generate -i "/tmp/schema.json" \
    -o "/tmp/apiclient" \
    $GENERATE_ARGS

else
  echo "Generating client with local CLI"

  openapi-generator-cli generate -i "$SCRIPTPATH/wasp_swagger_schema.json" -o "$SCRIPTPATH" \
    $GENERATE_ARGS
fi

## This is a temporary fix for the blob info response.
## The Schema generator does not properly handle the uint32 type and this is adjusted manually for now.

echo "Patching blob info response int=>uint"

sed -i "/uint32/! s/int32/uint32/g" "$SCRIPTPATH/model_blob_info_response.go"
sed -i "/uint32/! s/int32/uint32/g" "$SCRIPTPATH/docs/BlobInfoResponse.md"
