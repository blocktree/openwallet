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
	"encoding/json"
	"fmt"

	"github.com/blocktree/OpenWallet/crypto"
	"github.com/tidwall/gjson"
)

//TransactionDecoder 交易单解析器
type TransactionDecoder interface {
	//SendRawTransaction 广播交易单
	//SendTransaction func(amount, feeRate string, to []string, wallet *Wallet, account *AssetsAccount) (*RawTransaction, error)
	//CreateRawTransaction 创建交易单
	CreateRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error
	//SignRawTransaction 签名交易单
	SignRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error
	//SendRawTransaction 广播交易单
	SubmitRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error
	//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
	VerifyRawTransaction(wrapper WalletDAI, rawTx *RawTransaction) error
	//GetRawTransactionFeeRate 获取交易单的费率
	GetRawTransactionFeeRate() (feeRate string, unit string, err error)
}

//TransactionDecoderBase 实现TransactionDecoder的基类
type TransactionDecoderBase struct {
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoderBase) CreateRawTransaction(wrapper *WalletWrapper, rawTx *RawTransaction) error {
	return fmt.Errorf("not implement")
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoderBase) SignRawTransaction(wrapper *WalletWrapper, rawTx *RawTransaction) error {
	return fmt.Errorf("not implement")
}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoderBase) SubmitRawTransaction(wrapper *WalletWrapper, rawTx *RawTransaction) error {
	return fmt.Errorf("not implement")
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoderBase) VerifyRawTransaction(wrapper *WalletWrapper, rawTx *RawTransaction) error {
	return fmt.Errorf("not implement")
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoderBase) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	return "", "", fmt.Errorf("not implement")
}

//RawTransaction 原始交易单
//
// Workflow：
// 	首先：App(openw-server)中，提供 Coin/To/Account 参数
// 	其次：APP 调用 tx_decoder 需处理：
// 		1. 第一步调用交易单构建：计算手续费/构建签名结构，给 RawHex，Fees，FeeRate，Signatures，IsBuilt = true 赋值
// 		2. 第二步调用交易单签名：获取到 RawHex 完成签名，完成参数 Signatures
// 		3. 第三步调用交易单广播：解析rawHex，合并签名，验证签名，广播交易, 设置参数 TxID，IsSubmit = true
type RawTransaction struct {
	Coin        Coin                       `json:"coin"`       //@required 区块链类型标识
	TxID        string                     `json:"txID"`       //交易单ID，广播后会生成
	Sid         string                     `json:"sid"`        //业务订单号，保证业务不重复交易而用
	RawHex      string                     `json:"rawHex"`     //区块链协议构造的交易原生数据
	FeeRate     string                     `json:"feeRate"`    //自定义费率
	To          map[string]string          `json:"to"`         //@required 目的地址:转账数量
	Account     *AssetsAccount             `json:"account"`    //@required 创建交易单的账户
	Signatures  map[string][]*KeySignature `json:"sigParts"`   //拥有者accountID: []未花签名
	Required    uint64                     `json:"reqSigs"`    //必要签名
	IsBuilt     bool                       `json:"isBuilt"`    //是否完成构建建议单
	IsCompleted bool                       `json:"isComplete"` //是否完成所有签名
	IsSubmit    bool                       `json:"isSubmit"`   //是否已广播
	Change      *Address                   `json:"change"`     //找零地址
	ExtParam    string                     `json:"extParam"`   //扩展参数，用于调用智能合约，json结构
	Fees        string                     `json:"fees"`       //手续费
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
	WxID        string   `json:"wxid" storm:"id"` //@required 通过GenTransactionWxID计算
	TxID        string   `json:"txid"`            //@required
	AccountID   string   `json:"accountID"`
	Coin        Coin     `json:"coin"` //@required 区块链类型标识
	From        []string `json:"from"` //@required 格式："地址":"数量"
	To          []string `json:"to"`   //@required 格式："地址":"数量"
	Amount      string   `json:"amount"`
	Decimal     int32    `json:"decimal"` //@required
	TxType      uint64   `json:"txType"`  // 0:未知，1:主币转账，2:合约调用
	Confirm     int64    `json:"confirm"`
	BlockHash   string   `json:"blockHash"`   //@required
	BlockHeight uint64   `json:"blockHeight"` //@required
	IsMemo      bool     `json:"isMemo"`
	Memo        string   `json:"memo"`
	Fees        string   `json:"fees"` //@required
	Received    bool     `json:"received"`
	SubmitTime  int64    `json:"submitTime"`  //@required
	ConfirmTime int64    `json:"confirmTime"` //@required
	Status      string   `json:"status"`      //链上状态
	Reason      string   `json:"reason"`      //失败原因
}

//GenTransactionWxID 生成交易单的WxID，格式为 base64(sha1(tx_{txID}_{symbol}_contractID}))

func GenTransactionWxID2(txid string, coinsymbol string, contractId string) string {
	//txid := tx.TxID
	symbol := coinsymbol + "_" + contractId
	plain := fmt.Sprintf("tx_%s_%s", txid, symbol)
	//log.Debug("wxID plain:", plain)
	wxid := base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(plain)))
	return wxid
}

func GenTransactionWxID(tx *Transaction) string {
	//txid := tx.TxID
	//symbol := tx.Coin.Symbol + "_" + tx.Coin.ContractID
	//plain := fmt.Sprintf("tx_%s_%s", txid, symbol)
	//log.Debug("wxID plain:", plain)
	//wxid := base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(plain)))
	//return wxid
	return GenTransactionWxID2(tx.TxID, tx.Coin.Symbol, tx.Coin.ContractID)
}

type Recharge struct {
	Sid         string `json:"sid" storm:"id"` //@required base64(sha1(txid+n+addr))，对于账户模型，只有一个输入输出，n = 0。
	TxID        string `json:"txid"`           //@required
	AccountID   string `json:"accountID"`
	Address     string `json:"address"` //@required
	Symbol      string `json:"symbol"`  //Deprecated: use Coin
	Coin        Coin   //@required 区块链类型标识
	Amount      string `json:"amount"` //@required
	Confirm     int64  `json:"confirm"`
	BlockHash   string `json:"blockHash"`                 //@required
	BlockHeight uint64 `json:"blockHeight" storm:"index"` //@required
	IsMemo      bool   `json:"isMemo"`
	Memo        string `json:"memo"`
	Index       uint64 `json:"index"` //@required
	Received    bool
	CreateAt    int64 `json:"createdAt"` //@required
	Delete      bool
}

//GenRechargeSID
func GenRechargeSID(txid string, coinsymbol string, contractId string, n uint64, prefix string) string {
	//txid := tx.TxID
	symbol := coinsymbol + "_" + contractId
	plain := fmt.Sprintf("%s_%s_%s_%d", prefix, txid, symbol, n)
	sid := base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(plain)))
	return sid
}

// TxInput 交易输入，则出账记录
type TxInput struct {
	//SourceTxID和SourceIndex是utxo模型上的上一个交易输入源，account模型不需要填
	SourceTxID  string //源交易单ID
	SourceIndex uint64 //源交易单输出所因为
	Recharge    `storm:"inline"`
}

func GenTxInputSID(txid string, coinsymbol string, contractId string, n uint64) string {
	return GenRechargeSID(txid, coinsymbol, contractId, n, "input")
}

// TxOutPut 交易输出，则到账记录
type TxOutPut struct {
	Recharge `storm:"inline"`
	ExtParam string //扩展参数，用于记录utxo的解锁字段，json格式
}

func GenTxOutPutSID(txid string, coinsymbol string, contractId string, n uint64) string {
	return GenRechargeSID(txid, coinsymbol, contractId, n, "output")
}

//SetExtParam
func (txOut *TxOutPut) SetExtParam(key string, value interface{}) error {
	var ext map[string]interface{}

	if len(txOut.ExtParam) == 0 {
		ext = make(map[string]interface{})
	} else {
		err := json.Unmarshal([]byte(txOut.ExtParam), &ext)
		if err != nil {
			return err
		}
	}

	ext[key] = value

	json, err := json.Marshal(ext)
	if err != nil {
		return err
	}
	txOut.ExtParam = string(json)

	return nil
}

//GetExtParam
func (txOut *TxOutPut) GetExtParam() gjson.Result {
	//如果param没有值，使用inputs初始化
	return gjson.ParseBytes([]byte(txOut.ExtParam))
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
