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

package tron

import (
	"errors"
	"fmt"
	"strconv"

	// "sort"

	// "github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	// "github.com/blocktree/OpenWallet/assets/qtum/btcLikeTxDriver"
	// "github.com/blocktree/OpenWallet/log"
	// "github.com/shopspring/decimal"
)

//TransactionDecoder for Interface TransactionDecode
type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	// 检测 wrapper 参数是否满足构建交易单需求

	// ------------------------ 检测 rawTx 参数是否满足构建交易单需求 -------------------------------------
	if err := decoder.wm.LoadConfig(); err != nil {
		// decoder.wm.Log.Std.Error(string(err))
		return err
	}

	if decoder.wm.Config.IsTestNet == true {
		return errors.New("Testnet not support")
	}
	if rawTx.Coin.Symbol != "TRON" {
		return errors.New("CreateRawTransaction: Symbol is not <Tron>")
	}

	var toAddress string
	var amount float64
	if len(rawTx.To) == 0 {
		return errors.New("CreateRawTransaction: Receiver addresses is empty")
	}
	for addr, v := range rawTx.To {
		toAddress = addr
		// amount = int64(numb)
		if fv, err := strconv.ParseFloat(v, 64); err != nil {
			return err
		} else {
			amount = fv
		}

	}

	var ownerAddress string
	if rawTx.Account.AccountID == "" {
		return errors.New("CreateRawTransaction: AccountID is empty")
	}
	if addressList, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID); err != nil {
		return err
	} else {
		if len(addressList) == 0 {
			return fmt.Errorf("[%s] account: %s has not addresses", decoder.wm.Symbol(), rawTx.Account.AccountID)
		}

		ownerAddress = addressList[0].Address

		// // Check balance
		// if act, err := decoder.wm.GetAccount(ownerAddress); err != nil {
		// 	log.Println(err)
		// 	return err
		// } else {
		// 	if act.Balance == "" {
		// 		return errors.New("Balance not enough")
		// 	}

		// 	if balance, err := strconv.ParseFloat(act.Balance, 64); err != nil {
		// 		return err
		// 	} else {
		// 		if balance < amount*1000000 {
		// 			return errors.New("Balance not enough")
		// 		}
		// 	}
		// }

	}

	// --------------------------- 查转出地址和余额 --------------------------------------

	if len(rawTx.To) == 0 {
		return errors.New("Receiver addresses is empty")
	}

	rawHex, err := decoder.wm.CreateTransactionRef(toAddress, ownerAddress, amount)
	if err != nil {
		return err
	}

	rawTx.Fees = "0"
	rawTx.FeeRate = "0"
	rawTx.RawHex = rawHex
	rawTx.Signatures[rawTx.Account.AccountID] = nil //500
	rawTx.IsBuilt = true

	return nil
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return nil
}

//SubmitRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return nil
}
