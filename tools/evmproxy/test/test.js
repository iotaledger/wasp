const {test} = require('zora');
const Web3 = require('web3');

const addr = '0x82a6fd991da00572a21348fbcb54a5d1b1e0300f';

let web3 = new Web3(new Web3.providers.HttpProvider('http://localhost:8545'));

web3.extend({
    methods: [{
        name: 'requestFunds',
        call: 'test_requestFunds',
        params: 1,
        inputFormatter: [web3.extend.formatters.inputAddressFormatter]
    }]
});

test('requestFunds', async t => {
  t.equal(await web3.eth.getBalance(addr), "0");
  await web3.requestFunds(addr);
  t.equal(await web3.eth.getBalance(addr), "1000000000000000000");
});

