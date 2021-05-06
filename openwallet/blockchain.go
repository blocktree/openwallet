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
	"fmt"
	"github.com/asdine/storm"
	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/crypto"
)

type Blockchain struct {
	Blocks     uint64
	Headers    uint64
	ScanHeight uint64
}

// BlockHeader 区块头
type BlockHeader struct {
	Hash              string `json:"hash"`
	Confirmations     uint64 `json:"confirmations"`
	Merkleroot        string `json:"merkleroot"`
	Previousblockhash string `json:"previousblockhash"`
	Height            uint64 `json:"height" storm:"id"`
	Version           uint64 `json:"version"`
	Time              uint64 `json:"time"`
	Fork              bool   `json:"fork"`
	Symbol            string `json:"symbol"`
}

//BlockScanNotify 新区块扫描完成通知
//@param  txs 每扫完区块链，与地址相关的交易到
type BlockScanNotify func(header *BlockHeader)

//UnscanRecord 扫描失败的区块及交易
type UnscanRecord struct {
	ID          string `storm:"id"` // primary key
	BlockHeight uint64 `json:"blockHeight"`
	TxID        string `json:"txid"`
	Reason      string `json:"reason"`
	Symbol      string `json:"symbol"`
}

//NewUnscanRecord new UnscanRecord
func NewUnscanRecord(height uint64, txID, reason, symbol string) *UnscanRecord {
	obj := UnscanRecord{}
	obj.BlockHeight = height
	obj.TxID = txID
	obj.Reason = reason
	obj.Symbol = symbol
	obj.ID = common.Bytes2Hex(crypto.SHA256([]byte(fmt.Sprintf("%s_%d_%s", symbol, height, txID))))
	return &obj
}

//BlockchainDAI 区块链数据访问接口
type BlockchainDAI interface {
	SaveCurrentBlockHead(header *BlockHeader) error
	GetCurrentBlockHead(symbol string) (*BlockHeader, error)
	SaveLocalBlockHead(header *BlockHeader) error
	GetLocalBlockHeadByHeight(height uint64, symbol string) (*BlockHeader, error)
	SaveUnscanRecord(record *UnscanRecord) error
	DeleteUnscanRecordByHeight(height uint64, symbol string) error
	DeleteUnscanRecordByID(id string, symbol string) error
	GetTransactionsByTxID(txid, symbol string) ([]*Transaction, error)
	GetUnscanRecords(symbol string) ([]*UnscanRecord, error)
	SetMaxBlockCache(max uint64, symbol string) error
	SaveTransaction(tx *Transaction) error
}

//BlockchainDAIBase 区块链数据访问接口基本实现
type BlockchainDAIBase int

func (base *BlockchainDAIBase) SaveCurrentBlockHead(header *BlockHeader) error {
	return fmt.Errorf("SaveCurrentBlockHead is not implemented")
}

func (base *BlockchainDAIBase) GetCurrentBlockHead(symbol string) (*BlockHeader, error) {
	return nil, fmt.Errorf("GetCurrentBlockHead is not implemented")
}

func (base *BlockchainDAIBase) SaveLocalBlockHead(header *BlockHeader) error {
	return fmt.Errorf("SaveLocalBlockHead is not implemented")
}

func (base *BlockchainDAIBase) GetLocalBlockHeadByHeight(height uint64, symbol string) (*BlockHeader, error) {
	return nil, fmt.Errorf("GetLocalBlockHeadByHeight is not implemented")
}

func (base *BlockchainDAIBase) SaveUnscanRecord(record *UnscanRecord) error {
	return fmt.Errorf("SaveUnscanRecord is not implemented")
}

func (base *BlockchainDAIBase) DeleteUnscanRecordByHeight(height uint64, symbol string) error {
	return fmt.Errorf("DeleteUnscanRecordByHeight is not implemented")
}

func (base *BlockchainDAIBase) DeleteUnscanRecordByID(id string, symbol string) error {
	return fmt.Errorf("DeleteUnscanRecordByID is not implemented")
}

func (base *BlockchainDAIBase) GetTransactionsByTxID(txid, symbol string) ([]*Transaction, error) {
	return nil, fmt.Errorf("GetTransactionsByTxID is not implemented")
}

func (base *BlockchainDAIBase) GetUnscanRecords(symbol string) ([]*UnscanRecord, error) {
	return nil, fmt.Errorf("GetUnscanRecords is not implemented")
}

func (base *BlockchainDAIBase) SetMaxBlockCache(size uint64, symbol string) error {
	return fmt.Errorf("SetMaxBlockCache is not implemented")
}

func (base *BlockchainDAIBase) SaveTransaction(tx *Transaction) error {
	return fmt.Errorf("SaveTransaction is not implemented")
}

const (
	blockchainBucket             = "blockchain"
	CurrentBlockHeaderKey        = "current_block_header"
	CurrentBlockIncreaseIndexKey = "current_block_increase_index"
	BlockCacheIndexKey           = "block_cache_index"
	blockIndexBucket             = "block_index_bucket"
	blockCacheBucket             = "block_cache_bucket"
)

//BlockchainLocal 区块链数据访问本地数据库实现
type BlockchainLocal struct {
	BlockchainDAIBase
	blockchainDB     *storm.DB //区块链数据库
	blockchainDBFile string    //区块db文件
	keepOpen         bool      //保持打开状态
	blockCacheSize   uint64    //区块缓存数量
}

// NewBlockchainLocal 加载区块链数据库
func NewBlockchainLocal(dbFile string, keepOpen bool) (*BlockchainLocal, error) {
	base := new(BlockchainLocal)
	//区块链数据文件
	base.blockchainDBFile = dbFile

	base.keepOpen = keepOpen
	if base.keepOpen {
		blockchaindb, err := storm.Open(base.blockchainDBFile)
		if err != nil {
			return nil, err
		}
		base.blockchainDB = blockchaindb
	}

	//默认缓存1000个区块
	base.SetMaxBlockCache(1000, "")
	return base, nil
}

func (base *BlockchainLocal) getDB() (*storm.DB, error) {
	if !base.keepOpen {
		blockchaindb, err := storm.Open(base.blockchainDBFile)
		if err != nil {
			return nil, err
		}
		base.blockchainDB = blockchaindb
	}

	return base.blockchainDB, nil
}

func (base *BlockchainLocal) closeDB() error {
	//区块链数据文件
	if !base.keepOpen {
		return base.blockchainDB.Close()
	}
	return nil
}

func (base *BlockchainLocal) SaveCurrentBlockHead(header *BlockHeader) error {
	db, err := base.getDB()
	if err != nil {
		return err
	}
	defer base.closeDB()

	return db.Set(blockchainBucket, CurrentBlockHeaderKey, header)
}

func (base *BlockchainLocal) GetCurrentBlockHead(symbol string) (*BlockHeader, error) {
	db, err := base.getDB()
	if err != nil {
		return nil, err
	}
	defer base.closeDB()
	var header BlockHeader
	db.Get(blockchainBucket, CurrentBlockHeaderKey, &header)

	return &header, nil
}

func (base *BlockchainLocal) SaveLocalBlockHead(header *BlockHeader) error {

	db, err := base.getDB()
	if err != nil {
		return err
	}
	defer base.closeDB()

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	//查询当前记录的索引位
	var increaseIndex uint64
	tx.Get(blockIndexBucket, CurrentBlockIncreaseIndexKey, &increaseIndex)

	if increaseIndex >= base.blockCacheSize {
		increaseIndex = 0
	}

	//递增索引
	increaseIndex++
	tx.Set(blockIndexBucket, CurrentBlockIncreaseIndexKey, increaseIndex)

	//查询该索引位已存在的高度
	key := fmt.Sprintf("%s_%d", BlockCacheIndexKey, increaseIndex)
	var delHeight uint64
	tx.Get(blockIndexBucket, key, &delHeight)

	//移除该记录的高度
	if delHeight > 0 {
		delKey := fmt.Sprintf("%d", delHeight)
		tx.Delete(blockCacheBucket, delKey)
	}

	//记录新高度的区块信息
	saveKey := fmt.Sprintf("%d", header.Height)
	tx.Set(blockCacheBucket, saveKey, header)

	//记录改索引位的新高度
	tx.Set(blockIndexBucket, key, header.Height)

	//total, _ := tx.Count(&BlockHeader{})
	//if total > 500 {
	//	log.Debugf("blocks total = %d", total)
	//	err = tx.Drop("BlockHeader")
	//	if err != nil {
	//		return err
	//	}
	//}

	//lower := header.Height - 500
	//if lower > 0 {
	//	var blocks []*BlockHeader
	//	tx.Select(q.Lte("Height", lower)).Find(&blocks)
	//	log.Debugf("blocks count = %d", len(blocks))
	//	for _, b := range blocks {
	//		tx.DeleteStruct(b)
	//	}
	//}

	//err = tx.Save(header)
	//if err != nil {
	//	return err
	//}

	return tx.Commit()
}

func (base *BlockchainLocal) GetLocalBlockHeadByHeight(height uint64, symbol string) (*BlockHeader, error) {

	var (
		blockHeader BlockHeader
	)

	db, err := base.getDB()
	if err != nil {
		return nil, err
	}
	defer base.closeDB()

	//err = db.One("Height", height, &blockHeader)
	//if err != nil {
	//	return nil, err
	//}

	//return &blockHeader, nil

	key := fmt.Sprintf("%d", height)
	err = db.Get(blockCacheBucket, key, &blockHeader)
	if err != nil {
		return nil, err
	}

	return &blockHeader, nil
}

func (base *BlockchainLocal) SaveUnscanRecord(record *UnscanRecord) error {

	if record == nil {
		return fmt.Errorf("the unscan record to save is nil")
	}
	db, err := base.getDB()
	if err != nil {
		return err
	}
	defer base.closeDB()
	return db.Save(record)
}

func (base *BlockchainLocal) DeleteUnscanRecordByHeight(height uint64, symbol string) error {
	db, err := base.getDB()
	if err != nil {
		return err
	}
	defer base.closeDB()

	var list []*UnscanRecord
	err = db.Find("BlockHeight", height, &list)
	if err != nil {
		return err
	}
	for _, r := range list {
		db.DeleteStruct(r)
	}
	return nil
}

func (base *BlockchainLocal) DeleteUnscanRecordByID(id string, symbol string) error {
	db, err := base.getDB()
	if err != nil {
		return err
	}
	defer base.closeDB()

	var r UnscanRecord
	err = db.One("ID", id, &r)
	if err != nil {
		return err
	}

	return db.DeleteStruct(&r)
}

func (base *BlockchainLocal) GetTransactionsByTxID(txid, symbol string) ([]*Transaction, error) {
	db, err := base.getDB()
	if err != nil {
		return nil, err
	}
	defer base.closeDB()

	var list []*Transaction
	err = db.Find("TxID", txid, &list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (base *BlockchainLocal) GetUnscanRecords(symbol string) ([]*UnscanRecord, error) {
	db, err := base.getDB()
	if err != nil {
		return nil, err
	}
	defer base.closeDB()

	var list []*UnscanRecord
	err = db.All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (base *BlockchainLocal) SetMaxBlockCache(size uint64, symbol string) error {
	base.blockCacheSize = size
	return nil
}

func (base *BlockchainLocal) SaveTransaction(tx *Transaction) error {
	if tx == nil {
		return fmt.Errorf("the transaction to save is nil")
	}
	db, err := base.getDB()
	if err != nil {
		return err
	}
	defer base.closeDB()
	return db.Save(tx)
}
