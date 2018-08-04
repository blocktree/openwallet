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

package tezos

import (
	"github.com/blocktree/OpenWallet/openwallet"
	"io/ioutil"
	"path/filepath"
	"os"
	"strings"
	"github.com/asdine/storm"
)

type Key struct {
	Address    string `storm:"id"`
	PublicKey  string
	PrivateKey string
}

func NewKey(addr, pub, priv string) *Key {
	return &Key{Address: addr, PublicKey: pub, PrivateKey: priv}
}

//DecryptPrivateKey 解密私钥
func (k *Key) DecryptPrivateKey() string {

	return ""
}

//SaveKeyToWallet 保存私钥给钱包数据库
func SaveKeyToWallet(wallet *openwallet.Wallet, keys []*Key) error {
	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, a := range keys {
		err = tx.Save(a)
		if err != nil {
			continue
		}
	}

	return tx.Commit()
}



/***** 钱包相关 *****/


//GetWalletKeys 通过给定的文件路径加载keystore文件得到钱包列表
func GetWallets() ([]*openwallet.Wallet, error) {

	var (
		wallets = make([]*openwallet.Wallet, 0)
	)

	//扫描key目录的所有钱包
	files, err := ioutil.ReadDir(dbPath)
	if err != nil {
		return wallets, err
	}

	for _, fi := range files {
		// Skip any non-key files from the folder
		if skipKeyFile(fi) {
			continue
		}
		if fi.IsDir() {
			continue
		}
		w, err := GetWalletByID(fi.Name())
		if err != nil {
			continue
		}
		wallets = append(wallets, w)

	}

	return wallets, nil

}

//GetWalletByID 获取钱包
func GetWalletByID(walletID string) (*openwallet.Wallet, error) {

	dbFile := filepath.Join(dbPath, walletID+".db")
	db, err := storm.Open(dbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var wallet openwallet.Wallet
	err = db.One("WalletID", walletID, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// skipKeyFile ignores editor backups, hidden files and folders/symlinks.
func skipKeyFile(fi os.FileInfo) bool {
	// Skip editor backups and UNIX-style hidden files.
	if strings.HasSuffix(fi.Name(), "~") || strings.HasPrefix(fi.Name(), ".") {
		return true
	}
	// Skip misc special files, directories (yes, symlinks too).
	if fi.IsDir() || fi.Mode()&os.ModeType != 0 {
		return true
	}
	return false
}