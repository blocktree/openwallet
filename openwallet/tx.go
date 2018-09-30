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
	"encoding/base64"
	"fmt"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/tidwall/gjson"
	"github.com/blocktree/OpenWallet/log"
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
	Coin        Coin                       `json:"coin"`       //区块链类型标识
	TxID        string                     `json:"txID"`       //交易单ID，广播后会生成
	Sid         string                     `json:"sid"`        //业务订单号，保证业务不重复交易而用
	RawHex      string                     `json:"rawHex"`     //区块链协议构造的交易原生数据
	FeeRate     string                     `json:"feeRate"`    //自定义费率
	To          map[string]string          `json:"to"`         //目的地址:转账数量
	Account     *AssetsAccount             `json:"account"`    //创建交易单的账户
	Signatures  map[string][]*KeySignature `json:"sigParts"`   //拥有者accountID: []未花签名
	Required    uint64                     `json:"reqSigs"`    //必要签名
	IsBuilt     bool                       `json:"isBuilt"`    //是否完成构建建议单
	IsCompleted bool                       `json:"isComplete"` //是否完成所有签名
	IsSubmit    bool                       `json:"isSubmit"`   //是否已广播
	Change      *Address                   `json:"change"`     //找零地址
	ExtParam    string                     `json:"extParam"`   //扩展参数，用于调用智能合约，json结构
}

//KeySignature 签名信息
type KeySignature struct {
	EccType   uint32   `json:"eccType"` //曲线类型
	Nonce     string   `json:"nonce"`
	Address   *Address `json:"address"` //提供签名的地址
	Signature string   `json:"signed"`  //未花签名
	Message   string   `json:"msg"`     //被签消息
}

type Transaction struct {

	//openwallet自定义的ID，在不同链可能存在重复的txid，
	// 所以我们要生成一个全局不重复的
	WxID string `json:"wxid" storm:"id"`

	TxID      string `json:"txid"`
	AccountID string `json:"accountID"`
	//Address     string   `json:"address"`
	Coin        Coin     //区块链类型标识
	From        []string `json:"from"`
	To          []string `json:"to"`
	Amount      string   `json:"amount"`
	Decimal     int32    `json:"decimal"`
	TxType      uint64
	Confirm     int64  `json:"confirm"`
	BlockHash   string `json:"blockHash"`
	BlockHeight uint64 `json:"blockHeight"`
	IsMemo      bool   `json:"isMemo"`
	Memo        string `json:"memo"`
	Fees        string
	Received    bool
	SubmitTime  int64
	ConfirmTime int64
}

//GenTransactionWxID 生成交易单的WxID，格式为 base64(sha1(tx_{txID}_{symbol_contractID}))
func GenTransactionWxID(tx *Transaction) string {
	txid := tx.TxID
	symbol := tx.Coin.Symbol + "_" + tx.Coin.ContractID
	plain := fmt.Sprintf("tx_%s_%s", txid, symbol)
	log.Debug("wxID plain:", plain)
	wxid := base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(plain)))
	return wxid
}

type Recharge struct {
	Sid         string `json:"sid" storm:"id"` // base64(sha1(txid+n+addr))
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
	SourceTxID  string //源交易单ID
	SourceIndex uint64 //源交易单输出所因为
	Recharge    `storm:"inline"`
}

// TxOutPut 交易输出，则到账记录
type TxOutPut struct {
	Recharge `storm:"inline"`
	ExtParam string //扩展参数，用于记录utxo的解锁字段，json格式
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
