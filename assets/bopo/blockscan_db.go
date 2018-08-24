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

package bopo

import (
	"path/filepath"
	"strings"
	//"strings"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/pkg/errors"

	//"github.com/asdine/storm/q"
	"fmt"
	//"github.com/blocktree/OpenWallet/openwallet"
	//"github.com/tidwall/gjson"
)

//SaveLocalNewBlock 写入本地区块高度和hash
func (wm *WalletManager) SaveLocalNewBlock(blockHeight uint64, blockHash string) error {

	fmt.Println("EEE = ", filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Set(blockchainBucket, "blockHeight", &blockHeight); err != nil {
		return err
	}
	if err := db.Set(blockchainBucket, "blockHash", &blockHash); err != nil {
		return err
	}
	return nil
}

//GetLocalNewBlock 获取本地区块高度和hash
func (wm *WalletManager) GetLocalNewBlock() (uint64, string) {

	var (
		blockHeight uint64 = 0
		blockHash   string = ""
	)

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return 0, ""
	}
	defer db.Close()

	db.Get(blockchainBucket, "blockHeight", &blockHeight)
	db.Get(blockchainBucket, "blockHash", &blockHash)

	return blockHeight, blockHash
}

//SaveLocalBlock 记录本地新区块
func (wm *WalletManager) SaveLocalBlock(block *Block) error {

	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(block)
}

//GetLocalBlock 获取本地区块数据
func (wm *WalletManager) GetLocalBlock(height uint64) (*Block, error) {

	var (
		block Block
	)

	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.One("Height", height, &block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

//SaveTransaction 记录高度到本地
func (wm *WalletManager) SaveTransaction(blockHeight uint64) error {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Set(blockchainBucket, "blockHeight", &blockHeight)
}

// ----------------------------------------------------------------------------
//SaveTxToWalletDB 保存交易记录到钱包数据库
func (wm *WalletManager) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	//if record.BlockHeight == 0 {
	//	return errors.New("unconfirmed transaction do not rescan")
	//}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//获取未扫记录
func (wm *WalletManager) GetUnscanRecords() ([]*UnscanRecord, error) {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//DeleteUnscanRecord 删除指定高度的未扫记录
func (wm *WalletManager) DeleteUnscanRecord(height uint64) error {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.Find("BlockHeight", height, &list)
	if err != nil {
		return err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	for _, r := range list {
		tx.DeleteStruct(r)
	}

	return tx.Commit()
}

//DeleteUnscanRecordNotFindTX 删除未没有找到交易记录的重扫记录
func (wm *WalletManager) DeleteUnscanRecordNotFindTX() error {

	//删除找不到交易单
	reason := "[-5]No information available about transaction"

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	err = db.All(&list)
	if err != nil {
		return err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	for _, r := range list {
		if strings.HasPrefix(r.Reason, reason) {
			tx.DeleteStruct(r)
		}
	}
	return tx.Commit()
}

//DeleteUnscanRecordByTxID 删除未扫记录
func (wm *WalletManager) DeleteUnscanRecordByTxID(height uint64, txid string) error {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(wm.config.dbPath, wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	db.Select(q.And(
		q.Eq("TxID", txid),
		q.Eq("BlockHeight", height),
	)).Find(&list)
	//err = db.Find("TxID", txid, &list)
	if err != nil {
		return err
	}

	for _, r := range list {
		db.DeleteStruct(r)
	}

	return nil
}
