#!/bin/bash
CURRENT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

COMMIT=$1
MODULES="app constraints crypto ds kvstore lo logger objectstorage runtime serializer/v2 web"

if [ -z "$COMMIT" ]
then
    echo "ERROR: no commit hash given!"
    exit 1
fi

for i in $MODULES
do
	go get -u github.com/iotaledger/hive.go/$i@$COMMIT
done

bash ${CURRENT_DIR}/go_mod_tidy.sh
