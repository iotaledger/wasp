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

GENERATE_ARGS_TS="\
    --global-property=models,supportingFiles,apis,modelTests=false,apiTests=false \
    -g typescript \
    --package-name=apiclient-ts \
    --additional-properties preferUnsignedInt=TRUE
"

GENERATE_ARGS_GO="\
    --global-property=models,supportingFiles,apis,modelTests=false,apiTests=false \
    -g go \
    --package-name=apiclient \
    --additional-properties preferUnsignedInt=TRUE
"

(cd "$SCRIPTPATH/$APIGEN_FOLDER"; sh -c "./$APIGEN_SCRIPT >| $SCRIPTPATH/wasp_swagger_schema.json")

cp $SCRIPTPATH/.openapi-generator/FILES /tmp/prev_openapi_generator_files

if [ $GENERATE_MODE = "docker" ]; then
  echo "Generating client with Docker"

  docker pull lukasmoe/openapi-generator:latest
  docker run -u 1000 -v "$SCRIPTPATH"/wasp_swagger_schema.json:/tmp/schema.json:ro \
    -v "$SCRIPTPATH":/tmp/apiclient-ts \
    lukasmoe/openapi-generator:latest \
    generate -i "/tmp/schema.json" \
    -o "/tmp/apiclient-ts" \
    $GENERATE_ARGS_TS

  docker run -u 1000 -v "$SCRIPTPATH"/wasp_swagger_schema.json:/tmp/schema.json:ro \
      -v "$SCRIPTPATH":/tmp/apiclient-go \
      lukasmoe/openapi-generator:latest \
      generate -i "/tmp/schema.json" \
      -o "/tmp/apiclient-go" \
      $GENERATE_ARGS_GO

else
  echo "Generating client with local CLI"

  openapi-generator-cli generate -i "$SCRIPTPATH/wasp_swagger_schema.json" -o "$SCRIPTPATH" \
    $GENERATE_ARGS_TS

  openapi-generator-cli generate -i "$SCRIPTPATH/wasp_swagger_schema.json" -o "$SCRIPTPATH" \
    $GENERATE_ARGS_GO
fi

## This is a temporary fix for the blob info response.
## The Schema generator does not properly handle the uint32 type and this is adjusted manually for now.

## For some reason NullableInt is generated both in model_int.go and utils.go
## This is a temporary fix to remove the duplicate definition until we find its cause.
sed -i "/uint32/! s/NullableInt(/nullableIntUnused(/g" "$SCRIPTPATH/utils.go"
sed -i "/uint32/! s/NullableInt)/nullableIntUnused)/g" "$SCRIPTPATH/utils.go"
sed -i "/uint32/! s/NullableInt{/nullableIntUnused{/g" "$SCRIPTPATH/utils.go"
sed -i "/uint32/! s/NullableInt /nullableIntUnused /g" "$SCRIPTPATH/utils.go"

## Deleting obsolete files
OBSOLETE_FILES=$(comm -23 <(sort /tmp/prev_openapi_generator_files) <(sort $SCRIPTPATH/.openapi-generator/FILES))
if [ ! -z "$OBSOLETE_FILES" ]; then
  DELETED_FILES_DIR=/tmp/openapi_generator_deleted_files_$(date +%F_%H-%M-%S)
  
  echo "Deleting obsolete files - moving them to $DELETED_FILES_DIR:"
  echo $OBSOLETE_FILES
  (
    mkdir -p $DELETED_FILES_DIR &&
    cd $SCRIPTPATH &&
    echo -n $OBSOLETE_FILES | xargs -d' ' -I{} mv {} $DELETED_FILES_DIR/
  )
fi