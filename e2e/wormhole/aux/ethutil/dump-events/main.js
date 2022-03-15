#!/usr/bin/env node

const Web3 = require('web3');
const fs = require('fs')

function dumpevents(ethhost, abifile, contractaddr) {
    var web3 = new Web3(ethhost);
    fs.readFile(abifile, 'utf8', function (err,data) {
      if (err) {
        return console.log('error: ', err);
      }
      var abi = JSON.parse(data);
      var contractInstance = new web3.eth.Contract(abi, contractaddr);
      contractInstance.getPastEvents('allEvents', {fromBlock: 0, toBlock: 'latest'}, function(e,l){
          if (err) {
            return console.log('error: ', err);
          }
          console.log(JSON.stringify(l));
      });
    });
}

const args = process.argv.slice(2);
if (args.length != 3) {
    console.log('Usage: main.js http://eth-rpc-url path/to/abi/file contract');
    console.log('E.g.:  ' + process.argv[1] + ' http://localhost:8545 e2e/wormhole/optimism-dai-bridge-contracts/out/L1Escrow.abi 0xdeadbeef');
} else {
    dumpevents(args[0], args[1], args[2]);
}
