## TODO list for refactor

- [x] sc address, sc color -> chain address, chain color
- [x] chain id -> chain address. In the future -> chain color
- [x] contract id -> chain id + contract index 2 bytes
- [x] request code 2 bytes -> entry point code 4 bytes ((adjusted) hash of the function name)
- [x] request id -> request tx id + request index 2 bytes
- [ ] request target -> contract id
- [ ] state transaction -> block anchor
- [ ] batch of state updates -> block 