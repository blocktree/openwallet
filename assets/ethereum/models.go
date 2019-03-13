/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */
package ethereum

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strconv"
	"time"

	"github.com/asdine/storm"
	"github.com/blocktree/openwallet/common/file"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/blocktree/go-owcrypt"
	"github.com/bytom/common"
)

type Wallet struct {
	WalletID     string   `json:"rootid" storm:"id"`
	Alias        string   `json:"alias"`
	balance      *big.Int //string `json:"balance"`
	erc20Token   *ERC20Token
	Password     string `json:"password"`
	RootPub      string `json:"rootpub"`
	RootPath     string
	KeyFile      string
	HdPath       string
	PublicKey    string
	AddressCount uint64
}

type ERC20Token struct {
	Address  string `json:"address" storm:"id"`
	Symbol   string `json:"symbol" storm:"index"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	balance  *big.Int
}

type EthEvent struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `josn:"data"`
	//BlockNumber string
	LogIndex string `json:"logIndex"`
	Removed  bool   `json:"removed"`
}

type EthTransactionReceipt struct {
	Logs    []EthEvent `json:"logs"`
	GasUsed string     `json:"gasUsed"`
	Status  string     `json:"status"`
}

type TransferEvent struct {
	TokenFrom     string
	TokenTo       string
	Value         string
	FromSourceKey string //transaction scanning 的时候对其进行赋值
	ToSourceKey   string //transaction scanning 的时候对其进行赋值
}

func (this *EthTransactionReceipt) ParseTransferEvent() *TransferEvent {
	var transferEvent TransferEvent
	removePrefix0 := func(num string) string {
		num = removeOxFromHex(num)
		array := []byte(num)
		i := 0

		for i, _ = range num {
			if num[i] != '0' {
				break
			}
		}

		return string(array[i:len(num)])
	}

	for i, _ := range this.Logs {
		if len(this.Logs[i].Topics) != 3 {
			continue
		}

		if this.Logs[i].Topics[0] != ETH_TRANSFER_EVENT_ID {
			continue
		}

		if len(this.Logs[i].Data) != 66 {
			continue
		}

		prefix := string([]byte(this.Logs[i].Topics[1])[0:26:26])
		if prefix != "0x000000000000000000000000" {
			continue
		}

		prefix = string([]byte(this.Logs[i].Topics[2])[0:26:26])
		if prefix != "0x000000000000000000000000" {
			continue
		}

		transferEvent.TokenFrom = "0x" + string([]byte(this.Logs[i].Topics[1])[26:66:66])
		transferEvent.TokenTo = "0x" + string([]byte(this.Logs[i].Topics[2])[26:66:66])
		transferEvent.Value = "0x" + removePrefix0(this.Logs[i].Data)
		return &transferEvent
	}
	return nil
}

type Address struct {
	Address      string `json:"address" storm:"id"`
	Account      string `json:"account" storm:"index"`
	HDPath       string `json:"hdpath"`
	Index        int
	PublicKey    string
	balance      *big.Int //string `json:"balance"`
	tokenBalance *big.Int
	TxCount      uint64
	CreatedAt    time.Time
}

func (this *Address) CalcPrivKey(masterKey *hdkeystore.HDKey) ([]byte, error) {
	childKey, _ := masterKey.DerivedKeyWithPath(this.HDPath, owcrypt.ECC_CURVE_SECP256K1)
	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		log.Error("get private key bytes, err=", err)
		return nil, err
	}
	return keyBytes, nil
}

func (this *Address) CalcHexPrivKey(masterKey *hdkeystore.HDKey) (string, error) {
	prikey, err := this.CalcPrivKey(masterKey)
	if err != nil {
		return "", err
	}
	return common.ToHex(prikey), nil
}

type UnscanTransaction struct {
	TxID        string `storm:"id"` // primary key
	BlockNumber string `storm:"index"`
	BlockHash   string `storm:"index"`
	//	TxID        string
	TxSpec string
	Reason string
}

type BlockTransaction struct {
	Hash             string `json:"hash" storm:"id"`
	BlockNumber      string `json:"blockNumber" storm:"index"`
	BlockHash        string `json:"blockHash" storm:"index"`
	From             string `json:"from"`
	To               string `json:"to"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Value            string `json:"value"`
	Data             string `json:"input"`
	TransactionIndex string `json:"transactionIndex"`
	Timestamp        string `json:"timestamp"`
	BlockHeight      uint64 //transaction scanning 的时候对其进行赋值
	filterFunc       openwallet.BlockScanAddressFunc
	Status           uint64
}

func (this *BlockTransaction) GetAmountEthString() (string, error) {
	amount, err := ConvertToBigInt(this.Value, 16)
	if err != nil {
		log.Errorf("convert amount to big.int failed, err= %v", err)
		return "", err
	}
	amountVal, err := ConverWeiStringToEthDecimal(amount.String())
	if err != nil {
		log.Errorf("convert tx.Amount to eth decimal failed, err=%v", err)
		return "", err
	}
	return amountVal.String(), nil
}

func (this *BlockTransaction) GetTxFeeEthString() (string, error) {
	gasPrice, err := ConvertToBigInt(this.GasPrice, 16)
	if err != nil {
		log.Errorf("convert tx.GasPrice failed, err= %v", err)
		return "", err
	}

	gas, err := ConvertToBigInt(this.Gas, 16)
	if err != nil {
		log.Errorf("convert tx.Gas failed, err=%v", err)
		return "", err
	}
	fee := big.NewInt(0)
	fee.Mul(gasPrice, gas)
	feeprice, err := ConverWeiStringToEthDecimal(fee.String())
	if err != nil {
		log.Errorf("convert fee failed, err=%v", err)
		return "", err
	}
	return feeprice.String(), nil
}

type BlockHeader struct {
	BlockNumber     string `json:"number" storm:"id"`
	BlockHash       string `json:"hash"`
	GasLimit        string `json:"gasLimit"`
	GasUsed         string `json:"gasUsed"`
	Miner           string `json:"miner"`
	Difficulty      string `json:"difficulty"`
	TotalDifficulty string `json:"totalDifficulty"`
	PreviousHash    string `json:"parentHash"`
	blockHeight     uint64 //RecoverBlockHeader的时候进行初始化
}

func (this *Wallet) SaveAddress(dbpath string, addr *Address) error {
	db, err := this.OpenDB(dbpath)
	if err != nil {
		log.Errorf("open db failed, err = %v", err)
		return err
	}
	defer db.Close()

	return db.Save(addr)
}

func (this *Wallet) ClearAllTransactions(dbPath string) {
	db, err := this.OpenDB(dbPath)
	if err != nil {
		log.Errorf("open db failed, err = %v", err)
		return
	}
	defer db.Close()

	var txs []BlockTransaction
	err = db.All(&txs)
	if err != nil {
		log.Errorf("get transactions failed, err = %v", err)
		return
	}
	for i, _ := range txs {
		//fmt.Println("BlockHash:", txs[i].BlockHash, " BlockNumber:", txs[i].BlockNumber, "TransactionId:", txs[i].Hash)
		err := db.DeleteStruct(&txs[i])
		if err != nil {
			log.Errorf("delete tx in wallet failed, err=%v", err)
			break
		}
	}

}

func (this *Wallet) RestoreFromDb(dbPath string) error {
	db, err := this.OpenDB(dbPath)
	if err != nil {
		log.Errorf("open db failed, err = %v", err)
		return err
	}
	defer db.Close()

	var w Wallet
	err = db.One("WalletID", this.WalletID, &w)
	if err != nil {
		log.Error("find wallet id[", this.WalletID, "] failed, err=", err)
		return err
	}

	wstr, _ := json.MarshalIndent(w, "", " ")
	log.Debugf("wallet:%v", string(wstr))
	*this = w
	return nil
}

func (this *Wallet) DumpWalletDB(dbPath string) {
	db, err := this.OpenDB(dbPath)
	if err != nil {
		log.Errorf("open db failed, err = %v", err)
		return
	}
	defer db.Close()

	var addresses []Address
	err = db.All(&addresses)
	if err != nil {
		log.Errorf("get address failed, err=%v", err)
		return
	}

	for i, _ := range addresses {
		fmt.Println("Address:", addresses[i].Address, " account:", addresses[i].Account, "hdpath:", addresses[i].HDPath)
	}

	var txs []BlockTransaction
	err = db.All(&txs)
	if err != nil {
		log.Errorf("get transactions failed, err = %v", err)
		return
	}

	for i, _ := range txs {
		//fmt.Println("BlockHash:", txs[i].BlockHash, " BlockNumber:", txs[i].BlockNumber, "TransactionId:", txs[i].Hash),
		fmt.Printf("print tx[%v] in block [%v] = %v\n", txs[i].Hash, txs[i].BlockNumber, txs[i])
	}
}

func (this *WalletManager) ClearBlockScanDb() {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db failed, err = %v", err)
		return
	}
	defer db.Close()

	var unscanTransactions []UnscanTransaction
	var blocks []BlockHeader

	err = db.All(&unscanTransactions)
	if err != nil {
		this.Log.Errorf("get transactions failed, err = %v", err)
		return
	}

	for i, _ := range unscanTransactions {
		err = db.DeleteStruct(&unscanTransactions[i])
		if err != nil {
			this.Log.Errorf("delete transaction failed, err=%v", err)
			return
		}
	}

	err = db.All(&blocks)
	if err != nil {
		this.Log.Errorf("get blocks failed failed, err = %v", err)
		return
	}

	for i, _ := range blocks {
		err = db.DeleteStruct(&blocks[i])
		if err != nil {
			this.Log.Errorf("delete blocks failed, err=%v", err)
			return
		}
	}
}

//func (this *WalletManager) DumpBlockScanDb() {
//	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
//	if err != nil {
//		this.Log.Errorf("open db failed, err = %v", err)
//		return
//	}
//	defer db.Close()
//
//	var unscanTransactions []UnscanTransaction
//	var blocks []BlockHeader
//	var blockHeightStr string
//	err = db.All(&unscanTransactions)
//	if err != nil {
//		this.Log.Errorf("get transactions failed, err = %v", err)
//		return
//	}
//
//	for i, _ := range unscanTransactions {
//		fmt.Printf("Print unscanned transaction [%v] = %v\n", unscanTransactions[i].TxID, unscanTransactions[i])
//	}
//
//	err = db.All(&blocks)
//	if err != nil {
//		this.Log.Errorf("get blocks failed failed, err = %v", err)
//		return
//	}
//
//	for i, _ := range blocks {
//		fmt.Printf("print block [%v] = %v\n", blocks[i].BlockNumber, blocks[i])
//	}
//
//	err = db.Get(BLOCK_CHAIN_BUCKET, "BlockNumber", &blockHeightStr)
//	if err != nil {
//		this.Log.Errorf("get block height from db failed, err=%v", err)
//		return
//	}
//
//	fmt.Println("print block number = ", blockHeightStr)
//}

func (this *Wallet) SaveTransactions(dbPath string, txs []BlockTransaction) error {
	db, err := this.OpenDB(dbPath)
	if err != nil {
		log.Errorf("open db failed, err = %v", err)
		return err
	}
	defer db.Close()

	dbTx, err := db.Begin(true)
	if err != nil {
		log.Errorf("start transaction for db failed, err=%v", err)
		return err
	}
	defer dbTx.Rollback()

	for i, _ := range txs {
		err = dbTx.Save(&txs[i])
		if err != nil {
			log.Errorf("save transaction failed, err=%v", err)
			return err
		}
	}
	dbTx.Commit()
	return nil
}

func (this *Wallet) DeleteTransactionByHeight(dbPath string, height uint64) error {
	db, err := this.OpenDB(dbPath)
	if err != nil {
		log.Errorf("open db for delete txs failed, err = %v", err)
		return err
	}
	defer db.Close()

	var txs []BlockTransaction

	err = db.Find("BlockNumber", "0x"+strconv.FormatUint(height, 16), &txs)
	if err != nil && err != storm.ErrNotFound {
		log.Errorf("get transactions from block[%v] failed, err=%v", "0x"+strconv.FormatUint(height, 16), err)
		return err
	} else if err == storm.ErrNotFound {
		log.Infof("no transactions found in block[%v] ", "0x"+strconv.FormatUint(height, 16))
		return nil
	}

	txdb, err := db.Begin(true)
	if err != nil {
		log.Errorf("start dbtx for delete tx failed, err=%v", err)
		return err
	}
	defer txdb.Rollback()

	for i, _ := range txs {
		err = txdb.DeleteStruct(&txs[i])
		if err != nil {
			log.Errorf("delete tx[%v] failed, err=%v", txs[i].Hash, err)
			return err
		}
	}
	txdb.Commit()
	return nil
}

//HDKey 获取钱包密钥，需要密码
func (this *Wallet) HDKey2(password string) (*hdkeystore.HDKey, error) {

	if len(password) == 0 {
		log.Error("password of wallet empty.")
		return nil, fmt.Errorf("password is empty")
	}

	if len(this.KeyFile) == 0 {
		log.Error("keyfile empty in wallet.")
		return nil, errors.New("Wallet key is not exist!")
	}

	keyjson, err := ioutil.ReadFile(this.KeyFile)
	if err != nil {
		return nil, err
	}
	key, err := hdkeystore.DecryptHDKey(keyjson, password)
	if err != nil {
		return nil, err
	}
	return key, err
}

//openDB 打开钱包数据库
func (w *Wallet) OpenDB(dbPath string) (*storm.DB, error) {
	file.MkdirAll(dbPath)
	file := w.DBFile(dbPath)
	//	fmt.Println("dbpath:", dbPath, ", file:", file)
	return storm.Open(file)
}

func (w *Wallet) OpenDbByPath(path string) (*storm.DB, error) {
	return storm.Open(path)
}

//DBFile 数据库文件
func (w *Wallet) DBFile(dbPath string) string {
	return filepath.Join(dbPath, w.FileName()+".db")
}

//FileName 该钱包定义的文件名规则
func (w *Wallet) FileName() string {
	return w.Alias + "-" + w.WalletID
}

func OpenDB(dbPath string, dbName string) (*storm.DB, error) {
	file.MkdirAll(dbPath)
	//	fmt.Println("OpenDB dbpath:", dbPath+"/"+dbName)
	return storm.Open(dbPath + "/" + dbName)
}
