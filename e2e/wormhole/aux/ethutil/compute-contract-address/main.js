#!/usr/bin/env node

const rlp = require('rlp');
const keccak = require('keccak');

function convert(integer) {
    var str = Number(integer).toString(16);
    return str.length == 1 ? "0" + str : str;
};

function getaddr(nonce, sender) {
    nonce = '0x' + convert(nonce);
    var input_arr = [ sender, nonce ];
    var rlp_encoded = rlp.encode(input_arr);

    var rlp_encoded_buf = new Buffer.from(rlp_encoded);
    var contract_address_long = keccak('keccak256').update(rlp_encoded_buf).digest('hex');

    var contract_address = contract_address_long.substring(24); //Trim the first 24 characters.
    return contract_address;
}

const args = process.argv.slice(2);
if (args.length != 2) {
    console.log('Usage: ' + process.argv[1] + ' ETH_FROM_address nonce');
} else {
    console.log(getaddr(args[1], args[0]));
}
