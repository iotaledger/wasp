api-run:
	go run . webapi-validate http://148.251.89.241:10000 http://148.251.89.241:9090

balance-old:
	curl -X GET "http://148.251.89.241:10000/v1/chains/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5/core/accounts/account/0x04888B6b712D09020D3D0BEb30caDA2f554431F5@iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5/balance?block=50000"

balance-new:
	curl -X GET "http://148.251.89.241:9090/v1/chain/core/accounts/account/0x04888B6b712D09020D3D0BEb30caDA2f554431F5/balance?block=50000"

block-old:
	curl -X POST http://148.251.89.241:10000/v1/chains/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5/evm \
	-H "Content-Type: application/json" \
	-d "{\"jsonrpc\":\"2.0\", \"method\":\"eth_getBlockByNumber\", \"params\":[\"0x5C2909\", false], \"id\":1}"

block-new:
	curl -X POST http://148.251.89.241:9090/v1/chain/evm \
	-H "Content-Type: application/json" \
	-d "{\"jsonrpc\":\"2.0\", \"method\":\"eth_getBlockByNumber\", \"params\":[\"0x5C2909\", false], \"id\":1}"

tx-old:
	curl -X POST http://148.251.89.241:10000/v1/chains/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5/evm \
	-H "Content-Type: application/json" \
	-d "{\"jsonrpc\":\"2.0\", \"method\":\"eth_getTransactionByHash\", \"params\":[\"0x3f2187b233fc45afd586ecfffcc55e27771281b0a7e1b3457f77700559e59b03\"], \"id\":1}"

tx-new:
	curl -X POST http://148.251.89.241:9090/v1/chain/evm \
	-H "Content-Type: application/json" \
	-d "{\"jsonrpc\":\"2.0\", \"method\":\"eth_getTransactionByHash\", \"params\":[\"0x3f2187b233fc45afd586ecfffcc55e27771281b0a7e1b3457f77700559e59b03\"], \"id\":1}"

trace-tx-old:
	curl -X POST http://148.251.89.241:10000/v1/chains/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5/evm \
	-H "Content-Type: application/json" \
	-d "{\"jsonrpc\":\"2.0\", \"method\":\"debug_traceTransaction\", \"params\":[\"0x4c8ceda9be2bd06bb0bf613bee6b01ae293b4f68c28157cd8926ebc9c89af305\", {\"tracer\": \"callTracer\"}], \"id\":1}"

trace-tx-new:
	curl -X POST http://148.251.89.241:9090/v1/chain/evm \
	-H "Content-Type: application/json" \
	-d "{\"jsonrpc\":\"2.0\", \"method\":\"debug_traceTransaction\", \"params\":[\"0x4c8ceda9be2bd06bb0bf613bee6b01ae293b4f68c28157cd8926ebc9c89af305\", {\"tracer\": \"callTracer\"}], \"id\":1}"

trace-block-old:
	curl -X POST http://148.251.89.241:10000/v1/chains/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5/evm \
	-H "Content-Type: application/json" \
	-d "{\"jsonrpc\":\"2.0\", \"method\":\"debug_traceBlockByNumber\", \"params\":[\"0x5ce4e0\", {\"tracer\": \"callTracer\"}], \"id\":1}"

trace-block-new:
	curl -X POST http://148.251.89.241:9090/v1/chain/evm \
	-H "Content-Type: application/json" \
	-d "{\"jsonrpc\":\"2.0\", \"method\":\"debug_traceBlockByNumber\", \"params\":[\"0x5ce4e0\", {\"tracer\": \"callTracer\"}], \"id\":1}"
