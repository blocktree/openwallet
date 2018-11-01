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

package qtum

import (
	"errors"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/imroc/req"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"math/big"
	"net/http"
	"strings"
)

// Explorer是由bitpay的insight-API提供区块数据查询接口
// 具体接口说明查看https://github.com/bitpay/insight-api
type Explorer struct {
	BaseURL     string
	AccessToken string
	Debug       bool
	client      *req.Req
	//Client *req.Req
}

func NewExplorer(url string, debug bool) *Explorer {
	c := Explorer{
		BaseURL: url,
		//AccessToken: token,
		Debug: debug,
	}

	api := req.New()
	c.client = api

	return &c
}

// Call calls a remote procedure on another node, specified by the path.
func (b *Explorer) Call(path string, request interface{}, method string) (*gjson.Result, error) {

	if b.client == nil {
		return nil, errors.New("API url is not setup. ")
	}

	if b.Debug {
		log.Std.Debug("Start Request API...")
	}

	url := b.BaseURL + path

	r, err := b.client.Do(method, url, request)

	if b.Debug {
		log.Std.Debug("Request API Completed")
	}

	if b.Debug {
		log.Std.Debug("%+v", r)
	}

	if r.Response().StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", r.String())
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = b.isError(&resp)
	if err != nil {
		return nil, err
	}

	//result := resp.Get("result")

	return &resp, nil
}

//isError 是否报错
func (b *Explorer) isError(result *gjson.Result) error {
	var (
		err error
	)

	/*
		//failed 返回错误
		{
			"result": null,
			"error": {
				"code": -8,
				"message": "Block height out of range"
			},
			"id": "foo"
		}
	*/

	if !result.Get("error").Exists() {

		if !result.Exists() {
			return errors.New("Response is empty! ")
		}

		return nil
	}

	errInfo := fmt.Sprintf("[%d]%s",
		result.Get("status").Int(),
		result.Get("error").String())
	err = errors.New(errInfo)

	return err
}

//getBlockByExplorer 获取区块数据
func (wm *WalletManager) getBlockByExplorer(hash string) (*Block, error) {

	path := fmt.Sprintf("block/%s", hash)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return nil, err
	}

	return newBlockByExplorer(result), nil
}

//getBlockHashByExplorer 获取区块hash
func (wm *WalletManager) getBlockHashByExplorer(height uint64) (string, error) {

	path := fmt.Sprintf("block-index/%d", height)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return "", err
	}

	return result.Get("blockHash").String(), nil
}

//getBlockHeightByExplorer 获取区块链高度
func (wm *WalletManager) getBlockHeightByExplorer() (uint64, error) {

	path := "status"

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return 0, err
	}

	height := result.Get("info.blocks").Uint()

	return height, nil
}

//getTxIDsInMemPoolByExplorer 获取待处理的交易池中的交易单IDs
func (wm *WalletManager) getTxIDsInMemPoolByExplorer() ([]string, error) {

	return nil, fmt.Errorf("insight-api unsupport query mempool transactions")
}

//GetTransaction 获取交易单
func (wm *WalletManager) getTransactionByExplorer(txid string) (*Transaction, error) {

	path := fmt.Sprintf("tx/%s", txid)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return nil, err
	}

	tx := newTxByExplorer(result, wm.config.isTestNet)

	return tx, nil

}

//listUnspentByExplorer 获取未花交易
func (wm *WalletManager) listUnspentByExplorer(address ...string) ([]*Unspent, error) {

	var (
		utxos = make([]*Unspent, 0)
	)

	addrs := strings.Join(address, ",")

	request := req.Param{
		"addrs": addrs,
	}

	path := "addrs/utxo"

	result, err := wm.ExplorerClient.Call(path, request, "POST")
	if err != nil {
		return nil, err
	}

	array := result.Array()
	for _, a := range array {
		utxos = append(utxos, newUnspentByExplorer(&a))
	}

	return utxos, nil

}

func newUnspentByExplorer(json *gjson.Result) *Unspent {
	obj := &Unspent{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Vout = gjson.Get(json.Raw, "vout").Uint()
	obj.Address = gjson.Get(json.Raw, "address").String()
	obj.AccountID = gjson.Get(json.Raw, "account").String()
	obj.ScriptPubKey = gjson.Get(json.Raw, "scriptPubKey").String()
	obj.Amount = gjson.Get(json.Raw, "amount").String()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	isStake := gjson.Get(json.Raw, "isStake").Bool()
	if isStake {
		//挖矿的UTXO需要超过500个确认才能用
		if obj.Confirmations >= StakeConfirmations {
			obj.Spendable = true
		} else {
			obj.Spendable = false
		}
	} else {
		obj.Spendable = true
	}
	obj.Solvable = gjson.Get(json.Raw, "solvable").Bool()

	return obj
}

func newBlockByExplorer(json *gjson.Result) *Block {

	/*
		{
			"hash": "0000000000002bd2475d1baea1de4067ebb528523a8046d5f9d8ef1cb60460d3",
			"size": 549,
			"height": 1434016,
			"version": 536870912,
			"merkleroot": "ae4310c991ec16cfc7404aaad9fe5fbd533d0b6617c03eb1ac644c89d58b3e18",
			"tx": ["6767a8acc1a63c7978186c582fdea26c47da5e04b0b2b34740a1728bfd959a05", "226dee96373aedd8a3dd00021684b190b7f23f5e16bb186cee11d0560406c19d"],
			"time": 1539066282,
			"nonce": 4089837546,
			"bits": "1a3fffc0",
			"difficulty": 262144,
			"chainwork": "0000000000000000000000000000000000000000000000c6fce84fddeb57e5fb",
			"confirmations": 279,
			"previousblockhash": "0000000000001fdabb5efc93d15ccaf6980642918cd898df6b3ff5fbf26c19c4",
			"nextblockhash": "00000000000024f2bd323157e595613291f83485ddfbbf311323ed0c0dc46545",
			"reward": 0.78125,
			"isMainChain": true,
			"poolInfo": {}
		}
	*/
	obj := &Block{}
	//解析json
	obj.Hash = gjson.Get(json.Raw, "hash").String()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	obj.Merkleroot = gjson.Get(json.Raw, "merkleroot").String()

	txs := make([]string, 0)
	for _, tx := range gjson.Get(json.Raw, "tx").Array() {
		txs = append(txs, tx.String())
	}

	obj.tx = txs
	obj.Previousblockhash = gjson.Get(json.Raw, "previousblockhash").String()
	obj.Height = gjson.Get(json.Raw, "height").Uint()
	//obj.Version = gjson.Get(json.Raw, "version").String()
	obj.Time = gjson.Get(json.Raw, "time").Uint()

	return obj
}

func newTxByExplorer(json *gjson.Result, isTestnet bool) *Transaction {

	/*
			{
			"txid": "9f5eae5b95016825a437ceb9c9224d3e30d3b351f1100e4df5cc0cacac4e668c",
			"version": 1,
			"locktime": 1433760,
			"vin": [],
			"vout": [],
			"blockhash": "0000000000003ac968ee1ae321f35f76d4dcb685045968d60fc39edb20b0eed0",
			"blockheight": 1433761,
			"confirmations": 5,
			"time": 1539050096,
			"blocktime": 1539050096,
			"valueOut": 0.14652549,
			"size": 814,
			"valueIn": 0.14668889,
			"fees": 0.0001634
		}
	*/
	obj := Transaction{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Version = gjson.Get(json.Raw, "version").Uint()
	obj.LockTime = gjson.Get(json.Raw, "locktime").Int()
	obj.BlockHash = gjson.Get(json.Raw, "blockhash").String()
	obj.BlockHeight = gjson.Get(json.Raw, "blockheight").Uint()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	obj.Blocktime = gjson.Get(json.Raw, "blocktime").Int()
	obj.Size = gjson.Get(json.Raw, "size").Uint()
	obj.Fees = gjson.Get(json.Raw, "fees").String()

	obj.Vins = make([]*Vin, 0)
	if vins := gjson.Get(json.Raw, "vin"); vins.IsArray() {
		for _, vin := range vins.Array() {
			input := newTxVinByExplorer(&vin)
			obj.Vins = append(obj.Vins, input)
		}
	}

	obj.Vouts = make([]*Vout, 0)
	if vouts := gjson.Get(json.Raw, "vout"); vouts.IsArray() {
		for _, vout := range vouts.Array() {
			output := newTxVoutByExplorer(&vout, isTestnet)
			obj.Vouts = append(obj.Vouts, output)
		}
	}

	obj.Isqrc20Transfer = gjson.Get(json.Raw, "isqrc20Transfer").Bool()

	if obj.Isqrc20Transfer {
		obj.TokenReceipts = make([]*TokenReceipt, 0)
		if receipts := gjson.Get(json.Raw, "receipt"); receipts.IsArray() {
			for _, receipt := range receipts.Array() {
				token := newTokenReceiptByExplorer(&receipt, isTestnet)
				obj.TokenReceipts = append(obj.TokenReceipts, token)
			}
		}
	}

	return &obj
}

func newTxVinByExplorer(json *gjson.Result) *Vin {

	/*
		{
			"txid": "b8c00fff9208cb02f694666084fe0d65c471e92e45cdc3fb2e43af3a772e702d",
			"vout": 0,
			"sequence": 4294967294,
			"n": 0,
			"scriptSig": {
				"hex": "47304402201f77d18435931a6cb51b6dd183decf067f933e92647562f71a33e80988fbc8f6022012abe6824ffa70e5ccb7326e0dbb66144ba71133c1d4a1215da0b17358d7ca660121024d7be1242bd44619779a976cd1cd2d9351fcf58df59929b30a0c69d852302fb5",
				"asm": "304402201f77d18435931a6cb51b6dd183decf067f933e92647562f71a33e80988fbc8f6022012abe6824ffa70e5ccb7326e0dbb66144ba71133c1d4a1215da0b17358d7ca66[ALL] 024d7be1242bd44619779a976cd1cd2d9351fcf58df59929b30a0c69d852302fb5"
			},
			"addr": "msYiUQquCtGucnk3ZaWeJenYmY8WxRoeuv",
			"valueSat": 990000,
			"value": 0.0099,
			"doubleSpentTxID": null
		}
	*/
	obj := Vin{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Vout = gjson.Get(json.Raw, "vout").Uint()
	obj.N = gjson.Get(json.Raw, "n").Uint()
	obj.Addr = gjson.Get(json.Raw, "addr").String()
	obj.Value = gjson.Get(json.Raw, "value").String()
	obj.Coinbase = gjson.Get(json.Raw, "coinbase").String()

	return &obj
}

func newTxVoutByExplorer(json *gjson.Result, isTestnet bool) *Vout {

	/*
		{
			"value": "0.01652549",
			"n": 0,
			"scriptPubKey": {
				"hex": "76a9142760a760e8d22b5facb380444920e1197f272ea888ac",
				"asm": "OP_DUP OP_HASH160 2760a760e8d22b5facb380444920e1197f272ea8 OP_EQUALVERIFY OP_CHECKSIG",
				"addresses": ["mj7ASAGw8ia2o7Hqvo2XS1d7jGWr5UgEU9"],
				"type": "pubkeyhash"
			},
			"spentTxId": null,
			"spentIndex": null,
			"spentHeight": null
		}
	*/
	obj := Vout{}
	//解析json
	obj.Value = gjson.Get(json.Raw, "value").String()
	obj.N = gjson.Get(json.Raw, "n").Uint()
	obj.ScriptPubKey = gjson.Get(json.Raw, "scriptPubKey.hex").String()

	//提取地址
	if addresses := gjson.Get(json.Raw, "scriptPubKey.addresses"); addresses.IsArray() {
		obj.Addr = addresses.Array()[0].String()
	}

	obj.Type = gjson.Get(json.Raw, "scriptPubKey.type").String()

	//if len(obj.Addr) == 0 {
	//	scriptBytes, _ := hex.DecodeString(obj.ScriptPubKey)
	//	obj.Addr, _ = ScriptPubKeyToBech32Address(scriptBytes, isTestnet)
	//}

	return &obj
}

//newTokenReceiptByExplorer
func newTokenReceiptByExplorer(json *gjson.Result, isTestnet bool) *TokenReceipt {

	/*
			"receipt": [
		        {
		            "blockHash": "35d196cbd08cf7dcce08d99bb7267150c7ce328f08e0f66f706267cd75ab0d55",
		            "blockNumber": 249878,
		            "transactionHash": "eb8e496f7dd23554d6d45de30beab384c8e0d023c9c7f1fbc15d90d10bb873f8",
		            "transactionIndex": 17,
		            "from": "a20a4eec5c83fb9b61a9efc7fe6c0e06bb3dde43",
		            "to": "f2033ede578e17fa6231047265010445bca8cf1c",
		            "cumulativeGasUsed": 87782,
		            "gasUsed": 36423,
		            "contractAddress": "f2033ede578e17fa6231047265010445bca8cf1c",
		            "excepted": "None",
		            "log": [
		                {
		                    "address": "f2033ede578e17fa6231047265010445bca8cf1c",
		                    "topics": [
		                        "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		                        "000000000000000000000000a20a4eec5c83fb9b61a9efc7fe6c0e06bb3dde43",
		                        "000000000000000000000000e57e4a5f9ac130defb33a057729f10728fcdb9cb"
		                    ],
		                    "data": "0000000000000000000000000000000000000000000000000000034f80e83000"
		                }
		            ]
		        }
		    ],
		    "isqrc20Transfer": true,
	*/

	obj := TokenReceipt{}
	//解析json
	obj.BlockHash = gjson.Get(json.Raw, "blockHash").String()
	obj.BlockHeight = gjson.Get(json.Raw, "blockNumber").Uint()
	obj.TxHash = gjson.Get(json.Raw, "transactionHash").String()
	obj.Excepted = gjson.Get(json.Raw, "excepted").String()
	obj.GasUsed = gjson.Get(json.Raw, "gasUsed").Uint()
	obj.ContractAddress = "0x" + gjson.Get(json.Raw, "contractAddress").String()
	obj.Sender = HashAddressToBaseAddress(
		gjson.Get(json.Raw, "from").String(),
		isTestnet)

	logs := gjson.Get(json.Raw, "log").Array()
	for _, logInfo := range logs {
		topics := logInfo.Get("topics").Array()
		data := logInfo.Get("data").String()

		if len(topics) != 3 {
			continue
		}

		if "0x" + topics[0].String() != QTUM_TRANSFER_EVENT_ID {
			continue
		}

		if len(data) != 64 {
			continue
		}

		//log.Info("topics[1]:", topics[1].String())
		//log.Info("topics[2]:", topics[2].String())
		obj.From = strings.TrimPrefix(topics[1].String(), "000000000000000000000000")
		obj.To = strings.TrimPrefix(topics[2].String(), "000000000000000000000000")
		obj.From = HashAddressToBaseAddress(obj.From, isTestnet)
		obj.To = HashAddressToBaseAddress(obj.To, isTestnet)

		//转化为10进制
		//log.Debug("TokenReceipt TxHash", obj.TxHash)
		//log.Debug("TokenReceipt amount", logInfo.Get("data").String())
		value := new(big.Int)
		value, _ = value.SetString(data, 16)
		obj.Amount = decimal.NewFromBigInt(value, 0).String()
	}

	return &obj
}

//getBalanceByExplorer 获取地址余额
func (wm *WalletManager) getBalanceByExplorer(address string) (*openwallet.Balance, error) {

	path := fmt.Sprintf("addr/%s?noTxList=1", address)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return nil, err
	}

	return newBalanceByExplorer(result), nil
}

func newBalanceByExplorer(json *gjson.Result) *openwallet.Balance {

	/*

		{
			"addrStr": "mnMSQs3HZ5zhJrCEKbqGvcDLjAAxvDJDCd",
			"balance": 3136.82244887,
			"balanceSat": 313682244887,
			"totalReceived": 3136.82244887,
			"totalReceivedSat": 313682244887,
			"totalSent": 0,
			"totalSentSat": 0,
			"unconfirmedBalance": 0,
			"unconfirmedBalanceSat": 0,
			"unconfirmedTxApperances": 0,
			"txApperances": 3909
		}

	*/
	obj := openwallet.Balance{}
	//解析json
	obj.Address = gjson.Get(json.Raw, "addrStr").String()
	obj.Balance = gjson.Get(json.Raw, "balance").String()
	obj.UnconfirmBalance = gjson.Get(json.Raw, "unconfirmedBalance").String()
	u, _ := decimal.NewFromString(obj.UnconfirmBalance)
	b, _ := decimal.NewFromString(obj.UnconfirmBalance)
	obj.ConfirmBalance = b.Sub(u).StringFixed(8)

	return &obj
}

//getMultiAddrTransactionsByExplorer 获取多个地址的交易单数组
func (wm *WalletManager) getMultiAddrTransactionsByExplorer(offset, limit int, address ...string) ([]*Transaction, error) {

	var (
		trxs = make([]*Transaction, 0)
	)

	addrs := strings.Join(address, ",")

	request := req.Param{
		"addrs": addrs,
		"from":  offset,
		"to":    offset + limit,
	}

	path := fmt.Sprintf("addrs/txs")

	result, err := wm.ExplorerClient.Call(path, request, "POST")
	if err != nil {
		return nil, err
	}

	if items := result.Get("items"); items.IsArray() {
		for _, obj := range items.Array() {
			tx := newTxByExplorer(&obj, wm.config.isTestNet)
			trxs = append(trxs, tx)
		}
	}

	return trxs, nil
}

//estimateFeeRateByExplorer 通过浏览器获取费率
func (wm *WalletManager) estimateFeeRateByExplorer() (decimal.Decimal, error) {

	defaultRate, _ := decimal.NewFromString("0.004")

	path := fmt.Sprintf("utils/estimatefee?nbBlocks=%d", 2)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return decimal.New(0, 0), err
	}

	feeRate, _ := decimal.NewFromString(result.Get("2").String())

	if feeRate.LessThan(defaultRate) {
		feeRate = defaultRate
	}

	return feeRate, nil
}

//getTxOutByExplorer 获取交易单输出信息，用于追溯交易单输入源头
func (wm *WalletManager) getTxOutByExplorer(txid string, vout uint64) (*Vout, error) {

	tx, err := wm.getTransactionByExplorer(txid)
	if err != nil {
		return nil, err
	}

	for i, out := range tx.Vouts {
		if uint64(i) == vout {
			return out, nil
		}
	}

	return nil, fmt.Errorf("can not find ouput")

}

//sendRawTransactionByExplorer 广播交易
func (wm *WalletManager) sendRawTransactionByExplorer(txHex string) (string, error) {

	request := req.Param{
		"rawtx": txHex,
	}

	path := fmt.Sprintf("tx/send")

	result, err := wm.ExplorerClient.Call(path, request, "POST")
	if err != nil {
		return "", err
	}

	return result.Get("txid").String(), nil

}

//getAddressTokenBalanceByExplorer 通过合约地址查询用户地址的余额
func (wm *WalletManager) getAddressTokenBalanceByExplorer(token openwallet.SmartContract, address string) (decimal.Decimal, error) {

	trimContractAddr := strings.TrimPrefix(token.Address, "0x")

	tokenAddressBase := HashAddressToBaseAddress(trimContractAddr, wm.config.isTestNet)

	path := fmt.Sprintf("tokens/%s/addresses/%s/balance?format=object", tokenAddressBase, address)

	result, err := wm.ExplorerClient.Call(path, nil, "GET")
	if err != nil {
		return decimal.New(0, 0), err
	}

	balanceStr := result.Get("balance").String()

	balance, _ := decimal.NewFromString(balanceStr)
	decimals := decimal.New(1, int32(token.Decimals))

	balance = balance.Div(decimals)
	return balance, nil

}
