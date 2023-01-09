pragma solidity >=0.7.0 <0.9.0;

contract Consumers {
    address private owner;
    
    string[] public consumers;
    
    modifier onlyOwner {
        require(msg.sender == owner);
        _;
    }
    
    constructor() {
        owner = msg.sender;
    }
    
    function getConsumers() public view returns (string[] memory) {
        return consumers;
    }
    
    function add(string[] calldata addrs) public onlyOwner {
        for (uint i = 0; i < addrs.length; i++) {
            consumers.push(addrs[i]);
        }
    }
    
    function remove(uint index) public onlyOwner {
        require(index < consumers.length);
        consumers[index] = consumers[consumers.length-1];
        consumers.pop();
    }
}
