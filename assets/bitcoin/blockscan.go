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

package bitcoin

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"path/filepath"
)

/*
	步骤：
	1.添加需要扫块的钱包，及传入初始高度，-1为本地高度。
	2.获取已扫描的本地高度。
	3.获取高度+1的区块hash，通过区块链hash获取区块链数据，获取mempool数据。
	4.判断区块链的父区块hash是否与本地上一区块hash一致。
	5.解析新区块链的交易单数组。
	6.遍历交易单结构，检查每个output地址是否存在钱包的地址表中
	7.检查地址是否合法，存在地址表，生成充值记录。
	8.定时程推送充值记录到钱包的充值通道。先检查交易hash是否存在区块中。
	9.接口返回确认，标记充值记录已确认。
*/

const (
	//区块链数据集合
	blockchainBucket = "blockchain"
)

type BTCBlockScanner struct {
	addressInScanning  map[string]string //加入扫描的钱包资产账户
	CurrentBlockHeight uint64            //当前区块高度
	isScanning         bool
}

//exportTRXS 导出交易单回调函数
//@param  txs 每扫完区块链，与地址相关的交易到
type exportTRXS func(txs []*openwallet.Transaction) []string

//NewBTCBlockScanner 创建区块链扫描器
func NewBTCBlockScanner() *BTCBlockScanner {
	bs := BTCBlockScanner{}
	bs.addressInScanning = make(map[string]string)
	return &bs
}

func (bs *BTCBlockScanner) AddAddress(address string, accountID string) {
	bs.addressInScanning[address] = accountID
}

func (bs *BTCBlockScanner) Start() {

}

func (bs *BTCBlockScanner) Scanning() {
	//获取本地区块高度

}

//GetCurrentBlockHeight 获取当前区块高度
func (bs *BTCBlockScanner) GetCurrentBlockHeight() (uint64, error) {

	var (
		blockHeight uint64 = 0
		err         error
	)

	blockHeight = GetLocalBlockHeight()

	//如果本地没有记录，查询接口的高度
	if blockHeight == 0 {
		blockHeight, err = GetBlockHeight()
		if err != nil {
			return 0, err
		}
	}

	return blockHeight, nil
}

//extractRechargeRecords 从交易单中提取充值记录
func (bs *BTCBlockScanner) extractRechargeRecords(blockHeight uint64, txid string) error {

	var (
		transaction openwallet.Recharge
		saved       bool = true
		err         error
		trx         *gjson.Result
	)

	trx, err = GetTransaction(txid)
	if err != nil {
		//记录哪个区块哪个交易单没有完成扫描
		saved = false
	} else {

		blockhash := trx.Get("blockhash").String()
		confirmations := trx.Get("confirmations").Int()
		vout := trx.Get("vout")

		for _, output := range vout.Array() {

			amount := output.Get("value").String()
			n := output.Get("n").Uint()
			addresses := output.Get("scriptPubKey.addresses").Array()
			if len(addresses) > 0 {
				transaction = openwallet.Recharge{}
				transaction.TxID = txid
				transaction.Address = addresses[0].String()
				transaction.Confirm = confirmations
				transaction.BlockHash = blockhash
				transaction.Amount = amount
				transaction.BlockHeight = blockHeight
				transaction.Symbol = Symbol
				transaction.Index = n
				transaction.Sid = string(crypto.SHA256([]byte(fmt.Sprintf("%s_%d_%s", txid, n, addresses[0].String()))))
				//写入数据库地址相关的钱包数据库
				err = bs.SaveTxToWalletDB(&transaction)
				if err != nil {
					saved = false
				} else {
					saved = true
				}
			}

		}

	}

	//保存不成功加入到重扫表中
	if !saved {
		//

		unscanRecord := UnscanRecords{
			BlockHeight: blockHeight,
			TxID:        txid,
			Reason:      err.Error(),
		}
		bs.SaveUnscanRecord(&unscanRecord)
	} else {
		err = nil
	}

	return err
}

//SaveTxToWalletDB 保存交易记录到钱包数据库
func (bs *BTCBlockScanner) SaveTxToWalletDB(tx *openwallet.Recharge) error {

	if tx == nil {
		return errors.New("the transaction to save is nil")
	}

	accountID := bs.addressInScanning[tx.Address]
	tx.AccountID = accountID
	wallet, err := GetWalletInfo(accountID)
	if err != nil {
		return err
	}
	return wallet.SaveTx(tx)
}

//SaveTxToWalletDB 保存交易记录到钱包数据库
func (bs *BTCBlockScanner) SaveUnscanRecord(record *UnscanRecords) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//GetBlockHeight 获取区块链高度
func GetBlockHeight() (uint64, error) {

	result, err := client.Call("getblockcount", nil)
	if err != nil {
		return 0, err
	}

	return result.Uint(), nil
}

//GetLocalBlockHeight 获取本地记录的区块链高度
func GetLocalBlockHeight() uint64 {

	var (
		blockHeight uint64 = 0
	)

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return 0
	}
	defer db.Close()

	db.Get(blockchainBucket, "blockHeight", &blockHeight)

	return blockHeight
}

//SaveLocalBlockHeight 记录高度到本地
func SaveLocalBlockHeight(blockHeight uint64) {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
}

//SaveTransaction 记录高度到本地
func SaveTransaction(blockHeight uint64) {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(dbPath, blockchainFile))
	if err != nil {
		return
	}
	defer db.Close()

	db.Set(blockchainBucket, "blockHeight", &blockHeight)
}

//GetBlockHash 根据区块高度获得区块hash
func GetBlockHash(height uint64) (string, error) {

	request := []interface{}{
		height,
	}

	result, err := client.Call("getblockhash", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

//GetBlock 获取区块数据
func GetBlock(hash string) (*Block, error) {

	request := []interface{}{
		hash,
	}

	result, err := client.Call("getblock", request)
	if err != nil {
		return nil, err
	}

	return NewBlock(result), nil
}

//GetTransaction 获取交易单
func GetTransaction(txid string) (*gjson.Result, error) {

	request := []interface{}{
		txid,
		true,
	}

	result, err := client.Call("getrawtransaction", request)
	if err != nil {
		return nil, err
	}

	return result, nil

}
