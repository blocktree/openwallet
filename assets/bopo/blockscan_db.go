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
	"errors"
	"path/filepath"
	"strings"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

//SaveLocalNewBlock 写入本地区块高度和hash
func (bs *FabricBlockScanner) SaveLocalNewBlock(blockHeight uint64, blockHash string) error {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
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
func (bs *FabricBlockScanner) GetLocalNewBlock() (uint64, string) {

	var (
		blockHeight uint64 = 0
		blockHash   string = ""
	)

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
	if err != nil {
		return 0, ""
	}
	defer db.Close()

	db.Get(blockchainBucket, "blockHeight", &blockHeight)
	db.Get(blockchainBucket, "blockHash", &blockHash)

	return blockHeight, blockHash
}

//SaveLocalBlock 记录本地新区块
func (bs *FabricBlockScanner) SaveLocalBlock(block *Block) error {

	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(block)
}

//GetLocalBlock 获取本地区块数据
func (bs *FabricBlockScanner) GetLocalBlock(height uint64) (*Block, error) {

	var (
		block Block
	)

	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
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
func (bs *FabricBlockScanner) SaveTransaction(blockHeight uint64) error {

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Set(blockchainBucket, "blockHeight", &blockHeight)
}

// ----------------------------------------------------------------------------
//SaveTxToWalletDB 保存交易记录到钱包数据库
func (bs *FabricBlockScanner) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return errors.New("the unscan record to save is nil")
	}

	//if record.BlockHeight == 0 {
	//	return errors.New("unconfirmed transaction do not rescan")
	//}

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(record)
}

//获取未扫记录
func (bs *FabricBlockScanner) GetUnscanRecords() ([]*UnscanRecord, error) {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
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
func (bs *FabricBlockScanner) DeleteUnscanRecord(height uint64) error {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
	if err != nil {
		log.Error("Open Failed: ", err)
		return err
	}
	defer db.Close()

	var list []*UnscanRecord
	// err = db.Find("BlockHeight", height, &list)
	err = db.Find("ID", string(height), &list)
	if err != nil {
		log.Error("Find Faild: ", err)
		return err
	}

	tx, err := db.Begin(true)
	if err != nil {
		log.Error("Begin Failed: ", err)
		return err
	}

	for _, r := range list {
		tx.DeleteStruct(r)
	}

	return tx.Commit()
}

//DeleteUnscanRecordNotFindTX 删除未没有找到交易记录的重扫记录
func (bs *FabricBlockScanner) DeleteUnscanRecordNotFindTX() error {

	//删除找不到交易单
	reason := "[-5]No information available about transaction"

	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
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
func (bs *FabricBlockScanner) DeleteUnscanRecordByTxID(height uint64, txid string) error {
	//获取本地区块高度
	db, err := storm.Open(filepath.Join(bs.wm.config.dbPath, bs.wm.config.blockchainFile))
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

//GetWalletByAddress 获取地址对应的钱包
func (bs *FabricBlockScanner) GetWalletByAddress(address string) (*openwallet.Wallet, bool) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	account, ok := bs.addressInScanning[address]
	if ok {
		wallet, ok := bs.walletInScanning[account]
		return wallet, ok

	} else {
		return nil, false
	}
}

//SaveRechargeToWalletDB 保存交易单内的充值记录到钱包数据库
func (bs *FabricBlockScanner) SaveRechargeToWalletDB(height uint64, list []*openwallet.Recharge) error {

	var (
		saveSuccess = true
	)

	for _, r := range list {

		wallet, ok := bs.GetWalletByAddress(r.Address)
		if ok {
			reason := ""
			err := wallet.SaveUnreceivedRecharge(r)
			//如果blockHash没有值，添加到重扫，避免遗留
			if err != nil {
				saveSuccess = false
				//记录未扫区块
				reason = err.Error()
				log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
				unscanRecord := NewUnscanRecord(height, r.TxID, reason)

				err = bs.SaveUnscanRecord(unscanRecord)
				if err != nil {
					log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
				}

			} else {
				log.Info("block scanner save blockHeight:", height, "txid:", r.TxID, "address:", r.Address, "successfully.")
			}

			//if err != nil || len(r.BlockHash) == 0 {
			//	saveSuccess = false
			//	//记录未扫区块
			//	if err != nil {
			//		reason = err.Error()
			//		log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
			//	}
			//	unscanRecord := NewUnscanRecord(height, r.TxID, reason)
			//	err = bs.SaveUnscanRecord(unscanRecord)
			//	if err != nil {
			//		log.Std.Error("block height: %d, txID: %s save unscan record failed. unexpected error: %v", height, r.TxID, err.Error())
			//	}
			//
			//} else {
			//	log.Info("block scanner save blockHeight:", height, "txid:", r.TxID, "address:", r.Address, "successfully.")
			//}
		} else {
			log.Error("address:", r.Address, "in wallet is not found, txid:", r.TxID)
			return errors.New("address in wallet is not found")
		}

	}

	if !saveSuccess {
		return errors.New("have unscan record")
	}

	return nil
}

//DeleteRechargesByHeight 删除某区块高度的充值记录
func (bs *FabricBlockScanner) DeleteRechargesByHeight(height uint64) error {

	bs.mu.RLock()
	defer bs.mu.RUnlock()

	for _, wallet := range bs.walletInScanning {

		list, err := wallet.GetRecharges(false, height)
		if err != nil {
			return err
		}

		db, err := wallet.OpenDB()
		if err != nil {
			return err
		}

		tx, err := db.Begin(true)
		if err != nil {
			return err
		}

		for _, r := range list {
			err = db.DeleteStruct(&r)
			if err != nil {
				return err
			}
		}

		tx.Commit()

		db.Close()
	}

	return nil
}
