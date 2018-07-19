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

package openwallet

import (
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
	"reflect"
)

type Wallet struct {
	//Coin      string `json:"coin"`
	WalletID  string `json:"walletID"  storm:"id"`
	Alias     string `json:"alias"`
	Password  string `json:"password"`
	RootPub   string `json:"rootpub"`
	KeyFile   string `json:"keyFile"`   //钱包的密钥文件
	DBFile    string `json:"dbFile"`    //钱包的数据库文件
	WatchOnly bool   `json:"watchOnly"` //创建watchonly的钱包，没有私钥文件，只有db文件
	key       *keystore.HDKey

	//核心钱包指针
	core interface{}

	// 已解锁的钱包，集合（钱包地址, 钱包私钥）
	unlocked map[string]unlocked
}

func NewHDWallet(key *keystore.HDKey) (*Wallet, error) {

	return nil, nil
}

func NewWalletID() uuid.UUID {
	id := uuid.NewRandom()
	return id
}

//NewWatchOnlyWallet 只读钱包，用于观察冷钱包
func NewWatchOnlyWallet(walletID string, symbol string) *Wallet {

	dbDir := GetDBDir(symbol)
	file.MkdirAll(dbDir)

	dbFile := filepath.Join(dbDir, walletID+".db")

	w := Wallet{
		WalletID:  walletID,
		Alias:     walletID, //自定义ID也作为别名
		WatchOnly: true,
		DBFile:    dbFile,
	}

	return &w
}

//HDKey 获取钱包密钥，需要密码
func (w *Wallet) HDKey(password string) (*keystore.HDKey, error) {

	if len(w.KeyFile) == 0 {
		return nil, errors.New("Wallet key is not exist!")
	}

	keyjson, err := ioutil.ReadFile(w.KeyFile)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptHDKey(keyjson, password)
	if err != nil {
		return nil, err
	}
	return key, err
}

//openDB 打开钱包数据库
func (w *Wallet) OpenDB() (*storm.DB, error) {
	return storm.Open(w.DBFile)
}

//GetAssetsAccounts 获取某种区块链的全部资产账户
func (w *Wallet) GetAssetsAccounts(symbol string) []*AssetsAccount {
	return nil
}

//SingleAssetsAccount 把钱包作为一个单资产账户来使用
func (w *Wallet) SingleAssetsAccount(symbol string) *AssetsAccount {
	a := AssetsAccount{
		WalletID: w.WalletID,
		Alias:w.Alias,
		AccountID:w.WalletID,
		Index: 0,
		HDPath: "",
		Required:0,
		PublicKeys:[]string{w.RootPub},
		Symbol:symbol,
	}

	return &a
}

/*
//NewWallet 创建钱包
func NewWallet(publickeys []Bytes, users []*User, required uint, creator *User) (*Wallet, error) {

	var (
		owner    = make(map[string]string, 0)
		firstApp *App
	)

	//检查公钥和用户数量是否相等
	if len(publickeys) != len(users) {
		return nil, errors.New("PublicKey count is not equal Users count")
	}

	for i, pk := range publickeys {

		if i == 0 {
			firstApp = users[i].App
		}

		//检查公钥长度是否正确

		//检查用户key是否为空
		if len(users[i].UserKey) == 0 {
			return nil, errors.New("User's userkey is empty")
		}

		//检查用户的应用方是否一致
		if firstApp != nil && firstApp != users[i].App {
			return nil, errors.New("User's Application is not unified")
		}

		pkHex := common.Bytes2Hex(pk)
		owner[pkHex] = users[i].UserKey
	}

	w := &Wallet{
		PublicKeys: publickeys,
		Required:  required,
		Owners:    owner,
	}

	return w, nil
}
*/

////GetUserByPublicKey 通过公钥获取用户
//func (w *Wallet) GetUserByPublicKey(publickey PublicKey) *User {
//
//	pkHex := common.Bytes2Hex(publickey)
//
//	user := &User{
//		UserKey: w.Owners[pkHex],
//	}
//
//	return user
//}

// Deposit 充值
func (w *Wallet) Deposit(assets string) []byte {

	c := w.FindAssets(assets)

	return c.Deposit()
}

//FindAssets 寻找资产
func (w *Wallet) FindAssets(assets string) AssetsInferface {

	var (
		runAssets  reflect.Type
		findAssets bool
		assetsInfo *AssetsInfo
	)

	assetsInfo, findAssets = OpenWalletApp.Handlers.FindAssetsInfo(assets)

	if !findAssets {
		return nil
	}

	runAssets = assetsInfo.controllerType

	vc := reflect.New(runAssets)
	execController, ok := vc.Interface().(AssetsInferface)
	if !ok {
		return nil
	}

	execController.Init(w, vc.Interface())

	return execController
}
