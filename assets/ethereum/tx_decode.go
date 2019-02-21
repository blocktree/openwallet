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
package ethereum

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"

	"github.com/tidwall/gjson"

	//"log"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcrypt"
	"github.com/bytom/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type EthTxExtPara struct {
	Data     string `json:"data"`
	GasLimit string `json:"gasLimit"`
}

func NewEthTxExtPara(j gjson.Result) *EthTxExtPara {
	obj := EthTxExtPara{}
	obj.GasLimit = j.Get("gasLimit").String()
	obj.Data = j.Get("data").String()
	return &obj
}

/*func (this *EthTxExtPara) GetGasLimit() (uint64, error) {
	gasLimit, err := strconv.ParseUint(removeOxFromHex(this.GasLimit), 16, 64)
	if err != nil {
		this.wm.Log.Std.Error("parse gas limit to uint64 failed, err=%v", err)
		return 0, err
	}
	return gasLimit, nil
}*/

const (
	ADRESS_STATIS_OVERDATED_TIME = 30
)

type AddressTxStatistic struct {
	Address          string
	TransactionCount *uint64
	LastModifiedTime *time.Time
	Valid            *int //如果valid指针指向的整形为0, 说明该地址已经被清理线程清理
	AddressLocker    *sync.Mutex
	//1. 地址级别, 不可并发广播交易, 造成nonce混乱
	//2. 地址级别, 不可广播, 读取nonce同时进行, 会造成nonce混乱
}

func (this *AddressTxStatistic) UpdateTime() {
	now := time.Now()
	this.LastModifiedTime = &now
}

type EthTransactionDecoder struct {
	openwallet.TransactionDecoderBase
	AddrTxStatisMap *sync.Map
	//	DecoderLocker *sync.Mutex    //保护一些全局不可并发的操作, 如对AddrTxStatisMap的初始化
	wm *WalletManager //钱包管理者
}

func (this *EthTransactionDecoder) GetTransactionCount2(address string) (*AddressTxStatistic, uint64, error) {
	now := time.Now()
	valid := 1
	t := AddressTxStatistic{
		LastModifiedTime: &now,
		AddressLocker:    new(sync.Mutex),
		Valid:            &valid,
		Address:          address,
		TransactionCount: new(uint64),
	}

	v, loaded := this.AddrTxStatisMap.LoadOrStore(address, t)
	//LoadOrStore返回后, AddressLocker加锁前, map中的nonce可能已经被清理了, 需要检查valid是否为1
	txStatis := v.(AddressTxStatistic)
	txStatis.AddressLocker.Lock()
	txStatis.AddressLocker.Unlock()
	if loaded {
		if *txStatis.Valid == 0 {
			return nil, 0, errors.New("the node is busy, try it again later. ")
		}
		txStatis.UpdateTime()
		return &txStatis, *txStatis.TransactionCount, nil
	}
	nonce, err := this.wm.GetNonceForAddress2(appendOxToAddress(address))
	if err != nil {
		this.wm.Log.Std.Error("get nonce for address via rpc failed, err=%v", err)
		return nil, 0, err
	}
	*txStatis.TransactionCount = nonce
	return &txStatis, *txStatis.TransactionCount, nil
}

func (this *EthTransactionDecoder) GetTransactionCount(address string) (uint64, error) {
	if this.AddrTxStatisMap == nil {
		return 0, errors.New("map should be initialized before using.")
	}

	v, exist := this.AddrTxStatisMap.Load(address)
	if !exist {
		return 0, errors.New("no records found to the key passed through.")
	}

	txStatis := v.(AddressTxStatistic)
	return *txStatis.TransactionCount, nil
}

func (this *EthTransactionDecoder) SetTransactionCount(address string, transactionCount uint64) error {
	if this.AddrTxStatisMap == nil {
		return errors.New("map should be initialized before using.")
	}

	v, exist := this.AddrTxStatisMap.Load(address)
	if !exist {
		return errors.New("no records found to the key passed through.")
	}

	now := time.Now()
	valid := 1
	txStatis := AddressTxStatistic{
		TransactionCount: &transactionCount,
		LastModifiedTime: &now,
		AddressLocker:    new(sync.Mutex),
		Valid:            &valid,
		Address:          address,
	}

	if exist {
		txStatis.AddressLocker = v.(AddressTxStatistic).AddressLocker
	} else {
		txStatis.AddressLocker = &sync.Mutex{}
	}

	this.AddrTxStatisMap.Store(address, txStatis)
	return nil
}

func (this *EthTransactionDecoder) RemoveOutdatedAddrStatic() {
	addrStatisList := make([]AddressTxStatistic, 0)
	this.AddrTxStatisMap.Range(func(k, v interface{}) bool {
		addrStatis := v.(AddressTxStatistic)
		if addrStatis.LastModifiedTime.Before(time.Now().Add(-1 * (ADRESS_STATIS_OVERDATED_TIME * time.Minute))) {
			addrStatisList = append(addrStatisList, addrStatis)
		}
		return true
	})

	clear := func(statis *AddressTxStatistic) {
		statis.AddressLocker.Lock()
		defer statis.AddressLocker.Unlock()
		if statis.LastModifiedTime.Before(time.Now().Add(-1 * (ADRESS_STATIS_OVERDATED_TIME * time.Minute))) {
			*statis.Valid = 0
			this.AddrTxStatisMap.Delete(statis.Address)
		}
	}

	for i, _ := range addrStatisList {
		clear(&addrStatisList[i])
	}
}

func (this *EthTransactionDecoder) RunClearAddrStatic() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			this.RemoveOutdatedAddrStatic()
		}
	}()
}

func (this *EthTransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	price, err := this.wm.WalletClient.ethGetGasPrice()
	if err != nil {
		this.wm.Log.Errorf("get gas price failed, err=%v", err)
		return "", "Gas", err
	}

	pricedecimal, err := ConverWeiStringToEthDecimal(price.String())
	if err != nil {
		this.wm.Log.Errorf("wrong gas price format.")
	}
	return pricedecimal.String(), "Gas", nil
}

func VerifyRawTransaction(rawTx *openwallet.RawTransaction) error {
	if len(rawTx.To) != 1 {
		//this.wm.Log.Error("only one to address can be set.")
		return errors.New("only one to address can be set.")
	}

	return nil
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *EthTransactionDecoder {
	decoder := EthTransactionDecoder{}
	//	decoder.DecoderLocker = new(sync.Mutex)
	decoder.wm = wm
	decoder.AddrTxStatisMap = new(sync.Map)
	decoder.RunClearAddrStatic()
	return &decoder
}

func (this *EthTransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountID       = rawTx.Account.AccountID
		findAddrBalance *AddrBalance
		feeInfo         *txFeeInfo
	)

	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		return err
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1,
		"AccountID", accountID)
	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		return fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := this.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return err
	}

	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	amount, _ := ConvertEthStringToWei(amountStr)

	//地址余额从大到小排序
	sort.Slice(addrBalanceArray, func(i int, j int) bool {
		a_amount, _ := decimal.NewFromString(addrBalanceArray[i].Balance)
		b_amount, _ := decimal.NewFromString(addrBalanceArray[j].Balance)
		if a_amount.LessThan(b_amount) {
			return true
		} else {
			return false
		}
	})

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI, _ := ConvertEthStringToWei(addrBalance.Balance)

		//计算手续费
		feeInfo, err = this.wm.GetTransactionFeeEstimated(addrBalance.Address, to, amount, "")
		if err != nil {
			this.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, to, err)
			continue
		}

		if rawTx.FeeRate != "" {
			feeInfo.GasPrice, _ = ConvertEthStringToWei(rawTx.FeeRate)
			feeInfo.CalcFee()
		}

		//总消耗数量 = 转账数量 + 手续费
		totalAmount := new(big.Int)
		totalAmount.Add(amount, feeInfo.Fee)

		if addrBalance_BI.Cmp(totalAmount) < 0 {
			continue
		}

		//只要找到一个合适使用的地址余额就停止遍历
		findAddrBalance = &AddrBalance{Address: addrBalance.Address, Balance: addrBalance_BI}
		break
	}

	if findAddrBalance == nil {
		return fmt.Errorf("the balance: %s is not enough", amountStr)
	}

	//最后创建交易单
	err = this.createRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		feeInfo,
		"")
	if err != nil {
		return err
	}

	return err
}

func (this *EthTransactionDecoder) CreateErc20TokenRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountID       = rawTx.Account.AccountID
		findAddrBalance *AddrBalance
		feeInfo         *txFeeInfo
		errStr          string
		callData        string
	)

	tokenDecimals := int(rawTx.Coin.Contract.Decimals)
	contractAddress := rawTx.Coin.Contract.Address

	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		this.wm.Log.Std.Error("Verify raw tx failed, err=%v", err)
		return err
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1,
		"AccountID", accountID)
	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		return fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := this.wm.ContractDecoder.GetTokenBalanceByAddress(rawTx.Coin.Contract, searchAddrs...)
	if err != nil {
		return err
	}

	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	//地址余额从大到小排序
	sort.Slice(addrBalanceArray, func(i int, j int) bool {
		a_amount, _ := decimal.NewFromString(addrBalanceArray[i].Balance.Balance)
		b_amount, _ := decimal.NewFromString(addrBalanceArray[j].Balance.Balance)
		if a_amount.LessThan(b_amount) {
			return true
		} else {
			return false
		}
	})

	for _, addrBalance := range addrBalanceArray {
		callData = ""

		//检查余额是否超过最低转账
		addrBalance_BI, _ := ConvertFloatStringToBigInt(addrBalance.Balance.Balance, tokenDecimals)

		amount, _ := ConvertFloatStringToBigInt(amountStr, tokenDecimals)

		if addrBalance_BI.Cmp(amount) < 0 {
			errStr = fmt.Sprintf("the balance: %s is not enough", amountStr)
			continue
		}

		data, createErr := makeERC20TokenTransData(contractAddress, to, amount)
		if createErr != nil {
			continue
		}

		//this.wm.Log.Debug("sumAmount:", sumAmount)
		//计算手续费
		fee, createErr := this.wm.GetTransactionFeeEstimated(addrBalance.Balance.Address, contractAddress, nil, data)
		if createErr != nil {
			this.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Balance.Address, to, createErr)
			return createErr
		}

		if rawTx.FeeRate != "" {
			fee.GasPrice, _ = ConvertEthStringToWei(rawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
			fee.CalcFee()
		}

		coinBalance, err := this.wm.WalletClient.GetAddrBalance2(appendOxToAddress(addrBalance.Balance.Address), "latest")
		if err != nil {
			continue
		}

		if coinBalance.Cmp(fee.Fee) < 0 {
			coinBalance, _ := ConverWeiStringToEthDecimal(coinBalance.String())
			errStr = fmt.Sprintf("the [%s] balance: %s is not enough to call smart contract", rawTx.Coin.Symbol, coinBalance)
			continue
		}

		//只要找到一个合适使用的地址余额就停止遍历
		findAddrBalance = &AddrBalance{Address: addrBalance.Balance.Address, Balance: coinBalance, TokenBalance: addrBalance_BI}
		feeInfo = fee
		callData = data
		break
	}

	if findAddrBalance == nil {
		return fmt.Errorf(errStr)
	}

	//最后创建交易单
	err = this.createRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		feeInfo,
		callData)
	if err != nil {
		return err
	}

	return nil
}

//CreateRawTransaction 创建交易单
func (this *EthTransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return this.CreateSimpleRawTransaction(wrapper, rawTx)
	}
	return this.CreateErc20TokenRawTransaction(wrapper, rawTx)
}

//SignRawTransaction 签名交易单
func (this *EthTransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		this.wm.Log.Std.Error("Verify raw tx failed, err=%v", err)
		return err
	}

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		this.wm.Log.Error("get HDKey from wallet wrapper failed, err=%v", err)
		return err
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		this.wm.Log.Std.Error("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	if len(rawTx.Signatures[rawTx.Account.AccountID]) != 1 {
		this.wm.Log.Error("signature failed in account[%v].", rawTx.Account.AccountID)
		return errors.New("signature failed in account.")
	}

	signnode := rawTx.Signatures[rawTx.Account.AccountID][0]
	fromAddr := signnode.Address

	childKey, _ := key.DerivedKeyWithPath(fromAddr.HDPath, owcrypt.ECC_CURVE_SECP256K1)
	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		this.wm.Log.Error("get private key bytes, err=", err)
		return err
	}
	//prikeyStr := common.ToHex(keyBytes)
	//this.wm.Log.Debugf("pri:%v", common.ToHex(keyBytes))

	message, err := hex.DecodeString(signnode.Message)
	if err != nil {
		return err
	}
	//seckey := math.PaddedBigBytes(key.PrivateKey.D, key.PrivateKey.Params().BitSize/8)

	/*sig, err := secp256k1.Sign(message, keyBytes)
	if err != nil {
		this.wm.Log.Error("secp256k1.Sign failed, err=", err)
		return err
	}*/
	sig, ret := owcrypt.ETHsignature(keyBytes, message)
	if ret != owcrypt.SUCCESS {
		errdesc := fmt.Sprintln("signature error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		this.wm.Log.Error(errdesc)
		return errors.New(errdesc)
	}

	signnode.Signature = hex.EncodeToString(sig)

	//this.wm.Log.Debug("** pri:", hex.EncodeToString(keyBytes))
	this.wm.Log.Debug("** message:", signnode.Message)
	this.wm.Log.Debug("** Signature:", signnode.Signature)

	return nil
}

func (this *EthTransactionDecoder) SubmitSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		this.wm.Log.Std.Error("Verify raw tx failed, err=%v", err)
		return nil, err
	}
	if len(rawTx.Signatures) != 1 {
		this.wm.Log.Std.Error("len of signatures error. ")
		return nil, errors.New("len of signatures error. ")
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		this.wm.Log.Std.Error("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return nil, errors.New("wallet signature not found ")
	}

	from := rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address
	sig := rawTx.Signatures[rawTx.Account.AccountID][0].Signature

	this.wm.Log.Debug("rawTx.ExtParam:", rawTx.ExtParam)
	//extPara := NewEthTxExtPara(gjson.Parse(rawTx.ExtParam))
	//err = json.Unmarshal([]byte(rawTx.ExtParam), &extPara)
	//if err != nil {
	//	this.wm.Log.Error("decode json from extpara failed, err=%v", err)
	//	return err
	//}

	signer := types.NewEIP155Signer(big.NewInt(int64(this.wm.GetConfig().ChainID)))

	//var to, amountStr string
	//for k, v := range rawTx.To {
	//	to = k
	//	amountStr = v
	//	break
	//}
	//amount, err := ConvertEthStringToWei(amountStr) //ConvertToBigInt(amountStr, 10)
	//if err != nil {
	//	this.wm.Log.Std.Error("amount convert to big int failed, err=%v", err)
	//	return err
	//}

	txStatis, _, err := this.GetTransactionCount2(from)
	if err != nil {
		this.wm.Log.Std.Error("get transaction count2 failed, err=%v", err)
		return nil, errors.New("get transaction count2 faile")
	}

	//this.wm.Log.Debug("extPara.GasLimit:", extPara.GasLimit)
	//gaslimit, err := ConvertEthStringToWei(extPara.GasLimit) //extPara.GetGasLimit()
	//if err != nil {
	//	this.wm.Log.Std.Error("get gas limit failed, err=%v", err)
	//	return errors.New("get gas limit failed")
	//}

	//gasPrice, err := ConvertEthStringToWei(rawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
	//if err != nil {
	//	this.wm.Log.Std.Error("get gas price failed, err=%v", err)
	//	return errors.New("get gas price failed")
	//}

	rawHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		this.wm.Log.Error("rawTx.RawHex decode failed, err:", err)
		return nil, err
	}

	err = func() error {
		txStatis.AddressLocker.Lock()
		defer txStatis.AddressLocker.Unlock()
		//nonceSigned, err := strconv.ParseUint(removeOxFromHex(rawTx.Signatures[rawTx.Account.AccountID][0].Nonce),
		//	16, 64)
		//if err != nil {
		//	this.wm.Log.Std.Error("parse nonce from rawTx failed, err=%v", err)
		//	return errors.New("parse nonce from rawTx failed. ")
		//}
		//if nonceSigned != *txStatis.TransactionCount {
		//	this.wm.Log.Std.Error("nonce out of dated, please try to start ur tx once again. ")
		//	return errors.New("nonce out of dated, please try to start ur tx once again. ")
		//}

		tx := &types.Transaction{}
		err = rlp.DecodeBytes(rawHex, tx)
		if err != nil {
			this.wm.Log.Error("transaction RLP decode failed, err:", err)
			return err
		}

		if tx.Nonce() != *txStatis.TransactionCount {
			this.wm.Log.Std.Error("nonce out of dated, please try to start ur tx once again. ")
			return errors.New("nonce out of dated, please try to start ur tx once again. ")
		}

		//tx := types.NewTransaction(nonceSigned, ethcommon.HexToAddress(to),
		//	amount, gaslimit.Uint64(), gasPrice, nil)
		tx, err = tx.WithSignature(signer, common.FromHex(sig))
		if err != nil {
			this.wm.Log.Std.Error("tx with signature failed, err=%v ", err)
			return errors.New("tx with signature failed. ")
		}

		txstr, _ := json.MarshalIndent(tx, "", " ")
		this.wm.Log.Debug("**after signed txStr:", string(txstr))

		rawTxPara, err := rlp.EncodeToBytes(tx)
		if err != nil {
			this.wm.Log.Std.Error("encode tx to rlp failed, err=%v ", err)
			return errors.New("encode tx to rlp failed. ")
		}

		txid, err := this.wm.WalletClient.ethSendRawTransaction(common.ToHex(rawTxPara))
		if err != nil {
			this.wm.Log.Std.Error("sent raw tx faild, err=%v", err)
			return fmt.Errorf("sent raw tx faild. unexpected error: %v", err)
		}

		rawTx.TxID = txid
		rawTx.IsSubmit = true
		txStatis.UpdateTime()
		(*txStatis.TransactionCount)++

		this.wm.Log.Debug("transaction[", txid, "] has been sent out.")
		return nil
	}()

	if err != nil {
		this.wm.Log.Errorf("send raw transaction failed, err= %v", err)
		return nil, err
	}

	decimals := int32(0)
	if rawTx.Coin.IsContract {
		decimals = int32(rawTx.Coin.Contract.Decimals)
	} else {
		decimals = int32(this.wm.Decimal())
	}

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       rawTx.Fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	return tx, nil
}

func (this *EthTransactionDecoder) SubmitErc20TokenRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		this.wm.Log.Std.Error("Verify raw tx failed, err=%v", err)
		return nil, err
	}
	if len(rawTx.Signatures) != 1 {
		this.wm.Log.Std.Error("len of signatures error. ")
		return nil, errors.New("len of signatures error. ")
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		this.wm.Log.Std.Error("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return nil, errors.New("wallet signature not found ")
	}

	from := rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address
	sig := rawTx.Signatures[rawTx.Account.AccountID][0].Signature

	//extPara := NewEthTxExtPara(gjson.Parse(rawTx.ExtParam))
	//var extPara EthTxExtPara
	//this.wm.Log.Debug("rawTx.ExtParam:", rawTx.ExtParam)
	//err = json.Unmarshal([]byte(rawTx.ExtParam), &extPara)
	//if err != nil {
	//	this.wm.Log.Std.Error("decode json from extpara failed, err=%v", err)
	//	return err
	//}

	//data := extPara.Data
	//this.wm.Log.Debug("extPara.GasLimit:", extPara.GasLimit)
	//gaslimit, err := ConvertEthStringToWei(extPara.GasLimit) //extPara.GetGasLimit()
	//if err != nil {
	//	this.wm.Log.Std.Error("get gas limit failed, err=%v", err)
	//	return errors.New("get gas limit failed")
	//}

	signer := types.NewEIP155Signer(big.NewInt(int64(this.wm.GetConfig().ChainID)))

	txStatis, _, err := this.GetTransactionCount2(from)
	if err != nil {
		this.wm.Log.Std.Error("get transaction count2 failed, err=%v", err)
		return nil, errors.New("get transaction count2 faile")
	}

	//gasPrice, err := ConvertEthStringToWei(rawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
	//if err != nil {
	//	this.wm.Log.Std.Error("get gas price failed, err=%v", err)
	//	return errors.New("get gas price failed")
	//}

	rawHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		this.wm.Log.Error("rawTx.RawHex decode failed, err:", err)
		return nil, err
	}

	err = func() error {
		txStatis.AddressLocker.Lock()
		defer txStatis.AddressLocker.Unlock()
		//
		//nonceSigned, err := strconv.ParseUint(removeOxFromHex(rawTx.Signatures[rawTx.Account.AccountID][0].Nonce),
		//	16, 64)
		//if err != nil {
		//	this.wm.Log.Std.Error("parse nonce from rawTx failed, err=%v", err)
		//	return errors.New("parse nonce from rawTx failed. ")
		//}
		//
		//if nonceSigned != *txStatis.TransactionCount {
		//	this.wm.Log.Std.Error("nonce out of dated, please try to start ur tx once again. ")
		//	return errors.New("nonce out of dated, please try to start ur tx once again. ")
		//}

		tx := &types.Transaction{}
		err = rlp.DecodeBytes(rawHex, tx)
		if err != nil {
			this.wm.Log.Error("transaction RLP decode failed, err:", err)
			return err
		}

		if tx.Nonce() != *txStatis.TransactionCount {
			this.wm.Log.Std.Error("nonce out of dated, please try to start ur tx once again. ")
			return errors.New("nonce out of dated, please try to start ur tx once again. ")
		}

		//tx := types.NewTransaction(nonceSigned, ethcommon.HexToAddress(rawTx.Coin.Contract.Address),
		//	big.NewInt(0), gaslimit.Uint64(), gasPrice, common.FromHex(data))
		tx, err = tx.WithSignature(signer, common.FromHex(sig))
		if err != nil {
			this.wm.Log.Std.Error("tx with signature failed, err=%v ", err)
			return errors.New("tx with signature failed. ")
		}

		txstr, _ := json.MarshalIndent(tx, "", " ")
		this.wm.Log.Debug("**after signed txStr:", string(txstr))

		rawTxPara, err := rlp.EncodeToBytes(tx)
		if err != nil {
			this.wm.Log.Std.Error("encode tx to rlp failed, err=%v ", err)
			return errors.New("encode tx to rlp failed. ")
		}

		txid, err := this.wm.WalletClient.ethSendRawTransaction(common.ToHex(rawTxPara))
		if err != nil {
			this.wm.Log.Std.Error("sent raw tx faild, err=%v", err)
			return fmt.Errorf("sent raw tx faild. unexpected error: %v", err)
		}

		rawTx.TxID = txid
		rawTx.IsSubmit = true
		txStatis.UpdateTime()
		(*txStatis.TransactionCount)++

		this.wm.Log.Debug("transaction[", txid, "] has been sent out.")
		return nil
	}()

	if err != nil {
		this.wm.Log.Errorf("send raw transaction failed, err= %v", err)
		return nil, err
	}

	decimals := int32(0)
	if rawTx.Coin.IsContract {
		decimals = int32(rawTx.Coin.Contract.Decimals)
	} else {
		decimals = int32(this.wm.Decimal())
	}

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       "0",
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	return tx, nil
}

//SendRawTransaction 广播交易单
func (this *EthTransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	if !rawTx.Coin.IsContract {
		return this.SubmitSimpleRawTransaction(wrapper, rawTx)
	}
	return this.SubmitErc20TokenRawTransaction(wrapper, rawTx)
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (this *EthTransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		this.wm.Log.Std.Error("Verify raw tx failed, err=%v", err)
		return err
	}

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		this.wm.Log.Std.Error("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	sig := rawTx.Signatures[rawTx.Account.AccountID][0].Signature
	msg := rawTx.Signatures[rawTx.Account.AccountID][0].Message
	pubkey := rawTx.Signatures[rawTx.Account.AccountID][0].Address.PublicKey
	//curveType := rawTx.Signatures[rawTx.Account.AccountID][0].EccType

	this.wm.Log.Debug("-- pubkey:", pubkey)
	this.wm.Log.Debug("-- message:", msg)
	this.wm.Log.Debug("-- Signature:", sig)
	signature := common.FromHex(sig)
	publickKey := owcrypt.PointDecompress(common.FromHex(pubkey), owcrypt.ECC_CURVE_SECP256K1)
	publickKey = publickKey[1:len(publickKey)]
	ret := owcrypt.Verify(publickKey, nil, 0, common.FromHex(msg), 32, signature[0:len(signature)-1],
		owcrypt.ECC_CURVE_SECP256K1|owcrypt.HASH_OUTSIDE_FLAG)
	if ret != owcrypt.SUCCESS {
		errinfo := fmt.Sprintf("verify error, ret:%v\n", "0x"+strconv.FormatUint(uint64(ret), 16))
		fmt.Println(errinfo)
		return errors.New(errinfo)
	}

	return nil
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (this *EthTransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	if sumRawTx.Coin.IsContract {
		return this.CreateErc20TokenSummaryRawTransaction(wrapper, sumRawTx)
	} else {
		return this.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	}
}

//CreateSimpleSummaryRawTransaction 创建ETH汇总交易
func (this *EthTransactionDecoder) CreateSimpleSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {

	var (
		rawTxArray         = make([]*openwallet.RawTransaction, 0)
		accountID          = sumRawTx.Account.AccountID
		minTransfer, _     = ConvertEthStringToWei(sumRawTx.MinTransfer)
		retainedBalance, _ = ConvertEthStringToWei(sumRawTx.RetainedBalance)
	)

	if minTransfer.Cmp(retainedBalance) < 0 {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit,
		"AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := this.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return nil, err
	}

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI, _ := ConvertEthStringToWei(addrBalance.Balance)

		if addrBalance_BI.Cmp(minTransfer) < 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount_BI := new(big.Int)
		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)

		//this.wm.Log.Debug("sumAmount:", sumAmount)
		//计算手续费
		fee, createErr := this.wm.GetTransactionFeeEstimated(addrBalance.Address, sumRawTx.SummaryAddress, sumAmount_BI, "")
		if createErr != nil {
			this.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, sumRawTx.SummaryAddress, createErr)
			return nil, createErr
		}

		if sumRawTx.FeeRate != "" {
			fee.GasPrice, createErr = ConvertEthStringToWei(sumRawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
			if createErr != nil {
				this.wm.Log.Std.Error("fee rate passed through error, err=%v", createErr)
				return nil, createErr
			}
			fee.CalcFee()
		}

		//减去手续费
		sumAmount_BI.Sub(sumAmount_BI, fee.Fee)
		if sumAmount_BI.Cmp(big.NewInt(0)) <= 0 {
			continue
		}

		sumAmount, _ := ConverWeiStringToEthDecimal(sumAmount_BI.String())
		fees, _ := ConverWeiStringToEthDecimal(fee.Fee.String())

		this.wm.Log.Debugf("balance: %v", addrBalance.Balance)
		this.wm.Log.Debugf("fees: %v", fees)
		this.wm.Log.Debugf("sumAmount: %v", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:    sumRawTx.Coin,
			Account: sumRawTx.Account,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.StringFixed(this.wm.Decimal()),
			},
			Required: 1,
		}

		createErr = this.createRawTransaction(
			wrapper,
			rawTx,
			&AddrBalance{Address: addrBalance.Address, Balance: addrBalance_BI},
			fee,
			"")
		if createErr != nil {
			return nil, createErr
		}

		//创建成功，添加到队列
		rawTxArray = append(rawTxArray, rawTx)

	}

	return rawTxArray, nil
}

//CreateErc20TokenSummaryRawTransaction 创建ERC20Token汇总交易
func (this *EthTransactionDecoder) CreateErc20TokenSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {

	var (
		rawTxArray      = make([]*openwallet.RawTransaction, 0)
		accountID       = sumRawTx.Account.AccountID
		minTransfer     *big.Int
		retainedBalance *big.Int
	)

	//tokenCoin := sumRawTx.Coin.Contract.Token
	tokenDecimals := int(sumRawTx.Coin.Contract.Decimals)
	contractAddress := sumRawTx.Coin.Contract.Address
	//coinDecimals := this.wm.Decimal()

	minTransfer, _ = ConvertFloatStringToBigInt(sumRawTx.MinTransfer, tokenDecimals)
	retainedBalance, _ = ConvertFloatStringToBigInt(sumRawTx.RetainedBalance, tokenDecimals)

	if minTransfer.Cmp(retainedBalance) < 0 {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit,
		"AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	//查询Token余额
	addrBalanceArray, err := this.wm.ContractDecoder.GetTokenBalanceByAddress(sumRawTx.Coin.Contract, searchAddrs...)
	if err != nil {
		return nil, err
	}

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI, _ := ConvertFloatStringToBigInt(addrBalance.Balance.Balance, tokenDecimals)

		if addrBalance_BI.Cmp(minTransfer) < 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount_BI := new(big.Int)
		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)

		callData, err := makeERC20TokenTransData(contractAddress, sumRawTx.SummaryAddress, sumAmount_BI)

		//this.wm.Log.Debug("sumAmount:", sumAmount)
		//计算手续费
		fee, createErr := this.wm.GetTransactionFeeEstimated(addrBalance.Balance.Address, contractAddress, nil, callData)
		if createErr != nil {
			this.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Balance.Address, sumRawTx.SummaryAddress, createErr)
			return nil, createErr
		}

		if sumRawTx.FeeRate != "" {
			fee.GasPrice, createErr = ConvertEthStringToWei(sumRawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
			if createErr != nil {
				this.wm.Log.Std.Error("fee rate passed through error, err=%v", createErr)
				return nil, createErr
			}
			fee.CalcFee()
		}

		coinBalance, err := this.wm.WalletClient.GetAddrBalance2(appendOxToAddress(addrBalance.Balance.Address), "latest")
		if err != nil {
			continue
		}

		//判断主币余额是否够手续费
		if coinBalance.Cmp(fee.Fee) <= 0 {
			continue
		}

		sumAmount, _ := ConvertAmountToFloatDecimal(sumAmount_BI.String(), tokenDecimals)
		fees, _ := ConverWeiStringToEthDecimal(fee.Fee.String())

		this.wm.Log.Debugf("balance: %v", addrBalance.Balance.Balance)
		this.wm.Log.Debugf("%s fees: %v", sumRawTx.Coin.Symbol, fees)
		this.wm.Log.Debugf("sumAmount: %v", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:    sumRawTx.Coin,
			Account: sumRawTx.Account,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.StringFixed(int32(tokenDecimals)),
			},
			Required: 1,
		}

		createErr = this.createRawTransaction(
			wrapper,
			rawTx,
			&AddrBalance{Address: addrBalance.Balance.Address, Balance: coinBalance, TokenBalance: addrBalance_BI},
			fee,
			callData)
		if createErr != nil {
			return nil, createErr
		}

		//创建成功，添加到队列
		rawTxArray = append(rawTxArray, rawTx)

	}

	return rawTxArray, nil
}

//createRawTransaction
func (this *EthTransactionDecoder) createRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction, addrBalance *AddrBalance, fee *txFeeInfo, callData string) error {

	var (
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		keySignList      = make([]*openwallet.KeySignature, 0)
		amountStr        string
		destination      string
		tx               *types.Transaction
	)

	isContract := rawTx.Coin.IsContract
	//contractAddress := rawTx.Coin.Contract.Address
	//tokenCoin := rawTx.Coin.Contract.Token
	tokenDecimals := int(rawTx.Coin.Contract.Decimals)
	//coinDecimals := this.wm.Decimal()

	for k, v := range rawTx.To {
		destination = k
		amountStr = v
		break
	}

	//计算账户的实际转账amount
	accountTotalSentAddresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", destination)
	if findErr != nil || len(accountTotalSentAddresses) == 0 {
		amountDec, _ := decimal.NewFromString(amountStr)
		accountTotalSent = accountTotalSent.Add(amountDec)
	}

	txFrom = []string{fmt.Sprintf("%s:%s", appendOxToAddress(addrBalance.Address), amountStr)}
	txTo = []string{fmt.Sprintf("%s:%s", appendOxToAddress(destination), amountStr)}

	addr, err := wrapper.GetAddress(addrBalance.Address)
	if err != nil {
		return err
	}

	_, nonce, err := this.GetTransactionCount2(addrBalance.Address)
	if err != nil {
		this.wm.Log.Std.Error("GetTransactionCount2 failed, err=%v", err)
		return err
	}
	//this.wm.Log.Debug("chainID:", this.wm.GetConfig().ChainID)
	signer := types.NewEIP155Signer(big.NewInt(int64(this.wm.GetConfig().ChainID)))

	if isContract {
		//构建合约交易
		amount, _ := ConvertFloatStringToBigInt(amountStr, tokenDecimals)
		if addrBalance.TokenBalance.Cmp(amount) < 0 {
			return fmt.Errorf("the token balance: %s is not enough", amountStr)
		}

		if addrBalance.Balance.Cmp(fee.Fee) < 0 {
			coinBalance, _ := ConverWeiStringToEthDecimal(addrBalance.Balance.String())
			return fmt.Errorf("the [%s] balance: %s is not enough to call smart contract", rawTx.Coin.Symbol, coinBalance)
		}

		tx = types.NewTransaction(nonce, ethcommon.HexToAddress(rawTx.Coin.Contract.Address),
			big.NewInt(0), fee.GasLimit.Uint64(), fee.GasPrice, common.FromHex(callData))
	} else {
		//构建ETH交易
		amount, _ := ConvertEthStringToWei(amountStr)

		totalAmount := new(big.Int)
		totalAmount.Add(amount, fee.Fee)
		if addrBalance.Balance.Cmp(totalAmount) < 0 {
			return fmt.Errorf("the [%s] balance: %s is not enough", rawTx.Coin.Symbol, amountStr)
		}

		tx = types.NewTransaction(nonce, ethcommon.HexToAddress(destination),
			amount, fee.GasLimit.Uint64(), fee.GasPrice, []byte(""))
	}

	rawHex, err := rlp.EncodeToBytes(tx)
	if err != nil {
		this.wm.Log.Error("Transaction RLP encode failed, err:", err)
		return err
	}

	rawTx.RawHex = hex.EncodeToString(rawHex)

	txstr, _ := json.MarshalIndent(tx, "", " ")
	this.wm.Log.Debug("**txStr:", string(txstr))
	msg := signer.Hash(tx)

	gasLimitStr, err := ConverWeiStringToEthDecimal(fee.GasLimit.String())
	if err != nil {
		this.wm.Log.Error("ConverWeiStringToEthDecimal failed, err=", err)
		return err
	}

	extpara := EthTxExtPara{
		GasLimit: gasLimitStr.String(), //"0x" + fee.GasLimit.Text(16),
	}
	extparastr, _ := json.Marshal(extpara)

	gasprice, err := ConverWeiStringToEthDecimal(fee.GasPrice.String())
	if err != nil {
		this.wm.Log.Error("convert wei string to gas price failed, err=", err)
		return err
	}

	totalFeeDecimal, err := ConverWeiStringToEthDecimal(fee.Fee.String())
	if err != nil {
		this.wm.Log.Errorf("convert total fee from wei string to eth decimal failed, err=%v", err)
		return err
	}

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	signature := openwallet.KeySignature{
		EccType: this.wm.Config.CurveType,
		Nonce:   "0x" + strconv.FormatUint(nonce, 16),
		Address: addr,
		Message: hex.EncodeToString(msg[:]),
	}
	keySignList = append(keySignList, &signature)

	feesDec, _ := decimal.NewFromString(rawTx.Fees)
	accountTotalSent = accountTotalSent.Add(feesDec)
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.Signatures[rawTx.Account.AccountID] = keySignList
	rawTx.FeeRate = gasprice.String()
	rawTx.Fees = totalFeeDecimal.String()
	rawTx.ExtParam = string(extparastr)
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.StringFixed(this.wm.Decimal())
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}
