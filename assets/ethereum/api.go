/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */
package ethereum

import (
	"encoding/json"
	"errors"
	"fmt"

	//"log"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"time"

	tool "github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/bytom/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

type Client struct {
	BaseURL string
	Debug   bool
}

type Response struct {
	Id      int         `json:"id"`
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}

/*
1. eth block example
   "result": {
        "difficulty": "0x1a4f1f",
        "extraData": "0xd98301080d846765746888676f312e31302e338664617277696e",
        "gasLimit": "0x47e7c4",
        "gasUsed": "0x5b61",
        "hash": "0x85319757555e1cf069684dde286e3c34331dc27d2e54bed24e7291f1b84a0cc5",
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "miner": "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776",
        "mixHash": "0xb0cb0abb00c3fc77014abb2a520e3d2a14047cfa30a3b954f18fbeefd1a92f7b",
        "nonce": "0x4df323f58b7a7fd0",
        "number": "0x169cf",
        "parentHash": "0x3df7035473ec98c8c18d2785d5a345193a32b95fcf1ac2d3f09a93109feed3bc",
        "receiptsRoot": "0x441a5be885777bfdf0e985a8ef5046316b3384dd49db7ef95b2c546611c1e2fc",
        "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
        "size": "0x2aa",
        "stateRoot": "0xb0d76a848be723c72c9639b2de591320f4456b665354995be08a8fa83897efbb",
        "timestamp": "0x5b7babbe",
        "totalDifficulty": "0x2a844e200a",
        "transactions": [
            {
                "blockHash": "0x85319757555e1cf069684dde286e3c34331dc27d2e54bed24e7291f1b84a0cc5",
                "blockNumber": "0x169cf",
                "from": "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776",
                "gas": "0x15f90",
                "gasPrice": "0x430e23400",
                "hash": "0x925e33ac3ebaf40bb44a843860b6589ea2df78c955a27f9df16edcf789519671",
                "input": "0x70a082310000000000000000000000002a63b2203955b84fefe52baca3881b3614991b34",
                "nonce": "0x45",
                "to": "0x8847e5f841458ace82dbb0692c97115799fe28d3",
                "transactionIndex": "0x0",
                "value": "0x0",
                "v": "0x3c",
                "r": "0x8d2ffbe7cb7ac1159a999dfa4352fa27f5cce0df8755254393838aab229ecd33",
                "s": "0xe8ed1f7f8de902ccb008824fe39b2903b94f89e3ea0d5b9f9b880c302bae6cf"
            }
        ],
        "transactionsRoot": "0xa8cb62696679bc3d72762bd2aa5842fdd8aed9c9691fe82064c13e854c13d5cb",
        "uncles": []
    }
*/

type EthBlock struct {
	BlockHeader
	Transactions []BlockTransaction `json:"transactions"`
}

func (this *EthBlock) CreateOpenWalletBlockHeader() *openwallet.BlockHeader {
	header := &openwallet.BlockHeader{
		Hash:              this.BlockHash,
		Previousblockhash: this.PreviousHash,
		Height:            this.blockHeight,
		Time:              uint64(time.Now().Unix()),
	}
	return header
}

func (this *EthBlock) Init() error {
	var err error
	this.blockHeight, err = strconv.ParseUint(removeOxFromHex(this.BlockNumber), 16, 64) //ConvertToBigInt(this.BlockNumber, 16) //
	if err != nil {
		log.Errorf("init blockheight failed, err=%v", err)
		return err
	}
	return nil
}

type TxpoolContent struct {
	Pending map[string]map[string]BlockTransaction `json:"pending"`
}

func (this *TxpoolContent) GetSequentTxNonce(addr string) (uint64, uint64, uint64, error) {
	txpool := this.Pending
	var target map[string]BlockTransaction
	/*if _, exist := txpool[addr]; !exist {
		return 0, 0, 0, nil
	}
	if txpool[addr] == nil {
		return 0, 0, 0, nil
	}

	if len(txpool[addr]) == 0 {
		return 0, 0, 0, nil
	}*/
	for theAddr, _ := range txpool {
		//log.Debugf("theAddr:%v, addr:%v", strings.ToLower(theAddr), strings.ToLower(addr))
		if strings.ToLower(theAddr) == strings.ToLower(addr) {
			target = txpool[theAddr]
		}
	}

	nonceList := make([]interface{}, 0)
	for n, _ := range target {
		tn, err := strconv.ParseUint(n, 10, 64)
		if err != nil {
			log.Error("parse nonce[", n, "] in txpool to uint faile, err=", err)
			return 0, 0, 0, err
		}
		nonceList = append(nonceList, tn)
	}

	sort.Slice(nonceList, func(i, j int) bool {
		if nonceList[i].(uint64) < nonceList[j].(uint64) {
			return true
		}
		return false
	})

	var min, max, count uint64
	for i := 0; i < len(nonceList); i++ {
		if i == 0 {
			min = nonceList[i].(uint64)
			max = min
			count++
		} else if nonceList[i].(uint64) != max+1 {
			break
		} else {
			max++
			count++
		}
	}
	return min, max, count, nil
}

func (this *TxpoolContent) GetPendingTxCountForAddr(addr string) int {
	txpool := this.Pending
	if _, exist := txpool[addr]; !exist {
		return 0
	}
	if txpool[addr] == nil {
		return 0
	}
	return len(txpool[addr])
}

func (this *Client) ethGetTransactionCount(addr string) (uint64, error) {
	params := []interface{}{
		appendOxToAddress(addr),
		"latest",
	}

	result, err := this.Call("eth_getTransactionCount", 1, params)
	if err != nil {
		//errInfo := fmt.Sprintf("get block[%v] failed, err = %v \n", blockNumStr,  err)
		log.Errorf("get transaction count failed, err = %v \n", err)
		return 0, err
	}

	if result.Type != gjson.String {
		log.Errorf("result type failed. ")
		return 0, errors.New("result type failed. ")
	}

	//blockNum, err := ConvertToBigInt(result.String(), 16)
	nonceStr := result.String()
	nonceStr = strings.ToLower(nonceStr)
	nonceStr = removeOxFromHex(nonceStr)
	nonce, err := strconv.ParseUint(nonceStr, 16, 64)
	if err != nil {
		log.Errorf("parse nounce failed, err=%v", err)
		return 0, err
	}
	return nonce, nil
}

func (this *Client) ethGetTxPoolContent() (*TxpoolContent, error) {
	result, err := this.Call("txpool_content", 1, nil)
	if err != nil {
		//errInfo := fmt.Sprintf("get block[%v] failed, err = %v \n", blockNumStr,  err)
		log.Errorf("get tx pool failed, err = %v \n", err)
		return nil, err
	}

	if result.Type != gjson.JSON {
		errInfo := fmt.Sprintf("get tx pool content failed, result type is %v", result.Type)
		log.Errorf(errInfo)
		return nil, errors.New(errInfo)
	}

	var txpool TxpoolContent

	err = json.Unmarshal([]byte(result.Raw), &txpool)
	if err != nil {
		log.Errorf("decode json [%v] failed, err=%v", []byte(result.Raw), err)
		return nil, err
	}

	return &txpool, nil
}

func (this *Client) ethGetTransactionReceipt(transactionId string) (*EthTransactionReceipt, error) {
	params := []interface{}{
		transactionId,
	}

	var txReceipt EthTransactionReceipt
	result, err := this.Call("eth_getTransactionReceipt", 1, params)
	if err != nil {
		//errInfo := fmt.Sprintf("get block[%v] failed, err = %v \n", blockNumStr,  err)
		log.Errorf("get tx[%v] receipt failed, err = %v \n", transactionId, err)
		return nil, err
	}

	if result.Type != gjson.JSON {
		errInfo := fmt.Sprintf("get tx[%v] receipt result type failed, result type is %v", transactionId, result.Type)
		log.Errorf(errInfo)
		return nil, errors.New(errInfo)
	}

	err = json.Unmarshal([]byte(result.Raw), &txReceipt)
	if err != nil {
		log.Errorf("decode json [%v] failed, err=%v", []byte(result.Raw), err)
		return nil, err
	}

	return &txReceipt, nil

}

func (this *Client) ethGetBlockSpecByHash(blockHash string, showTransactionSpec bool) (*EthBlock, error) {
	params := []interface{}{
		blockHash,
		showTransactionSpec,
	}
	var ethBlock EthBlock

	result, err := this.Call("eth_getBlockByHash", 1, params)
	if err != nil {
		//errInfo := fmt.Sprintf("get block[%v] failed, err = %v \n", blockNumStr,  err)
		log.Errorf("get block[%v] failed, err = %v \n", blockHash, err)
		return nil, err
	}

	if result.Type != gjson.JSON {
		errInfo := fmt.Sprintf("get block[%v] result type failed, result type is %v", blockHash, result.Type)
		log.Errorf(errInfo)
		return nil, errors.New(errInfo)
	}

	err = json.Unmarshal([]byte(result.Raw), &ethBlock)
	if err != nil {
		log.Errorf("decode json [%v] failed, err=%v", []byte(result.Raw), err)
		return nil, err
	}

	err = ethBlock.Init()
	if err != nil {
		log.Errorf("init eth block failed, err=%v", err)
		return nil, err
	}
	return &ethBlock, nil
}

func (this *Client) EthGetTransactionByHash(txid string) (*BlockTransaction, error) {
	params := []interface{}{
		appendOxToAddress(txid),
	}

	var tx BlockTransaction

	result, err := this.Call("eth_getTransactionByHash", 1, params)
	if err != nil {
		//errInfo := fmt.Sprintf("get block[%v] failed, err = %v \n", blockNumStr,  err)
		log.Errorf("get transaction[%v] failed, err = %v \n", appendOxToAddress(txid), err)
		return nil, err
	}

	if result.Type != gjson.JSON {
		errInfo := fmt.Sprintf("get transaction[%v] result type failed, result type is %v", appendOxToAddress(txid), result.Type)
		log.Errorf(errInfo)
		return nil, errors.New(errInfo)
	}

	err = json.Unmarshal([]byte(result.Raw), &tx)
	if err != nil {
		log.Errorf("decode json [%v] failed, err=%v", result.Raw, err)
		return nil, err
	}

	return &tx, nil
}

func (this *Client) ethGetBlockSpecByBlockNum2(blockNum string, showTransactionSpec bool) (*EthBlock, error) {
	params := []interface{}{
		blockNum,
		showTransactionSpec,
	}
	var ethBlock EthBlock

	result, err := this.Call("eth_getBlockByNumber", 1, params)
	if err != nil {
		//errInfo := fmt.Sprintf("get block[%v] failed, err = %v \n", blockNumStr,  err)
		log.Errorf("get block[%v] failed, err = %v \n", blockNum, err)
		return nil, err
	}

	if showTransactionSpec {
		err = json.Unmarshal([]byte(result.Raw), &ethBlock)
	} else {
		err = json.Unmarshal([]byte(result.Raw), &ethBlock.BlockHeader)
	}
	if err != nil {
		log.Errorf("decode json [%v] failed, err=%v", result.Raw, err)
		return nil, err
	}

	err = ethBlock.Init()
	if err != nil {
		log.Errorf("init eth block failed, err=%v", err)
		return nil, err
	}
	return &ethBlock, nil
}

func (this *Client) ethGetBlockSpecByBlockNum(blockNum uint64, showTransactionSpec bool) (*EthBlock, error) {
	blockNumStr := "0x" + strconv.FormatUint(blockNum, 16)
	return this.ethGetBlockSpecByBlockNum2(blockNumStr, showTransactionSpec)
}

func (this *Client) ethGetTxpoolStatus() (uint64, uint64, error) {
	result, err := this.Call("txpool_status", 1, nil)
	if err != nil {
		//errInfo := fmt.Sprintf("get block[%v] failed, err = %v \n", blockNumStr,  err)
		//log.Errorf("get block[%v] failed, err = %v \n", err)
		return 0, 0, err
	}

	type TxPoolStatus struct {
		Pending string `json:"pending"`
		Queued  string `json:"queued"`
	}

	txStatusResult := TxPoolStatus{}
	err = json.Unmarshal([]byte(result.Raw), &txStatusResult)
	if err != nil {
		log.Errorf("decode from json failed, err=%v", err)
		return 0, 0, err
	}

	pendingNum, err := strconv.ParseUint(removeOxFromHex(txStatusResult.Pending), 16, 64)
	if err != nil {
		log.Errorf("convert txstatus pending number to uint failed, err=%v", err)
		return 0, 0, err
	}

	queuedNum, err := strconv.ParseUint(removeOxFromHex(txStatusResult.Queued), 16, 64)
	if err != nil {
		log.Errorf("convert queued number to uint failed, err=%v", err)
		return 0, 0, err
	}

	return pendingNum, queuedNum, nil
}

type SolidityParam struct {
	ParamType  string
	ParamValue interface{}
}

func makeRepeatString(c string, count uint) string {
	cs := make([]string, 0)
	for i := 0; i < int(count); i++ {
		cs = append(cs, c)
	}
	return strings.Join(cs, "")
}

func makeTransactionData(methodId string, params []SolidityParam) (string, error) {

	data := methodId
	for i, _ := range params {
		var param string
		if params[i].ParamType == SOLIDITY_TYPE_ADDRESS {
			param = strings.ToLower(params[i].ParamValue.(string))
			if strings.Index(param, "0x") != -1 {
				param = tool.Substr(param, 2, len(param))
			}

			if len(param) != 40 {
				return "", errors.New("length of address error.")
			}
			param = makeRepeatString("0", 24) + param
		} else if params[i].ParamType == SOLIDITY_TYPE_UINT256 {
			intParam := params[i].ParamValue.(*big.Int)
			param = intParam.Text(16)
			l := len(param)
			if l > 64 {
				return "", errors.New("integer overflow.")
			}
			param = makeRepeatString("0", uint(64-l)) + param
			//fmt.Println("makeTransactionData intParam:", intParam.String(), " param:", param)
		} else {
			return "", errors.New("not support solidity type")
		}

		data += param
	}
	return data, nil
}

func (this *Client) ERC20GetAddressBalance2(address string, contractAddr string, sign string) (*big.Int, error) {
	if sign != "latest" && sign != "pending" {
		return nil, errors.New("unknown sign was put through.")
	}
	contractAddr = "0x" + strings.Trim(contractAddr, "0x")
	var funcParams []SolidityParam
	funcParams = append(funcParams, SolidityParam{
		ParamType:  SOLIDITY_TYPE_ADDRESS,
		ParamValue: address,
	})
	trans := make(map[string]interface{})
	data, err := makeTransactionData(ETH_GET_TOKEN_BALANCE_METHOD, funcParams)
	if err != nil {
		log.Errorf("make transaction data failed, err = %v", err)
		return nil, err
	}

	trans["to"] = contractAddr
	trans["data"] = data
	params := []interface{}{
		trans,
		"latest",
	}
	result, err := this.Call("eth_call", 1, params)
	if err != nil {
		log.Errorf(fmt.Sprintf("get addr[%v] erc20 balance failed, err=%v\n", address, err))
		return big.NewInt(0), err
	}
	if result.Type != gjson.String {
		errInfo := fmt.Sprintf("get addr[%v] erc20 balance result type error, result type is %v\n", address, result.Type)
		log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}

	balance, err := ConvertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert addr[%v] erc20 balance format to bigint failed, response is %v, and err = %v\n", address, result.String(), err)
		log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return balance, nil

}

func (this *Client) ERC20GetAddressBalance(address string, contractAddr string) (*big.Int, error) {
	return this.ERC20GetAddressBalance2(address, contractAddr, "pending")
}

func (this *Client) GetAddrBalance2(address string, sign string) (*big.Int, error) {
	if sign != "latest" && sign != "pending" {
		return nil, errors.New("unknown sign was put through.")
	}

	params := []interface{}{
		appendOxToAddress(address),
		sign,
	}
	result, err := this.Call("eth_getBalance", 1, params)
	if err != nil {
		log.Errorf(fmt.Sprintf("get addr[%v] balance failed, err=%v\n", address, err))
		return big.NewInt(0), err
	}
	if result.Type != gjson.String {
		errInfo := fmt.Sprintf("get addr[%v] balance result type error, result type is %v\n", address, result.Type)
		log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}

	balance, err := ConvertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert addr[%v] balance format to bigint failed, response is %v, and err = %v\n", address, result.String(), err)
		log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return balance, nil
}

func (this *Client) GetAddrBalance(address string) (*big.Int, error) {
	return this.GetAddrBalance2(address, "pending")
}

func appendOxToAddress(addr string) string {
	if strings.Index(addr, "0x") == -1 {
		return "0x" + addr
	}
	return addr
}

func signEthTransaction(priKey []byte, toAddr string, amount *big.Int, nonce uint64, data string, fee *txFeeInfo, chainId uint64) (string, error) {
	var err error
	signer := types.NewEIP155Signer(big.NewInt(int64(chainId)))
	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(toAddr),
		amount, fee.GasLimit.Uint64(), fee.GasPrice, common.FromHex(data))

	msg := signer.Hash(tx)
	sig, err := secp256k1.Sign(msg[:], priKey)
	if err != nil {
		log.Error("secp256k1.Sign failed, err=", err)
		return "", err
	}
	/*sig, ret := owcrypt.ETHsignature(priKey, msg[:])
	if ret != owcrypt.SUCCESS {
		errdesc := fmt.Sprintln("signature error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		log.Error(errdesc)
		return "", errors.New(errdesc)
	}*/

	tx, err = tx.WithSignature(signer, sig)
	if err != nil {
		log.Errorf("tx with signature failed, err=%v ", err)
		return "", err
	}

	//jsonStr, _ := json.MarshalIndent(tx, "", " ")
	//log.Info("tx after sign:", string(jsonStr))

	rawTxPara, err := rlp.EncodeToBytes(tx)
	if err != nil {
		log.Errorf("encode tx to rlp failed, err=%v ", err)
		return "", err
	}

	//log.Info("rawTxPara:", common.ToHex(rawTxPara))
	return common.ToHex(rawTxPara), nil
}

func makeSimpleTransactionPara(fromAddr *Address, toAddr string, amount *big.Int, password string, fee *txFeeInfo) map[string]interface{} {
	paraMap := make(map[string]interface{})

	//use password to unlock the account
	paraMap["password"] = password
	//use the following attr to eth_sendTransaction
	paraMap["from"] = appendOxToAddress(fromAddr.Address)
	paraMap["to"] = appendOxToAddress(toAddr)
	paraMap["value"] = "0x" + amount.Text(16)
	paraMap["gas"] = "0x" + fee.GasLimit.Text(16)
	paraMap["gasPrice"] = "0x" + fee.GasPrice.Text(16)
	return paraMap
}

func makeSimpleTransactiomnPara2(fromAddr string, toAddr string, amount *big.Int, password string) map[string]interface{} {
	paraMap := make(map[string]interface{})
	paraMap["password"] = password
	paraMap["from"] = appendOxToAddress(fromAddr)
	paraMap["to"] = appendOxToAddress(toAddr)
	paraMap["value"] = "0x" + amount.Text(16)
	return paraMap
}

func makeSimpleTransGasEstimatedPara(fromAddr string, toAddr string, amount *big.Int) map[string]interface{} {
	//paraMap := make(map[string]interface{})
	//paraMap["from"] = fromAddr
	//paraMap["to"] = toAddr
	//paraMap["value"] = "0x" + amount.Text(16)
	return makeGasEstimatePara(fromAddr, toAddr, amount, "") //araMap
}

func makeERC20TokenTransData(contractAddr string, toAddr string, amount *big.Int) (string, error) {
	var funcParams []SolidityParam
	funcParams = append(funcParams, SolidityParam{
		ParamType:  SOLIDITY_TYPE_ADDRESS,
		ParamValue: toAddr,
	})

	funcParams = append(funcParams, SolidityParam{
		ParamType:  SOLIDITY_TYPE_UINT256,
		ParamValue: amount,
	})

	//fmt.Println("make token transfer data, amount:", amount.String())
	data, err := makeTransactionData(ETH_TRANSFER_TOKEN_BALANCE_METHOD, funcParams)
	if err != nil {
		log.Errorf("make transaction data failed, err = %v", err)
		return "", err
	}
	log.Debugf("data:%v", data)
	return data, nil
}

func makeGasEstimatePara(fromAddr string, toAddr string, value *big.Int, data string) map[string]interface{} {
	paraMap := make(map[string]interface{})
	paraMap["from"] = appendOxToAddress(fromAddr)
	paraMap["to"] = appendOxToAddress(toAddr)
	if data != "" {
		paraMap["data"] = data
	}

	if value != nil {
		paraMap["value"] = "0x" + value.Text(16)
	}
	return paraMap
}

func makeERC20TokenTransGasEstimatePara(fromAddr string, contractAddr string, data string) map[string]interface{} {

	//paraMap := make(map[string]interface{})

	//use password to unlock the account
	//use the following attr to eth_sendTransaction
	//paraMap["from"] = fromAddr //fromAddr.Address
	//paraMap["to"] = contractAddr
	//paraMap["value"] = "0x" + amount.Text(16)
	//paraMap["gas"] = "0x" + fee.GasLimit.Text(16)
	//paraMap["gasPrice"] = "0x" + fee.GasPrice.Text(16)
	//paraMap["data"] = data
	return makeGasEstimatePara(fromAddr, contractAddr, nil, data)
}

func (this *Client) ethGetGasEstimated(paraMap map[string]interface{}) (*big.Int, error) {
	trans := make(map[string]interface{})
	var temp interface{}
	var exist bool
	var fromAddr string
	var toAddr string

	if temp, exist = paraMap["from"]; !exist {
		log.Errorf("from not found")
		return big.NewInt(0), errors.New("from not found")
	} else {
		fromAddr = temp.(string)
		trans["from"] = fromAddr
	}

	if temp, exist = paraMap["to"]; !exist {
		log.Errorf("to not found")
		return big.NewInt(0), errors.New("to not found")
	} else {
		toAddr = temp.(string)
		trans["to"] = toAddr
	}

	if temp, exist = paraMap["value"]; exist {
		amount := temp.(string)
		trans["value"] = amount
	}

	if temp, exist = paraMap["data"]; exist {
		data := temp.(string)
		trans["data"] = data
	}

	params := []interface{}{
		trans,
	}

	result, err := this.Call("eth_estimateGas", 1, params)
	if err != nil {
		log.Errorf(fmt.Sprintf("get estimated gas limit from [%v] to [%v] faield, err = %v \n", fromAddr, toAddr, err))
		return big.NewInt(0), err
	}

	if result.Type != gjson.String {
		errInfo := fmt.Sprintf("get estimated gas from [%v] to [%v] result type error, result type is %v\n", fromAddr, toAddr, result.Type)
		log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}

	gasLimit, err := ConvertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert estimated gas[%v] format to bigint failed, err = %v\n", result.String(), err)
		log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return gasLimit, nil
}

func makeERC20TokenTransactionPara(fromAddr *Address, contractAddr string, data string,
	password string, fee *txFeeInfo) map[string]interface{} {

	paraMap := make(map[string]interface{})

	//use password to unlock the account
	paraMap["password"] = password
	//use the following attr to eth_sendTransaction
	paraMap["from"] = appendOxToAddress(fromAddr.Address)
	paraMap["to"] = appendOxToAddress(contractAddr)
	//paraMap["value"] = "0x" + amount.Text(16)
	paraMap["gas"] = "0x" + fee.GasLimit.Text(16)
	paraMap["gasPrice"] = "0x" + fee.GasPrice.Text(16)
	paraMap["data"] = data
	return paraMap
}

func (this *WalletManager) SendTransactionToAddr(param map[string]interface{}) (string, error) {
	//(addr *Address, to string, amount *big.Int, password string, fee *txFeeInfo) (string, error) {
	var exist bool
	var temp interface{}
	if temp, exist = param["from"]; !exist {
		log.Errorf("from not found.")
		return "", errors.New("from not found.")
	}

	fromAddr := temp.(string)

	if temp, exist = param["password"]; !exist {
		log.Errorf("password not found.")
		return "", errors.New("password not found.")
	}

	password := temp.(string)

	err := this.WalletClient.UnlockAddr(fromAddr, password, 300)
	if err != nil {
		log.Errorf("unlock addr failed, err = %v", err)
		return "", err
	}

	txId, err := this.WalletClient.ethSendTransaction(param)
	if err != nil {
		log.Errorf("ethSendTransaction failed, err = %v", err)
		return "", err
	}

	err = this.WalletClient.LockAddr(fromAddr)
	if err != nil {
		log.Errorf("lock addr failed, err = %v", err)
		return txId, err
	}

	return txId, nil
}

func (this *WalletManager) EthSendRawTransaction(signedTx string) (string, error) {
	return this.WalletClient.ethSendRawTransaction(signedTx)
}

func (this *Client) ethSendRawTransaction(signedTx string) (string, error) {
	params := []interface{}{
		signedTx,
	}

	result, err := this.Call("eth_sendRawTransaction", 1, params)
	if err != nil {
		log.Errorf(fmt.Sprintf("start raw transaction faield, err = %v \n", err))
		return "", err
	}

	if result.Type != gjson.String {
		log.Errorf("eth_sendRawTransaction result type error")
		return "", errors.New("eth_sendRawTransaction result type error")
	}
	return result.String(), nil
}

func (this *Client) ethSendTransaction(paraMap map[string]interface{}) (string, error) {
	//(fromAddr string, toAddr string, amount *big.Int, fee *txFeeInfo) (string, error) {
	trans := make(map[string]interface{})
	var temp interface{}
	var exist bool
	var fromAddr string
	var toAddr string

	if temp, exist = paraMap["from"]; !exist {
		log.Errorf("from not found")
		return "", errors.New("from not found")
	} else {
		fromAddr = temp.(string)
		trans["from"] = fromAddr
	}

	if temp, exist = paraMap["to"]; !exist {
		log.Errorf("to not found")
		return "", errors.New("to not found")
	} else {
		toAddr = temp.(string)
		trans["to"] = toAddr
	}

	if temp, exist = paraMap["value"]; exist {
		amount := temp.(string)
		trans["value"] = amount
	}

	if temp, exist = paraMap["gas"]; exist {
		gasLimit := temp.(string)
		trans["gas"] = gasLimit
	}

	if temp, exist = paraMap["gasPrice"]; exist {
		gasPrice := temp.(string)
		trans["gasPrice"] = gasPrice
	}

	if temp, exist = paraMap["data"]; exist {
		data := temp.(string)
		trans["data"] = data
	}

	params := []interface{}{
		trans,
	}

	result, err := this.Call("eth_sendTransaction", 1, params)
	if err != nil {
		log.Errorf(fmt.Sprintf("start transaction from [%v] to [%v] faield, err = %v \n", fromAddr, toAddr, err))
		return "", err
	}

	if result.Type != gjson.String {
		log.Errorf("eth_sendTransaction result type error")
		return "", errors.New("eth_sendTransaction result type error")
	}
	return result.String(), nil
}

func (this *Client) ethGetAccounts() ([]string, error) {
	param := make([]interface{}, 0)
	accounts := make([]string, 0)
	result, err := this.Call("eth_accounts", 1, param)
	if err != nil {
		log.Errorf("get eth accounts faield, err = %v \n", err)
		return nil, err
	}

	log.Debugf("result type of eth_accounts is %v", result.Type)

	accountList := result.Array()
	for i, _ := range accountList {
		acc := accountList[i].String()
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (this *Client) ethGetBlockNumber() (uint64, error) {
	param := make([]interface{}, 0)
	result, err := this.Call("eth_blockNumber", 1, param)
	if err != nil {
		log.Errorf("get block number faield, err = %v \n", err)
		return 0, err
	}

	if result.Type != gjson.String {
		log.Errorf("result of block number type error")
		return 0, errors.New("result of block number type error")
	}

	blockNum, err := ConvertToUint64(result.String(), 16)
	if err != nil {
		log.Errorf("parse block number to big.Int failed, err=%v", err)
		return 0, err
	}

	return blockNum, nil
}

func (c *Client) Call(method string, id int64, params []interface{}) (*gjson.Result, error) {
	authHeader := req.Header{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	body := make(map[string]interface{}, 0)
	body["jsonrpc"] = "2.0"
	body["id"] = id
	body["method"] = method
	body["params"] = params

	if c.Debug {
		log.Debug("Start Request API...")
	}

	r, err := req.Post(c.BaseURL, req.BodyJSON(&body), authHeader)

	if c.Debug {
		log.Debug("Request API Completed")
	}

	if c.Debug {
		log.Debugf("%+v\n", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return nil, err
	}

	result := resp.Get("result")

	return &result, nil
}

//isError 是否报错
func isError(result *gjson.Result) error {
	var (
		err error
	)

	if !result.Get("error").IsObject() {

		if !result.Get("result").Exists() {
			return errors.New("Response is empty! ")
		}

		return nil
	}

	errInfo := fmt.Sprintf("[%d]%s",
		result.Get("error.code").Int(),
		result.Get("error.message").String())
	err = errors.New(errInfo)

	return err
}
