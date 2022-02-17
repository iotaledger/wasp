find . -type f -name "consts.*" -delete
find . -type f -name "contract.*" -delete
find . -type f -name "keys.*" -delete
find . -type f -name "lib.*" -delete
find . -type f -name "params.*" -delete
find . -type f -name "results.*" -delete
find . -type f -name "state.*" -delete
find . -type f -name "typedefs.*" -delete
find . -type f -name "types.*" -delete
find . -type f -name "main.go" -delete
find . -type f -name "*.wasm" -delete

# remove careful, this could fuck up fairroulette frontend
for dir in ./*; do
 if [ -d "$dir" ]; then
    find . -type f -name "$dir/ts/$dir/index.ts" -delete
    find . -type f -name "$dir/ts/$dir/tsconfig.json" -delete
    find . -type f -name "$dir/pkg/*.*" -delete
    find . -type f -name "$dir/ts/pkg/*.*" -delete
  fi
done
find . -type f -name "target/*.*" -delete
