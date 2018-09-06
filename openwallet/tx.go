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

import (
	"github.com/tidwall/gjson"
)

//TransactionDecoder 交易单解析器
type TransactionDecoder interface {

	//SendRawTransaction 广播交易单
	//SendTransaction func(amount, feeRate string, to []string, wallet *Wallet, account *AssetsAccount) (*RawTransaction, error)
	//CreateRawTransaction 创建交易单
	CreateRawTransaction(wrapper *WalletWrapper, rawTx *RawTransaction) error
	//SignRawTransaction 签名交易单
	SignRawTransaction(wrapper *WalletWrapper, rawTx *RawTransaction) error
	//SendRawTransaction 广播交易单
	SubmitRawTransaction(wrapper *WalletWrapper, rawTx *RawTransaction) error
	//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
	VerifyRawTransaction(wrapper *WalletWrapper, rawTx *RawTransaction) error
}

//RawTransaction 原始交易单
type RawTransaction struct {
	Coin        Coin                      //区块链类型标识
	TxID        string                    //交易单ID，广播后会生成
	RawHex      string                    //区块链协议构造的交易原生数据
	Amount      string                    //转账数量
	FeeRate     string                    //自定义费率
	To          []string                  //目的地址
	Account     *AssetsAccount            //创建交易单的账户
	Signatures  map[string][]KeySignature //拥有者accountID: []未花签名
	Required    uint64                    //必要签名
	IsBuilt     bool                      //是否完成构建建议单
	IsCompleted bool                      //是否完成所有签名
	IsSubmit    bool                      //是否已广播
}

//KeySignature 签名信息
type KeySignature struct {
	EccType    uint32   //曲线类型
	Address    *Address //提供签名的地址
	Signatures string   //未花签名
	Message    string   //被签消息
}

type Transaction struct {
	TxID        string   `json:"txid"`
	AccountID   string `json:"accountID"`
	Address     string `json:"address"`
	Coin        Coin     //区块链类型标识
	From        []string `json:"from"`
	To          []string `json:"to"`
	Amount      string   `json:"amount"`
	TxType      uint64
	Confirm     int64  `json:"confirm"`
	BlockHash   string `json:"blockHash"`
	BlockHeight uint64 `json:"blockHeight"`
	IsMemo      bool   `json:"isMemo"`
	Memo        string `json:"memo"`
	Received    bool
	SubmitTime  int64
	ConfirmTime int64
}

type Recharge struct {
	Sid         string `json:"sid"  storm:"id"` // base64(sha1(txid+n+addr))
	TxID        string `json:"txid"`
	AccountID   string `json:"accountID"`
	Address     string `json:"address"`
	Symbol      string `json:"symbol"` //Deprecated: use Coin
	Coin        Coin   //区块链类型标识
	Amount      string `json:"amount"`
	Confirm     int64  `json:"confirm"`
	BlockHash   string `json:"blockHash"`
	BlockHeight uint64 `json:"blockHeight" storm:"index"`
	IsMemo      bool   `json:"isMemo"`
	Memo        string `json:"memo"`
	Index       uint64 `json:"index"`
	Received    bool
	CreateAt    int64 `json:"createdAt"`
	Delete      bool
}

// TxInput 交易输入，则出账记录
type TxInput struct {
	Recharge
}

// TxOutPut 交易输出，则到账记录
type TxOutPut struct {
	Recharge
}

type Withdraw struct {
	Symbol   string `json:"coin"`
	WalletID string `json:"walletID"`
	Sid      string `json:"sid"  storm:"id"`
	IsMemo   bool   `json:"isMemo"`
	Address  string `json:"address"`
	Amount   string `json:"amount"`
	Memo     string `json:"memo"`
	Password string `json:"password"`
	TxID     string `json:"txid"`
}

//NewWithdraw 创建提现单
func NewWithdraw(json gjson.Result) *Withdraw {
	w := &Withdraw{}
	//解析json
	w.Symbol = gjson.Get(json.Raw, "coin").String()
	w.WalletID = gjson.Get(json.Raw, "walletID").String()
	w.Sid = gjson.Get(json.Raw, "sid").String()
	w.IsMemo = gjson.Get(json.Raw, "isMemo").Bool()
	w.Address = gjson.Get(json.Raw, "address").String()
	w.Amount = gjson.Get(json.Raw, "amount").String()
	w.Memo = gjson.Get(json.Raw, "memo").String()
	w.Password = gjson.Get(json.Raw, "password").String()
	return w
}
