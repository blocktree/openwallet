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
package ethereum

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"strconv"
	//"log"
	"math/big"

	"github.com/blocktree/openwallet/log"
	"github.com/shopspring/decimal"
)

const (
	TRANS_AMOUNT_UNIT_LIST = `
	1: wei
	2: Kwei
	3: Mwei
	4: GWei
	5: microether
	6: milliether
	7: ether
	`
	TRANS_AMOUNT_UNIT_WEI          = 1
	TRANS_AMOUNT_UNIT_K_WEI        = 2
	TRANS_AMOUNT_UNIT_M_WEI        = 3
	TRANS_AMOUNT_UNIT_G_WEI        = 4
	TRANS_AMOUNT_UNIT_MICRO_ETHER  = 5
	TRANS_AMOUNT_UNIT_MILLIE_ETHER = 6
	TRNAS_AMOUNT_UNIT_ETHER        = 7
)

func ConvertFloatStringToBigInt(amount string, decimals int) (*big.Int, error) {
	vDecimal, _ := decimal.NewFromString(amount)
	//if err != nil {
	//	log.Error("convert from string to decimal failed, err=", err)
	//	return nil, err
	//}

	if decimals <= 0 || decimals > 30 {
		return nil, errors.New("wrong decimal input through")
	}

	decimalInt := big.NewInt(1)
	for i := 0; i < decimals; i++ {
		decimalInt.Mul(decimalInt, big.NewInt(10))
	}

	d, _ := decimal.NewFromString(decimalInt.String())
	vDecimal = vDecimal.Mul(d)
	rst := new(big.Int)
	if _, valid := rst.SetString(vDecimal.String(), 10); !valid {
		log.Error("conver to big.int failed")
		return nil, errors.New("conver to big.int failed")
	}
	return rst, nil
}

func ConvertEthStringToWei(amount string) (*big.Int, error) {
	//log.Debug("amount:", amount)
	// vDecimal, err := decimal.NewFromString(amount)
	// if err != nil {
	// 	log.Error("convert from string to decimal failed, err=", err)
	// 	return nil, err
	// }

	// ETH, _ := decimal.NewFromString(strings.Replace("1,000,000,000,000,000,000", ",", "", -1))
	// vDecimal = vDecimal.Mul(ETH)
	// rst := new(big.Int)
	// if _, valid := rst.SetString(vDecimal.String(), 10); !valid {
	// 	log.Error("conver to big.int failed")
	// 	return nil, errors.New("conver to big.int failed")
	// }
	//return rst, nil
	return ConvertFloatStringToBigInt(amount, 18)
}

func ConvertAmountToFloatDecimal(amount string, decimals int) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		log.Error("convert string to deciaml failed, err=", err)
		return d, err
	}

	if decimals <= 0 || decimals > 30 {
		return d, errors.New("wrong decimal input through ")
	}

	decimalInt := big.NewInt(1)
	for i := 0; i < decimals; i++ {
		decimalInt.Mul(decimalInt, big.NewInt(10))
	}

	w, _ := decimal.NewFromString(decimalInt.String())
	d = d.Div(w)
	return d, nil
}

func ConverWeiStringToEthDecimal(amount string) (decimal.Decimal, error) {
	// d, err := decimal.NewFromString(amount)
	// if err != nil {
	// 	log.Error("convert string to deciaml failed, err=", err)
	// 	return d, err
	// }

	// ETH, _ := decimal.NewFromString(strings.Replace("1,000,000,000,000,000,000", ",", "", -1))
	// d = d.Div(ETH)
	// return d, nil
	return ConvertAmountToFloatDecimal(amount, 18)
}

func ConverEthDecimalToWei(amount decimal.Decimal) (*big.Int, error) {
	return ConvertFloatStringToBigInt(amount.String(), 18)
}

func toHexBigIntForEtherTrans(value string, base int, unit int64) (*big.Int, error) {
	amount, err := ConvertToBigInt(value, base)
	if err != nil {
		//this.Log.Errorf("format transaction value failed, err = %v", err)
		return big.NewInt(0), err
	}

	switch unit {
	case TRANS_AMOUNT_UNIT_WEI:
	case TRANS_AMOUNT_UNIT_K_WEI:
		amount.Mul(amount, big.NewInt(1000))
	case TRANS_AMOUNT_UNIT_M_WEI:
		amount.Mul(amount, big.NewInt(1000*1000))
	case TRANS_AMOUNT_UNIT_G_WEI:
		amount.Mul(amount, big.NewInt(1000*1000*1000))
	case TRANS_AMOUNT_UNIT_MICRO_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000))
	case TRANS_AMOUNT_UNIT_MILLIE_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000*1000))
	case TRNAS_AMOUNT_UNIT_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000*1000*1000))
	default:
		return big.NewInt(0), errors.New("wrong unit inputed")
	}

	return amount, nil
}

func (this *WalletManager) GetLocalBlockHeight() (uint64, error) {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for get local block height failed, err=%v", err)
		return 0, err
	}
	defer db.Close()
	var blockHeight uint64
	err = db.Get(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &blockHeight)
	if err != nil {
		this.Log.Errorf("get block height from db failed, err=%v", err)
		return 0, err
	}
	// blockHeight, err := ConvertToUint64(blockHeightStr, 16) //ConvertToBigInt(blockHeightStr, 16)
	// if err != nil {
	// 	this.Log.Errorf("convert block height string failed, err=%v", err)
	// 	return 0, err
	// }
	return blockHeight, nil
}

func (this *WalletManager) SaveLocalBlockScanned(blockHeight uint64, blockHash string) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for update local block height failed, err=%v", err)
		return err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		this.Log.Errorf("start transaction for save block scanned failed, err=%v", err)
		return err
	}
	defer tx.Rollback()

	//blockHeightStr := "0x" + strconv.FormatUint(blockHeight, 16) //blockHeight.Text(16)
	err = tx.Set(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &blockHeight)
	if err != nil {
		this.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	err = tx.Set(BLOCK_CHAIN_BUCKET, BLOCK_HASH_KEY, &blockHash)
	if err != nil {
		this.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	tx.Commit()
	return nil
}

func (this *WalletManager) UpdateLocalBlockHeight(blockHeight uint64) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for update local block height failed, err=%v", err)
		return err
	}
	defer db.Close()

	//blockHeightStr := "0x" + strconv.FormatUint(blockHeight, 16) //blockHeight.Text(16)
	err = db.Set(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &blockHeight)
	if err != nil {
		this.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	return nil
}

func (this *WalletManager) RecoverBlockHeader(height uint64) (*EthBlock, error) {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for save block failed, err=%v", err)
		return nil, err
	}
	defer db.Close()
	var block EthBlock

	err = db.One("BlockNumber", "0x"+strconv.FormatUint(height, 16), &block.BlockHeader)
	if err != nil {
		this.Log.Errorf("get block failed, block number=%v, err=%v", "0x"+strconv.FormatUint(height, 16), err)
		return nil, err
	}

	block.blockHeight, err = ConvertToUint64(block.BlockNumber, 16) //ConvertToBigInt(block.BlockNumber, 16)
	if err != nil {
		this.Log.Errorf("conver block height to big int failed, err= %v", err)
		return nil, err
	}
	return &block, nil
}

func (this *WalletManager) SaveBlockHeader(block *EthBlock) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()
	err = db.Save(&block.BlockHeader)
	if err != nil {
		this.Log.Errorf("save block failed, err = %v", err)
		return err
	}
	return nil
}

func (this *WalletManager) SaveBlockHeader2(block *EthBlock) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		this.Log.Errorf("start transaction for save block header failed, err=%v", err)
		return err
	}
	defer tx.Rollback()

	err = tx.Save(&block.BlockHeader)
	if err != nil {
		this.Log.Errorf("save block failed, err = %v", err)
		return err
	}

	//blockHeightStr := "0x" + strconv.FormatUint(block.blockHeight, 16) //block.blockHeight.Text(16)
	err = tx.Set(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &block.blockHeight)
	if err != nil {
		this.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	err = tx.Set(BLOCK_CHAIN_BUCKET, BLOCK_HASH_KEY, &block.BlockHash)
	if err != nil {
		this.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	tx.Commit()
	return nil
}

/*func (this *WalletManager) SaveTransaction(tx *BlockTransaction) error {
	db, err := OpenDB(DbPath, BLOCK_CHAIN_DB)
	if err != nil {
		this.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()

	err = db.Save(tx)
	if err != nil {
		this.Log.Errorf("save block transaction failed, err = %v", err)
		return err
	}
	return nil
}*/

func (this *WalletManager) RecoverUnscannedTransactions(unscannedTxs []UnscanTransaction) ([]BlockTransaction, error) {
	allTxs := make([]BlockTransaction, 0, len(unscannedTxs))
	for i, _ := range unscannedTxs {
		var tx BlockTransaction
		err := json.Unmarshal([]byte(unscannedTxs[i].TxSpec), &tx)
		if err != nil {
			this.Log.Errorf("decode json [%v] from unscanned transactions failed, err=%v", unscannedTxs[i].TxSpec, err)
			return nil, err
		}
		allTxs = append(allTxs, tx)
	}
	return allTxs, nil
}

func (this *WalletManager) GetAllUnscannedTransactions() ([]UnscanTransaction, error) {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for save block failed, err=%v", err)
		return nil, err
	}
	defer db.Close()

	var allRecords []UnscanTransaction
	err = db.All(&allRecords)
	if err != nil {
		this.Log.Errorf("get all unscanned transactions failed, err = %v", err)
		return nil, err
	}

	return allRecords, nil
}

func (this *WalletManager) DeleteUnscannedTransactions(list []UnscanTransaction) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		log.Errorf("start transaction failed, err=%v", err)
		return err
	}
	defer tx.Rollback()

	for i, _ := range list {
		err = tx.DeleteStruct(&list[i])
		if err != nil {
			log.Errorf("delete unscanned tx faled, err= %v", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error("commit failed, err=%v", err)
		return err
	}
	return nil
}

func (this *WalletManager) DeleteUnscannedTransactionByHeight(height uint64) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()

	var list []UnscanTransaction
	heightStr := "0x" + strconv.FormatUint(height, 16)
	err = db.Find("BlockNumber", heightStr, &list)
	if err != nil && err != storm.ErrNotFound {
		this.Log.Errorf("find unscanned tx failed, block height=%v, err=%v", heightStr, err)
		return err
	} else if err == storm.ErrNotFound {
		this.Log.Infof("no unscanned tx found in block [%v]", heightStr)
		return nil
	}

	for _, r := range list {
		err = db.DeleteStruct(&r)
		if err != nil {
			this.Log.Errorf("delete unscanned tx faled, block height=%v, err=%v", heightStr, err)
			return err
		}
	}
	return nil
}

func (this *WalletManager) SaveUnscannedTransaction(tx *BlockTransaction, reason string) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		this.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()

	txSpec, _ := json.Marshal(tx)

	unscannedRecord := &UnscanTransaction{
		TxID:        tx.Hash,
		BlockNumber: tx.BlockNumber,
		BlockHash:   tx.BlockHash,
		TxSpec:      string(txSpec),
		Reason:      reason,
	}
	err = db.Save(unscannedRecord)
	if err != nil {
		this.Log.Errorf("save unscanned record failed, err=%v", err)
		return err
	}
	return nil
}

//GetAssetsLogger 获取资产账户日志工具
func (this *WalletManager) GetAssetsLogger() *log.OWLogger {
	return this.Log
}
