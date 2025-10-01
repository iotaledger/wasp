#!/bin/bash
set -euo pipefail

root_path=$(git rev-parse --show-toplevel)

git clone --depth=1 --branch sdk-bindings --single-branch https://github.com/iotaledger/iota-rust-sdk.git "$root_path/clients/bindings/iota-rust-sdk" || true

iota_binding_dir="$root_path/clients/bindings"
iota_go_ffi_dir="$iota_binding_dir/iota_sdk_ffi"
iota_go_ffi_file="$iota_go_ffi_dir/iota_sdk_ffi.go"

iota_rust_sdk_path="$iota_binding_dir/iota-rust-sdk"
go_ffi_dir="$iota_rust_sdk_path/bindings/go/iota_sdk_ffi/iota_sdk_ffi"
go_ffi_file="$go_ffi_dir/iota_sdk_ffi.go"

go_ffi_target_dir="$root_path/clients/bindings/iota-rust-sdk/target"

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

command -v uniffi-bindgen-go >/dev/null || {
echo "uniffi-bindgen-go not found in PATH" >&2
exit 1
}
[[ -f "$lib_path" ]] || {
echo "Library not found: $lib_path" >&2
exit 1
}

# Clean up old generated files before generating new ones
rm -rf "$root_path/clients/bindings/iota_sdk_ffi"

uniffi-bindgen-go --library "$lib_path" --out-dir "$iota_binding_dir" --no-format
echo "Generated Go bindings to $iota_binding_dir using $lib_path"

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


# cp -R $go_ffi_target_dir $root_path/clients/bindings/iota_sdk_ffi



