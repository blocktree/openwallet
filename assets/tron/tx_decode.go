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

	// "sort"

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

	// 检测 rawTx 参数是否满足构建交易单需求
	if rawTx.Coin.Symbol != "TRON" {
		return errors.New("CreateRawTransaction: Symbol is not <Tron>")
	}

	if len(rawTx.To) == 0 {
		return errors.New("CreateRawTransaction: Receiver addresses is empty")
	}

	if rawTx.Account.AccountID == "" {
		return errors.New("CreateRawTransaction: AccountID is empty")
	}
	accountID := rawTx.Account.AccountID

	// isTestNet := decoder.wm.Config.IsTestNet

	address, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)
	if err != nil {
		return err
	}

	if len(address) == 0 {
		return fmt.Errorf("[%s] account: %s has not addresses", decoder.wm.Symbol(), accountID)
	}
	fmt.Println(address)

	// 查余额

	if len(rawTx.To) == 0 {
		return errors.New("Receiver addresses is empty")
	}

	rawTx.Signatures[rawTx.Account.AccountID] = nil //"keySigs"
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
