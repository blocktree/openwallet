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
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/crypto"
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

const (
	blockchainBucket      = "blockchain"
	CurrentBlockHeaderKey = "current_block_header"
)

//BlockchainLocal 区块链数据访问本地数据库实现
type BlockchainLocal struct {
	blockchainDB     *storm.DB //区块链数据库
	blockchainDBFile string
	keepOpen         bool //保持打开状态
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
	return db.Save(header)
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
	err = db.One("Height", height, &blockHeader)
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
	defer db.Close()
	return db.Save(record)
}

func (base *BlockchainLocal) DeleteUnscanRecordByHeight(height uint64, symbol string) error {
	db, err := base.getDB()
	if err != nil {
		return err
	}
	defer db.Close()

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
	defer db.Close()

	var r UnscanRecord
	err = db.One("ID", id, &r)
	if err != nil {
		return err
	}
	return db.DeleteStruct(&r)
}

func (base *BlockchainLocal) GetTransactionsByTxID(txid, symbol string) ([]*Transaction, error) {
	return nil, fmt.Errorf("GetTransactionsByTxID is not implemented")
}

func (base *BlockchainLocal) GetUnscanRecords(symbol string) ([]*UnscanRecord, error) {
	db, err := base.getDB()
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
