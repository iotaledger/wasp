const {test} = require('zora');
const Web3 = require('web3');

const addr = '0x82a6fd991da00572a21348fbcb54a5d1b1e0300f';
const faucetAddress = '0x71562b71999873DB5b286dF957af199Ec94617F7';

let web3 = new Web3(new Web3.providers.HttpProvider('http://localhost:8545'));

web3.extend({
    methods: [{
        name: 'requestFunds',
        call: 'test_requestFunds',
        params: 1,
        inputFormatter: [web3.extend.formatters.inputAddressFormatter]
    }]
});

test('all', async t => {
  await t.test('basics', async t => {
    t.equal(await web3.eth.getBalance(addr), "0");
    t.equal(await web3.eth.getTransactionCount(faucetAddress, 'latest'), 0);
    t.equal(await web3.eth.getTransactionCount(addr, 'latest'), 0);
    t.equal(await web3.eth.getBlockNumber(), 0);

    await web3.requestFunds(addr);

    t.equal(await web3.eth.getBalance(addr), "1000000000000000000");
    t.equal(await web3.eth.getTransactionCount(faucetAddress, 'latest'), 1);
    t.equal(await web3.eth.getTransactionCount(addr, 'latest'), 0);
    t.equal(await web3.eth.getBlockNumber(), 1);

    await web3.requestFunds(addr);

    t.equal(await web3.eth.getBalance(addr), "2000000000000000000");
    t.equal(await web3.eth.getTransactionCount(faucetAddress, 'latest'), 2);
    t.equal(await web3.eth.getTransactionCount(addr, 'latest'), 0);
    t.equal(await web3.eth.getBlockNumber(), 2);
  });
});
