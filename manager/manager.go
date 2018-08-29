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
	"github.com/coreos/bbolt"
	"path/filepath"
	"sync"
	"time"
	"github.com/blocktree/OpenWallet/openwallet"
)

type WalletManager struct {
	//	TODO:自定义持久化数据源
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

	wm.initialized = true
}

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

	dbFile := filepath.Join(wm.cfg.dbPath, appID+".db")
	db, err = openwallet.OpenStormDB(
		dbFile,
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

func (wm *WalletManager) CloseDB(appID string) error {

	//数据库文件
	db, ok := wm.appDB[appID]
	if ok {
		db.Close()
	}

	return nil
}
