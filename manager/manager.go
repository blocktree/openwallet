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

package manager

import (
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/coreos/bbolt"
	"path/filepath"
	"sync"
	"time"
	"io/ioutil"
	"strings"
)

type appWalletWrapper struct {
	*openwallet.WalletWrapper
}

//newAppWalletWrapper 创建App专用的包装器
func newAppWalletWrapper(db *openwallet.StormDB, arg ...interface{}) (*appWalletWrapper, error) {

	//var wallet openwallet.Wallet
	//err := db.One("WalletID", walletID, &wallet)
	//if err != nil {
	//	return nil, err
	//}

	wrapper, err := openwallet.NewWalletWrapper(db, arg)
	if err != nil {
		return nil, err
	}

	return &appWalletWrapper{wrapper}, nil
}

//WalletManager OpenWallet钱包管理器
type WalletManager struct {
	appDB       map[string]*openwallet.StormDB
	cfg         *Config
	initialized bool
	mu          sync.RWMutex
}

// NewWalletManager
func NewWalletManager(config *Config) *WalletManager {
	wm := WalletManager{}
	wm.cfg = config

	return &wm
}

func (wm *WalletManager) Init() {

	//新建文件目录
	file.MkdirAll(wm.cfg.dbPath)
	file.MkdirAll(wm.cfg.keyDir)

	wm.appDB = make(map[string]*openwallet.StormDB)

	wm.initBlockScanner()

	wm.initialized = true
}

// initBlockScanner 初始化区块链扫描器
func (wm *WalletManager) initBlockScanner() error {

	//加载已存在所有app
	appIDs, err := wm.loadAllAppIDs()
	if err != nil {
		return err
	}

	for _, symbol := range wm.cfg.supportAssets {
		assetsMgr, err := GetAssetsManager(symbol)
		if err != nil {
			log.Error(symbol, "is not support")
			continue
		}
		scanner := assetsMgr.GetBlockScanner()

		//加载地址时，暂停区块扫描
		scanner.Pause()

		for _, appID := range appIDs {

			wrapper, err :=wm.newWalletWrapper(appID)
			if err != nil {
				log.Error("wallet manager init unexpected error:", err)
				continue
			}

			addrs, err := wrapper.GetAddressList(0, -1)

			for _, address := range addrs {

				//TODO:加载所有应用钱包地址到扫描器
				scanner.AddAddress(address.Address, appID)
			}

		}

		scanner.Run()
	}

	return nil
}

//DBFile 应用数据库文件
func (wm *WalletManager) DBFile(appID string) string {
	return filepath.Join(wm.cfg.dbPath, appID+".db")
}

//OpenDB 打开应用数据库文件
func (wm *WalletManager) OpenDB(appID string) (*openwallet.StormDB, error) {

	var (
		db  *openwallet.StormDB
		err error
		ok  bool
	)

	//数据库文件
	db, ok = wm.appDB[appID]

	if ok && db.Opened {
		return db, nil
	}

	//保证数据库文件并发下不被同时打开
	wm.mu.Lock()
	defer wm.mu.Unlock()

	//解锁进入后，再次确认是否已经存在
	db, ok = wm.appDB[appID]
	if ok && db.Opened {
		return db, nil

	}

	db, err = openwallet.OpenStormDB(
		wm.DBFile(appID),
		storm.Batch(),
		storm.BoltOptions(0600, &bolt.Options{Timeout: 3 * time.Second}),
	)
	log.Debug("open storm db appID:", appID)
	if err != nil {
		return nil, err
	}

	//return opendb, nil
	wm.appDB[appID] = db

	return db, nil
}

//CloseDB 关闭应用数据库文件
func (wm *WalletManager) CloseDB(appID string) error {

	//数据库文件
	db, ok := wm.appDB[appID]
	if ok {
		db.Close()
	}

	return nil
}

//loadAllAppIDs 加载全部应用ID
func  (wm *WalletManager) loadAllAppIDs() ([]string, error) {

	var (
		apps = make([]string, 0)
		dir = wm.cfg.dbPath
	)

	//扫描key目录的所有钱包
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, fi := range files {
		// Skip any non-key files from the folder
		if !file.IsUserFile(fi) {
			continue
		}
		if fi.IsDir() {
			continue
		}

		appID := strings.TrimSuffix(fi.Name(), ".db")
		apps = append(apps, appID)

	}

	return apps, nil
}

//newWalletWrapper 创建App专用的包装器
func (wm *WalletManager) newWalletWrapper(appID string) (*appWalletWrapper, error) {

	//打开数据库
	db, err := wm.OpenDB(appID)
	if err != nil {
		return nil, err
	}
	dbFile := openwallet.WalletDBFile(wm.DBFile(appID))
	return newAppWalletWrapper(db, dbFile)
}