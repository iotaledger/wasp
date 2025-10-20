#!/bin/bash
set -euo pipefail

root_path=$(git rev-parse --show-toplevel)

# use commit b921ab3e65acbe377cd4da0ec6d8421fecc15ad8
# git clone --branch modify-sdk --single-branch git@github.com:howjmay/iota-rust-sdk.git "$root_path/clients/bindings/iota-rust-sdk" || true
# cd $root_path/clients/bindings/iota-rust-sdk
# git switch --detach b921ab3e65acbe377cd4da0ec6d8421fecc15ad8
# cd ..

iota_binding_dir="$root_path/clients/bindings"
iota_go_ffi_dir="$iota_binding_dir/iota_sdk_ffi"
iota_go_ffi_file="$iota_go_ffi_dir/iota_sdk_ffi.go"

iota_rust_sdk_path="$iota_binding_dir/iota-rust-sdk"
iota_rust_go_ffi_dir="$iota_rust_sdk_path/bindings/go/iota_sdk_ffi"

cd "$iota_rust_sdk_path"
cargo build --all-features -p iota-sdk-ffi --lib --release

lib_base="$iota_rust_sdk_path/target/release/libiota_sdk_ffi"
case "$(uname -s)" in
    Darwin*) ext="dylib" ;;
    Linux*)  ext="so" ;;
    MINGW*|MSYS*|CYGWIN*) ext="dll" ;;
*) echo "Unsupported OS: $(uname -s)"; exit 1 ;;
esac
lib_path="${lib_base}.${ext}"

[[ -f "$lib_path" ]] || {
echo "Library not found: $lib_path" >&2
exit 1
}

# Overwrite local Go FFI bindings with those from iota-rust-sdk
rm -rf "$iota_go_ffi_dir"
if [[ -d "$iota_rust_go_ffi_dir" ]]; then
  cp -R "$iota_rust_go_ffi_dir" "$iota_binding_dir/"
  echo "Copied Go bindings from $iota_rust_go_ffi_dir to $iota_binding_dir"
else
  echo "Source bindings not found: $iota_rust_go_ffi_dir" >&2
  exit 1
fi

if [[ -f "$iota_go_ffi_file" ]]; then
  # Only replace if the placeholder include line exists

  if grep -qE '^// #include <iota_sdk_ffi\.h>$' "$iota_go_ffi_file"; then
    # Replace the single-line include with a full cgo block
    # Note: Escape $ in ${SRCDIR} as \${SRCDIR} so Perl doesn't interpolate it.
    perl -0777 -pe 's{// #include <iota_sdk_ffi\.h>}{/*\n#cgo darwin  LDFLAGS: -L\${SRCDIR}/../iota-rust-sdk/target/release -liota_sdk_ffi -Wl,-rpath,\${SRCDIR}/../iota-rust-sdk/target/release\n#cgo linux   LDFLAGS: -L\${SRCDIR}/../iota-rust-sdk/target/release -liota_sdk_ffi -Wl,-rpath,\${SRCDIR}/../iota-rust-sdk/target/release\n#cgo windows LDFLAGS: -L\${SRCDIR}/../iota-rust-sdk/target/release -liota_sdk_ffi\n#include "iota_sdk_ffi.h"\n*/}g' -i.bak "$iota_go_ffi_file"
    rm -f "$iota_go_ffi_file.bak"
    echo "Patched cgo block in: $iota_go_ffi_file"
  else
    echo "No placeholder include found; skipping patch: $iota_go_ffi_file"
  fi
else
  echo "File not found: $iota_go_ffi_file" >&2
  exit 1
fi
