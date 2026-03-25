#!/bin/bash
# Post-vendor fix: copy CGO subdirectories that `go work vendor` omits.
# Go's vendor spec only copies Go package directories, skipping C/C++ source
# trees in subdirectories. Some dependencies bundle headers and static libraries
# in subdirectories referenced via CGO CFLAGS/LDFLAGS with ${SRCDIR}.
#
# This script copies those missing subdirectories from the module cache into vendor/.

set -euo pipefail

GOMODCACHE=$(go env GOMODCACHE)
VENDOR_DIR="${1:-vendor}"

# grocksdb: bundles rocksdb headers + static libs in dist/<os_arch>/
GROCKSDB_MOD="github.com/iotaledger/grocksdb"
GROCKSDB_VER=$(grep -m1 "^# ${GROCKSDB_MOD} " "${VENDOR_DIR}/modules.txt" | awk '{print $3}')
if [ -n "${GROCKSDB_VER}" ]; then
    SRC="${GOMODCACHE}/${GROCKSDB_MOD}@${GROCKSDB_VER}/dist"
    DST="${VENDOR_DIR}/${GROCKSDB_MOD}/dist"
    if [ -d "${SRC}" ]; then
        echo "vendor fix: copying ${GROCKSDB_MOD} dist/ (rocksdb headers + static libs)"
        rm -rf "${DST}"
        cp -r "${SRC}" "${DST}"
        chmod -R u+w "${DST}"
    fi
fi

# zondax/hid: bundles libusb + hidapi C sources in subdirectories
HID_MOD="github.com/zondax/hid"
HID_VER=$(grep -m1 "^# ${HID_MOD} " "${VENDOR_DIR}/modules.txt" | awk '{print $3}')
if [ -n "${HID_VER}" ]; then
    for SUBDIR in libusb hidapi; do
        SRC="${GOMODCACHE}/${HID_MOD}@${HID_VER}/${SUBDIR}"
        DST="${VENDOR_DIR}/${HID_MOD}/${SUBDIR}"
        if [ -d "${SRC}" ]; then
            echo "vendor fix: copying ${HID_MOD} ${SUBDIR}/"
            rm -rf "${DST}"
            cp -r "${SRC}" "${DST}"
            chmod -R u+w "${DST}"
        fi
    done
fi

echo "vendor fix: done"
