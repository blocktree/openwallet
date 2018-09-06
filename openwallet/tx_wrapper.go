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

import "fmt"

// TransactionWrapper 交易包装器，扩展钱包交易单相关功能
type TransactionWrapper struct {
	*WalletWrapper

}

func NewTransactionWrapper(args ...interface{}) *TransactionWrapper {

	wrapper := NewWalletWrapper(args...)

	walletWrapper := TransactionWrapper{WalletWrapper: wrapper}

	for _, arg := range args {
		switch obj := arg.(type) {
		case *Wallet:
			walletWrapper.wallet = obj
		case WalletDBFile:
			walletWrapper.sourceFile = string(obj)
		case WalletKeyFile:
			walletWrapper.keyFile = string(obj)
		case *WalletWrapper:
			walletWrapper.WalletWrapper = obj
		}
	}

	return &walletWrapper
}

//SaveBlockExtractData 保存区块提取数据
func (wrapper *TransactionWrapper) SaveBlockExtractData(data *BlockExtractData) error {

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return err
	}
	defer wrapper.CloseDB()

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	//保存出账的记录
	for _, input := range data.TxInputs {
		err = tx.Save(input)
		if err != nil {
			return fmt.Errorf("wallet save TxInputs failed, unexpected error: %v", err)
		}
	}

	//保存入账的记录
	for _, output := range data.TxOutputs {
		err = tx.Save(output)
		if err != nil {
			return fmt.Errorf("wallet save TxOutputs failed, unexpected error: %v", err)
		}
	}

	//保存入账的记录
	for _, trx := range data.Transactions {
		err = tx.Save(trx)
		if err != nil {
			return fmt.Errorf("wallet save Transactions failed, unexpected error: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("wallet save BlockExtractData failed, unexpected error: %v", err)
	}

	return nil
}