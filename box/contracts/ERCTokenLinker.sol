pragma solidity ^0.4.21;

/*
    ERC-TOKEN合约的映射合约
    用于与主链合约建立连接关系
*/
contract ERCTokenLinker {

    /* 主链地址（合约）*/
    address contractAddress;

    /* 合约二进制接口 */
    string contractABI;

    /* token的名称 */
    string tokenName;

    /* 方法映射ID */
    mapping (bytes32 => bytes4) public methodIDs;

    function ERCTokenLinker(string _tokenName,
        address _contractAddress,
        string _abi,
        bytes4 _transferMethodID,
        bytes4 _balanceOfMethodID,
        ) public {
        tokenName = _tokenName;
        contractAddress = _contractAddress;
        contractABI = _abi;
//        methodIDs[keccak256("transfer")] = _transferMethodID;
//        methodIDs[keccak256("balanceOf")] = _balanceOfMethodID;
        methodIDs = _methodIDs;
    }
}
