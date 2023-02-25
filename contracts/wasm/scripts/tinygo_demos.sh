cd ..

tinygo build -o corecontracts/go/pkg/corecontracts_go.wasm -target wasm -gc=leaking -opt 2 -no-debug corecontracts/go/main.go

tinygo build -o dividend/go/pkg/dividend_go.wasm -target wasm -gc=leaking -opt 2 -no-debug dividend/go/main.go

tinygo build -o donatewithfeedback/go/pkg/donatewithfeedback_go.wasm -target wasm -gc=leaking -opt 2 -no-debug donatewithfeedback/go/main.go

tinygo build -o erc20/go/pkg/erc20_go.wasm -target wasm -gc=leaking -opt 2 -no-debug erc20/go/main.go

tinygo build -o erc721/go/pkg/erc721_go.wasm -target wasm -gc=leaking -opt 2 -no-debug erc721/go/main.go

tinygo build -o fairauction/go/pkg/fairauction_go.wasm -target wasm -gc=leaking -opt 2 -no-debug fairauction/go/main.go

tinygo build -o fairroulette/go/pkg/fairroulette_go.wasm -target wasm -gc=leaking -opt 2 -no-debug fairroulette/go/main.go

tinygo build -o helloworld/go/pkg/helloworld_go.wasm -target wasm -gc=leaking -opt 2 -no-debug helloworld/go/main.go

tinygo build -o inccounter/go/pkg/inccounter_go.wasm -target wasm -gc=leaking -opt 2 -no-debug inccounter/go/main.go

tinygo build -o schemacomment/go/pkg/schemacomment_go.wasm -target wasm -gc=leaking -opt 2 -no-debug schemacomment/go/main.go

tinygo build -o testcore/go/pkg/testcore_go.wasm -target wasm -gc=leaking -opt 2 -no-debug testcore/go/main.go

tinygo build -o testwasmlib/go/pkg/testwasmlib_go.wasm -target wasm -gc=leaking -opt 2 -no-debug testwasmlib/go/main.go

tinygo build -o timestamp/go/pkg/timestamp_go.wasm -target wasm -gc=leaking -opt 2 -no-debug timestamp/go/main.go

tinygo build -o tokenregistry/go/pkg/tokenregistry_go.wasm -target wasm -gc=leaking -opt 2 -no-debug tokenregistry/go/main.go

cd scripts
