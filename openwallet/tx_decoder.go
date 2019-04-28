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

//TransactionDecoder 交易单解析器
type TransactionDecoder interface {
	//SendRawTransaction 广播交易单
	//SendTransaction func(amount, feeRate string, to []string, wallet *Wallet, account *AssetsAccount) (*RawTransaction, error)
	//CreateRawTransaction 创建交易单
	CreateRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error
	//SignRawTransaction 签名交易单
	SignRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error
	//SubmitRawTransaction 广播交易单
	SubmitRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) (*Transaction, error)
	//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
	VerifyRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error
	//GetRawTransactionFeeRate 获取交易单的费率
	GetRawTransactionFeeRate() (feeRate string, unit string, err error)
	//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
	CreateSummaryRawTransaction(wrapper WalletDAI, sumRawTx *SummaryRawTransaction) ([]*RawTransaction, error)
	//EstimateRawTransactionFee 预估手续费
	EstimateRawTransactionFee(wrapper WalletDAI, rawTx *RawTransaction) error
	//CreateSummaryRawTransactionWithError 创建汇总交易，返回能原始交易单数组（包含带错误的原始交易单）
	CreateSummaryRawTransactionWithError(wrapper WalletDAI, sumRawTx *SummaryRawTransaction) ([]*RawTransactionWithError, error)
}

//TransactionDecoderBase 实现TransactionDecoder的基类
type TransactionDecoderBase struct {
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoderBase) CreateRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error {
	return fmt.Errorf("not implement")
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoderBase) SignRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error {
	return fmt.Errorf("not implement")
}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoderBase) SubmitRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) (*Transaction, error) {
	return nil, fmt.Errorf("not implement")
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoderBase) VerifyRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error {
	return fmt.Errorf("not implement")
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoderBase) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	return "", "", fmt.Errorf("not implement")
}

//EstimateRawTransactionFee 预估手续费
func (decoder *TransactionDecoderBase) EstimateRawTransactionFee(wrapper WalletDAI, rawTx *RawTransaction) error {
	return fmt.Errorf("EstimateRawTransactionFee not implement")
}

//CreateSummaryRawTransaction 创建汇总交易
func (decoder *TransactionDecoderBase) CreateSummaryRawTransaction(wrapper WalletDAI, sumRawTx *SummaryRawTransaction) ([]*RawTransaction, error) {
	return nil, fmt.Errorf("CreateSummaryRawTransaction not implement")
}

//CreateSummaryRawTransactionWithError 创建汇总交易，返回能原始交易单数组（包含带错误的原始交易单）
func (decoder *TransactionDecoderBase) CreateSummaryRawTransactionWithError(wrapper WalletDAI, sumRawTx *SummaryRawTransaction) ([]*RawTransactionWithError, error) {

	//默认兼容CreateSummaryRawTransaction
	rawTxArrayWithErr := make([]*RawTransactionWithError, 0)
	rawTxArray, err := decoder.CreateSummaryRawTransaction(wrapper, sumRawTx)
	if len(rawTxArray) > 0 {
		for _, rawTx := range rawTxArray {
			rawTxWithErr := &RawTransactionWithError{
				RawTx: rawTx,
				Error: nil,
			}
			rawTxArrayWithErr = append(rawTxArrayWithErr, rawTxWithErr)
		}
	}
	return rawTxArrayWithErr, err
}
