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
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/coreos/bbolt"
)

type WrapperSourceFile string

// Wrapper 基于OpenWallet钱包体系模型，专门处理钱包的持久化问题，关系数据查询
type Wrapper struct {
	WalletDAIBase
	sourceDB     *StormDB     //存储钱包相关数据的数据库，目前使用boltdb作为持久方案
	mu           sync.RWMutex //锁
	isExternalDB bool         //是否外部加载的数据库，非内部打开，内部打开需要关闭
	sourceFile   string       //钱包数据库文件路径，用于内部打开
}

func NewWrapper(args ...interface{}) *Wrapper {

	wrapper := Wrapper{}

	for _, arg := range args {
		switch obj := arg.(type) {
		case *StormDB:
			if obj != nil {
				//if !obj.Opened {
				//	return nil, fmt.Errorf("wallet db is close")
				//}

				wrapper.isExternalDB = true
				wrapper.sourceDB = obj
			}
		case WrapperSourceFile:
			wrapper.sourceFile = string(obj)
		}
	}

	return &wrapper
}

//OpenStormDB 打开数据库
func (wrapper *Wrapper) OpenStormDB() (*StormDB, error) {

	var (
		db  *StormDB
		err error
	)

	if wrapper.sourceDB != nil && wrapper.sourceDB.Opened {
		return wrapper.sourceDB, nil
	}

	//保证数据库文件并发下不被同时打开
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	//解锁进入后，再次确认是否已经存在
	if wrapper.sourceDB != nil && wrapper.sourceDB.Opened {
		return wrapper.sourceDB, nil
	}

	//log.Debugf("sourceFile :%v", wrapper.sourceFile)
	db, err = OpenStormDB(
		wrapper.sourceFile,
		storm.BoltOptions(0600, &bolt.Options{Timeout: 3 * time.Second}),
	)

	if err != nil {
		return nil, err
	}

	wrapper.isExternalDB = false
	wrapper.sourceDB = db

	return db, nil
}

//SetStormDB 设置钱包的应用数据库
func (wrapper *Wrapper) SetExternalDB(db *StormDB) error {

	//关闭之前的数据库
	wrapper.CloseDB()

	//保证数据库文件并发下不被同时打开
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()

	wrapper.sourceDB = db
	wrapper.isExternalDB = true

	return nil
}

//CloseDB 关闭数据库
func (wrapper *Wrapper) CloseDB() {
	// 如果是外部引入的数据库不进行关闭，因为这样会外部无法再操作同一个数据库实力
	if wrapper.isExternalDB == false {
		if wrapper.sourceDB != nil && wrapper.sourceDB.Opened {
			wrapper.sourceDB.Close()
		}
	}
}
