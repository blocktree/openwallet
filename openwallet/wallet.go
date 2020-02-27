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

package openwallet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/blocktree/openwallet/common/file"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/blocktree/openwallet/log"
	"github.com/pkg/errors"
	"time"
)

//WalletDAI 钱包数据访问接口
type WalletDAI interface {
	//获取当前钱包
	GetWallet() *Wallet
	//根据walletID查询钱包
	GetWalletByID(walletID string) (*Wallet, error)

	//获取单个资产账户
	GetAssetsAccountInfo(accountID string) (*AssetsAccount, error)
	//查询资产账户列表
	GetAssetsAccountList(offset, limit int, cols ...interface{}) ([]*AssetsAccount, error)
	//根据地址查询资产账户
	GetAssetsAccountByAddress(address string) (*AssetsAccount, error)

	//获取单个地址
	GetAddress(address string) (*Address, error)
	//查询地址列表
	GetAddressList(offset, limit int, cols ...interface{}) ([]*Address, error)
	//设置地址的扩展字段
	SetAddressExtParam(address string, key string, val interface{}) error
	//获取地址的扩展字段
	GetAddressExtParam(address string, key string) (interface{}, error)

	//解锁钱包，指定时间内免密
	UnlockWallet(password string, time time.Duration) error
	//获取钱包HDKey
	HDKey(password ...string) (*hdkeystore.HDKey, error)

	//获取钱包所创建的交易单
	GetTransactionByTxID(txid, symbol string) ([]*Transaction, error)

}

//TransactionDecoderBase 实现TransactionDecoder的基类
type WalletDAIBase struct {
}

func (base *WalletDAIBase) GetWallet() *Wallet {
	return nil
}

func (base *WalletDAIBase) GetWalletByID(walletID string) (*Wallet, error) {
	return nil, fmt.Errorf("GetWalletByID not implement")
}

func (base *WalletDAIBase) GetAssetsAccountInfo(accountID string) (*AssetsAccount, error) {
	return nil, fmt.Errorf("GetAssetsAccountInfo not implement")
}

func (base *WalletDAIBase) GetAssetsAccountList(offset, limit int, cols ...interface{}) ([]*AssetsAccount, error) {
	return nil, fmt.Errorf("GetAssetsAccountList not implement")
}

func (base *WalletDAIBase) GetAssetsAccountByAddress(address string) (*AssetsAccount, error) {
	return nil, fmt.Errorf("GetAssetsAccountByAddress not implement")
}

func (base *WalletDAIBase) GetAddress(address string) (*Address, error) {
	return nil, fmt.Errorf("GetAddress not implement")
}

func (base *WalletDAIBase) GetAddressList(offset, limit int, cols ...interface{}) ([]*Address, error) {
	return nil, fmt.Errorf("GetAddressList not implement")
}

//设置地址的扩展字段
func (base *WalletDAIBase) SetAddressExtParam(address string, key string, val interface{}) error {
	return fmt.Errorf("SetAddressExtParam not implement")
}

//获取地址的扩展字段
func (base *WalletDAIBase) GetAddressExtParam(address string, key string) (interface{}, error) {
	return nil, fmt.Errorf("GetAddressExtParam not implement")
}

func (base *WalletDAIBase) UnlockWallet(password string, time time.Duration) error {
	return fmt.Errorf("UnlockWallet not implement")
}

func (base *WalletDAIBase) HDKey(password ...string) (*hdkeystore.HDKey, error) {
	return nil, fmt.Errorf("HDKey not implement")
}

//获取钱包所创建的交易单
func (base *WalletDAIBase) GetTransactionByTxID(txid, symbol string) ([]*Transaction, error) {
	return nil, fmt.Errorf("GetTransactionByTxID not implement")
}



type Wallet struct {
	AppID        string              `json:"appID"`
	WalletID     string              `json:"walletID"  storm:"id"`
	Alias        string              `json:"alias"`
	Password     string              `json:"password"`
	RootPub      string              `json:"rootpub"` //弃用
	RootPath     string              `json:"rootPath"`
	KeyFile      string              `json:"keyFile"`      //钱包的密钥文件
	DBFile       string              `json:"dbFile"`       //钱包的数据库文件
	WatchOnly    bool                `json:"watchOnly"`    //创建watchonly的钱包，没有私钥文件，只有db文件
	IsTrust      bool                `json:"isTrust"`      //是否托管密钥
	AccountIndex int                 `json:"accountIndex"` //账户索引数，-1代表未创建账户
	ExtParam     string              `json:"extParam"`     //扩展参数，用于调用智能合约，json结构
	key          *hdkeystore.HDKey   //Deprecated
	fileName     string              //钱包文件命名，所有与钱包相关的都以这个filename命名
	core         interface{}         //核心钱包指针 Deprecated
	unlocked     map[string]unlocked // 已解锁的钱包，集合（钱包地址, 钱包私钥）Deprecated
}

//Deprecated
func NewWallet(walletID string, symbol string) *Wallet {

	dbDir := GetDBDir(symbol)
	keyDir := GetKeyDir(symbol)

	file.MkdirAll(dbDir)
	file.MkdirAll(keyDir)

	//检查目录是否已存在钱包私钥文件，有则用私钥创建这个钱包
	wallets, err := GetWalletsByKeyDir(keyDir)
	if err != nil {
		return nil
	}

	for _, w := range wallets {
		if w.WalletID == walletID {
			w.KeyFile = filepath.Join(dbDir, w.FileName()+".key")
			w.DBFile = filepath.Join(dbDir, w.FileName()+".db")
			return w
		}
	}

	watchOnlyWallet := NewWatchOnlyWallet(walletID, symbol)

	return watchOnlyWallet

}

//NewWatchOnlyWallet 只读钱包，用于观察冷钱包 Deprecated
func NewWatchOnlyWallet(walletID string, symbol string) *Wallet {

	dbDir := GetDBDir(symbol)
	file.MkdirAll(dbDir)

	dbFile := filepath.Join(dbDir, walletID+".db")

	w := Wallet{
		WalletID:  walletID,
		Alias:     walletID, //自定义ID也作为别名
		WatchOnly: true,
		DBFile:    dbFile,
		fileName:  walletID,
	}

	return &w
}

//HDKey 获取钱包密钥，需要密码
func (w *Wallet) HDKey(password ...string) (*hdkeystore.HDKey, error) {

	pw := ""

	if len(password) > 0 {
		pw = password[0]
	} else {
		pw = w.Password
	}

	if len(pw) == 0 {
		return nil, fmt.Errorf("password is empty")
	}

	if len(w.KeyFile) == 0 {
		return nil, errors.New("Wallet key is not exist!")
	}

	keyjson, err := ioutil.ReadFile(w.KeyFile)
	if err != nil {
		return nil, err
	}
	key, err := hdkeystore.DecryptHDKey(keyjson, pw)
	if err != nil {
		return nil, err
	}
	return key, err
}

//FileName 钱包文件名
func (w *Wallet) FileName() string {
	return w.fileName
}

// openDB 打开钱包数据库
func (w *Wallet) OpenDB() (*storm.DB, error) {
	abspath, _ := filepath.Abs(w.DBFile)
	return storm.Open(abspath)
}

//SaveToDB 保存到数据库
func (w *Wallet) SaveToDB() error {
	db, err := w.OpenDB()
	if err != nil {
		return nil
	}
	defer db.Close()
	return db.Save(w)
}

//GetAssetsAccounts 获取某种区块链的全部资产账户
func (w *Wallet) GetAssetsAccounts(symbol string) []*AssetsAccount {
	return nil
}

//GetAddress 通过地址字符串获取地址对象
func (w *Wallet) GetAddress(address string) *Address {
	db, err := w.OpenDB()
	if err != nil {
		log.Error("open db failed, err=", err)
		return nil
	}
	defer db.Close()

	var obj Address
	err = db.One("Address", address, &obj)
	if err != nil {
		log.Debugf("get address failed, err=%v", err)
		return nil
	}
	return &obj
}

//GetAddressesByAccountID 通过账户ID获取地址列表
func (w *Wallet) GetAddressesByAccount(accountID string) []*Address {
	db, err := w.OpenDB()
	if err != nil {
		return nil
	}
	defer db.Close()

	var obj []*Address
	db.Find("AccountID", accountID, &obj)
	return obj
}

//SingleAssetsAccount 把钱包作为一个单资产账户来使用
func (w *Wallet) SingleAssetsAccount(symbol string) *AssetsAccount {
	a := AssetsAccount{
		WalletID:  w.WalletID,
		Alias:     w.Alias,
		AccountID: w.WalletID,
		Index:     0,
		HDPath:    "",
		Required:  0,
		OwnerKeys: []string{w.RootPub},
		Symbol:    symbol,
	}

	return &a
}

//SaveRecharge 保存交易记录
func (w *Wallet) SaveRecharge(tx *Recharge) error {
	db, err := w.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Save(tx)
	if err != nil {
		return err
	}
	return nil
}

//SaveUnreceivedRecharge 保存未提交的充值记录
func (w *Wallet) SaveUnreceivedRecharge(tx *Recharge) error {
	db, err := w.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	//找是否已经有存在发送过的记录
	var findReceived []*Recharge
	err = db.Select(q.And(
		q.Eq("Received", true),
		q.Eq("Sid", tx.Sid),
	)).Find(&findReceived)

	if findReceived != nil {
		//不执行保存
		log.Error("findReceived:", findReceived)
		return nil
	}

	err = db.Save(tx)
	if err != nil {
		return err
	}
	return nil
}

//DropRecharge 删除充值记录表
func (w *Wallet) DropRecharge() error {
	db, err := w.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Drop("Recharge")
	//return db.Save(tx)
}

//GetRecharges 获取钱包相关的充值记录
func (w *Wallet) GetRecharges(received bool, height ...uint64) ([]*Recharge, error) {

	var (
		list []*Recharge
	)

	db, err := w.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if len(height) > 0 {
		err = db.Select(q.And(
			q.Eq("Received", received),
			q.Eq("BlockHeight", height[0]),
		)).Find(&list)
		//err = db.Find("BlockHeight", height[0], &list)
	} else {
		err = db.Select(q.And(q.Eq("Received", received))).Find(&list)
		//err = db.All(&list)
	}

	if err != nil {
		return nil, err
	}

	return list, nil
}

//GetUnconfrimRecharges
func (w *Wallet) GetUnconfrimRecharges(limitTime int64) ([]*Recharge, error) {
	var (
		list []*Recharge
	)

	db, err := w.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Select(q.And(
		q.Eq("BlockHeight", 0),
		q.Eq("Delete", false),
		q.Lte("CreateAt", limitTime),
	)).Find(&list)

	if err != nil {
		return nil, err
	}

	return list, nil
}

//GetWalletsByKeyDir 通过给定的文件路径加载keystore文件得到钱包列表
func GetWalletsByKeyDir(dir string) ([]*Wallet, error) {

	var (
		wallets = make([]*Wallet, 0)
	)

	//扫描key目录的所有钱包
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return wallets, err
	}

	for _, fi := range files {
		// Skip any non-key files from the folder
		if !file.IsUserFile(fi) {
			continue
		}
		if fi.IsDir() {
			continue
		}

		fileName := strings.TrimSuffix(fi.Name(), ".key")

		path := filepath.Join(dir, fi.Name())

		w := ReadWalletByKey(path)
		w.KeyFile = path
		w.fileName = fileName
		wallets = append(wallets, w)

	}

	return wallets, nil

}

//ReadWalletByKey 加载文件，实例化钱包
func ReadWalletByKey(keyPath string) *Wallet {

	var (
		buf = new(bufio.Reader)
		key struct {
			Alias string `json:"alias"`
			KeyID string `json:"keyid"`
		}
	)

	fd, err := os.Open(keyPath)
	defer fd.Close()
	if err != nil {
		return nil
	}

	buf.Reset(fd)
	// Parse the address.
	key.Alias = ""
	key.KeyID = ""
	err = json.NewDecoder(buf).Decode(&key)
	if err != nil {
		return nil
	}

	return &Wallet{WalletID: key.KeyID, Alias: key.Alias}
}
