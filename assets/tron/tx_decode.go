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

package tron

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/openwallet/assets/tron/grpc-gateway/core"
	"github.com/blocktree/openwallet/common"
	"github.com/golang/protobuf/proto"
	"math/big"
	"sort"
	"strings"

	"time"

	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
	// "github.com/blocktree/openwallet/assets/qtum/btcLikeTxDriver"
	// "github.com/blocktree/openwallet/log"
	// "github.com/shopspring/decimal"
)

//TransactionDecoder for Interface TransactionDecode
type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

func CheckRawTransaction(rawTx *openwallet.RawTransaction) error {
	//账户模型原始账单只有一个To
	if len(rawTx.To) != 1 {
		return fmt.Errorf("noly one to address can be set!")
	}
	return nil
}

func InsertSignatureIntoRawTransaction(txHex string, signature string) (string, error) {

	tx := &core.Transaction{}
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return "", err
	}
	if err := proto.Unmarshal(txBytes, tx); err != nil {
		return "", err
	}

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		//log.Errorf("invalid transaction signature hex data;unexpected err:%v", err)
		return "", fmt.Errorf("invalid signature hex data")
	}

	tx.Signature = append(tx.Signature, signatureBytes)
	x, err := proto.Marshal(tx)
	if err != nil {
		//wm.Log.Info("marshal tx failed;unexpected error:%v", err)
		return "", err
	}

	mergeTxHex := hex.EncodeToString(x)
	return mergeTxHex, nil

}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateSimpleTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountID       = rawTx.Account.AccountID
		findAddrBalance *AddrBalance
		rawHex          string
		feeInfo         *txFeeInfo
	)

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", accountID) //wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
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

	addrBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
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
		a_amount, _ := decimal.NewFromString(addrBalanceArray[i].Balance)
		b_amount, _ := decimal.NewFromString(addrBalanceArray[j].Balance)
		if a_amount.LessThan(b_amount) {
			return true
		} else {
			return false
		}
	})

	amountDec, _ := decimal.NewFromString(amountStr)

	for _, addrBalance := range addrBalanceArray {

		addrBalance_dec, _ := decimal.NewFromString(addrBalance.Balance)

		//创建空交易单
		rawHex, err = decoder.wm.CreateTokenTransaction(to, addrBalance.Address, amountStr, openwallet.SmartContract{})
		if err != nil {
			return err
		}

		rawTx.RawHex = rawHex

		//计算手续费
		feeInfo, err = decoder.wm.GetTransactionFeeEstimated(addrBalance.Address, rawHex)
		if err != nil {
			decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, to, err)
			continue
		}

		//总消耗数量 = 转账数量 + 手续费
		if addrBalance_dec.LessThan(amountDec.Add(feeInfo.Fee)) {
			continue
		}

		//只要找到一个合适使用的地址余额就停止遍历
		findAddrBalance = &AddrBalance{Address: addrBalance.Address, TronBalance: common.StringNumToBigIntWithExp(amountStr, Decimals)}
		break
	}

	if findAddrBalance == nil {
		return fmt.Errorf("the balance: %s is not enough", amountStr)
	}

	//最后创建交易单
	err = decoder.createRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		feeInfo,
		"")
	if err != nil {
		return err
	}

	return nil
}

//CreateTokenTransaction 创建代币交易单
func (decoder *TransactionDecoder) CreateTokenTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountID       = rawTx.Account.AccountID
		findAddrBalance *AddrBalance
		rawHex          string
		feeInfo         *txFeeInfo
	)

	tokenDecimals := rawTx.Coin.Contract.Decimals

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", accountID) //wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
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

	addrBalanceArray, err := decoder.wm.ContractDecoder.GetTokenBalanceByAddress(rawTx.Coin.Contract, searchAddrs...)
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

	amountDec, _ := decimal.NewFromString(amountStr)

	for _, addrBalance := range addrBalanceArray {

		trxBalance := big.NewInt(0)

		addrBalance_dec, _ := decimal.NewFromString(addrBalance.Balance.Balance)

		//创建空交易单
		rawHex, err = decoder.wm.CreateTokenTransaction(to, addrBalance.Balance.Address, amountStr, rawTx.Coin.Contract)
		if err != nil {
			return err
		}

		rawTx.RawHex = rawHex

		//计算手续费
		feeInfo, err = decoder.wm.GetTransactionFeeEstimated(addrBalance.Balance.Address, rawHex)
		if err != nil {
			decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Balance.Address, to, err)
			continue
		}

		//查询主币余额是否足够
		addrTRXBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(addrBalance.Balance.Address)
		if err != nil {
			return err
		}
		if len(addrTRXBalanceArray) > 0 {
			trxBalance = common.StringNumToBigIntWithExp(addrTRXBalanceArray[0].Balance, decoder.wm.Decimal())
		}

		//总消耗数量 = 转账数量 + 手续费
		if addrBalance_dec.LessThan(amountDec.Add(feeInfo.Fee)) {
			continue
		}

		//只要找到一个合适使用的地址余额就停止遍历
		findAddrBalance = &AddrBalance{
			Address:      addrBalance.Balance.Address,
			TokenBalance: common.StringNumToBigIntWithExp(amountStr, int32(tokenDecimals)),
			TronBalance:  trxBalance,
		}
		break
	}

	if findAddrBalance == nil {
		return fmt.Errorf("the balance: %s is not enough", amountStr)
	}

	//最后创建交易单
	err = decoder.createRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		feeInfo,
		"")
	if err != nil {
		return err
	}

	return nil
}

func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return decoder.CreateSimpleTransaction(wrapper, rawTx)
	} else {
		return decoder.CreateTokenTransaction(wrapper, rawTx)
	}
	//contract To Do
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		decoder.wm.Log.Info("wrapper HDkey failed;unexpected error:%v", err)
		return err
	}
	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	//fmt.Println("keySignatures:=", keySignatures)
	if keySignatures != nil {
		for _, keySignature := range keySignatures {
			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, decoder.wm.CurveType())
			if err != nil {
				decoder.wm.Log.Info("derived key with path failed;unexpected error:%v", err)
				return err
			}
			priKeyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				decoder.wm.Log.Info("get privatekey bytes failed;unexpected error:%v", err)
				return err
			}
			//txHashBytes, err := getTxHash1(rawTx.RawHex)
			//if err != nil {
			//	decoder.wm.Log.Info("get Tx hash failed;unexpected error:%v", err)
			//	return err
			//}
			//txHash := hex.EncodeToString(txHashBytes)
			txHash := keySignature.Message
			priKey := hex.EncodeToString(priKeyBytes)
			signature, err := decoder.wm.SignTransactionRef(txHash, priKey)
			if err != nil {
				decoder.wm.Log.Info("sign Tx failed;unexpected error:%v", err)
				return err
			}
			keySignature.Signature = signature
		}
	}
	decoder.wm.Log.Info("Tx hash sign success")
	//rawTx.Signatures[rawTx.Account.AccountID] = keySignatures
	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//检测交易单基本字段
	err := CheckRawTransaction(rawTx)
	if err != nil {
		decoder.wm.Log.Info("verify Tx base field failed;unexpected error:%v", err)
		return err
	}

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	sig, exist := rawTx.Signatures[rawTx.Account.AccountID]
	if !exist {
		return fmt.Errorf("wallet signature not found")
	}

	if len(sig) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	mergeTxHex, err := InsertSignatureIntoRawTransaction(rawTx.RawHex, sig[0].Signature)
	if err != nil {
		decoder.wm.Log.Info("merge empty transaction and signature failed;unexpected error:%v", err)
		return err
	}
	verifyRet := decoder.wm.ValidSignedTokenTransaction(mergeTxHex, rawTx.Coin.Contract)
	if verifyRet != nil {
		decoder.wm.Log.Info("Tx signature verify failed;unexpected error:%v", verifyRet)
		return fmt.Errorf("Tx signature verify failed")
	} else {
		rawTx.IsCompleted = true
		//rawTx.RawHex = mergeTxHex
	}
	return nil
}

//SubmitRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}
	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}
	//********合并交易单********
	sig, exist := rawTx.Signatures[rawTx.Account.AccountID]
	if !exist {
		return nil, fmt.Errorf("wallet signature not found")
	}

	if len(sig) == 0 {
		return nil, fmt.Errorf("transaction signature is empty")
	}

	mergeTxHex, err := InsertSignatureIntoRawTransaction(rawTx.RawHex, sig[0].Signature)
	if err != nil {
		decoder.wm.Log.Info("merge empty transaction and signature failed;unexpected error:%v", err)
		return nil, err
	}

	contractType := core.Transaction_Contract_TransferContract
	if rawTx.Coin.Contract.Address == "" {
		contractType = core.Transaction_Contract_TransferContract
	} else if strings.EqualFold(rawTx.Coin.Contract.Protocol, TRC10) {
		contractType = core.Transaction_Contract_TransferAssetContract
	} else if strings.EqualFold(rawTx.Coin.Contract.Protocol, TRC20) {
		contractType = core.Transaction_Contract_TriggerSmartContract
	}

	rawTx.RawHex = mergeTxHex
	//********广播交易单********
	txid, err := decoder.wm.BroadcastTransaction(rawTx.RawHex, contractType)
	if err != nil {
		decoder.wm.Log.Infof("submit transaction failed;unexpected error: %v", err)
		return nil, err
	}
	rawTx.TxID = txid
	rawTx.IsSubmit = true
	decimals := decoder.wm.Decimal()
	tx := openwallet.Transaction{
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
	tx.WxID = openwallet.GenTransactionWxID(&tx)
	return &tx, nil
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	if sumRawTx.Coin.IsContract {
		return decoder.CreateTokenSummaryRawTransaction(wrapper, sumRawTx)
	} else {
		return decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	}
}

//CreateSimpleSummaryRawTransaction 创建主币汇总交易
func (decoder *TransactionDecoder) CreateSimpleSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {

	var (
		rawTxArray      = make([]*openwallet.RawTransaction, 0)
		accountID       = sumRawTx.Account.AccountID
		minTransfer     = common.StringNumToBigIntWithExp(sumRawTx.MinTransfer, Decimals)
		retainedBalance = common.StringNumToBigIntWithExp(sumRawTx.RetainedBalance, Decimals)
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

	addrBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return nil, err
	}

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI := common.StringNumToBigIntWithExp(addrBalance.Balance, Decimals)

		if addrBalance_BI.Cmp(minTransfer) < 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount_BI := new(big.Int)
		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)
		sumAmount := common.BigIntToDecimals(sumAmount_BI, Decimals)

		//创建空交易单
		rawHex, createErr := decoder.wm.CreateTokenTransaction(sumRawTx.SummaryAddress,
			addrBalance.Address, sumAmount.String(), sumRawTx.Coin.Contract)
		if createErr != nil {
			return nil, createErr
		}

		//this.wm.Log.Debug("sumAmount:", sumAmount)
		//计算手续费
		fee, createErr := decoder.wm.GetTransactionFeeEstimated(addrBalance.Address, rawHex)
		if createErr != nil {
			decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, sumRawTx.SummaryAddress, createErr)
			return nil, createErr
		}



		//减去手续费
		//sumAmount = sumAmount.Sub(fee.Fee)
		//if sumAmount.LessThanOrEqual(decimal.Zero) {
		//	continue
		//}

		decoder.wm.Log.Debugf("balance: %v", addrBalance.Balance)
		decoder.wm.Log.Debugf("fees: %v", fee.Fee)
		decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:    sumRawTx.Coin,
			Account: sumRawTx.Account,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.StringFixed(decoder.wm.Decimal()),
			},
			Required: 1,
			RawHex:   rawHex,
		}

		createErr = decoder.createRawTransaction(
			wrapper,
			rawTx,
			&AddrBalance{Address: addrBalance.Address, TronBalance: addrBalance_BI},
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

//CreateTokenSummaryRawTransaction 创建代币汇总交易
func (decoder *TransactionDecoder) CreateTokenSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {

	var (
		rawTxArray      = make([]*openwallet.RawTransaction, 0)
		accountID       = sumRawTx.Account.AccountID
		minTransfer     = common.StringNumToBigIntWithExp(sumRawTx.MinTransfer, Decimals)
		retainedBalance = common.StringNumToBigIntWithExp(sumRawTx.RetainedBalance, Decimals)
	)

	tokenDecimals := int32(sumRawTx.Coin.Contract.Decimals)

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

	addrBalanceArray, err := decoder.wm.ContractDecoder.GetTokenBalanceByAddress(sumRawTx.Coin.Contract, searchAddrs...)
	if err != nil {
		return nil, err
	}

	for _, addrBalance := range addrBalanceArray {

		trxBalance := big.NewInt(0)

		//检查余额是否超过最低转账
		addrBalance_BI := common.StringNumToBigIntWithExp(addrBalance.Balance.Balance, tokenDecimals)

		if addrBalance_BI.Cmp(minTransfer) < 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount_BI := new(big.Int)
		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)
		sumAmount := common.BigIntToDecimals(sumAmount_BI, tokenDecimals)

		//创建空交易单
		rawHex, createErr := decoder.wm.CreateTokenTransaction(sumRawTx.SummaryAddress,
			addrBalance.Balance.Address, sumAmount.String(), sumRawTx.Coin.Contract)
		if createErr != nil {
			return nil, createErr
		}

		//this.wm.Log.Debug("sumAmount:", sumAmount)
		//计算手续费
		fee, createErr := decoder.wm.GetTransactionFeeEstimated(addrBalance.Balance.Address, rawHex)
		if createErr != nil {
			decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Balance.Address, sumRawTx.SummaryAddress, createErr)
			return nil, createErr
		}

		//减去手续费
		//sumAmount = sumAmount.Sub(fee.Fee)
		//if sumAmount.LessThanOrEqual(decimal.Zero) {
		//	continue
		//}

		//查询主币余额是否足够
		addrTRXBalanceArray, createErr := decoder.wm.Blockscanner.GetBalanceByAddress(addrBalance.Balance.Address)
		if createErr != nil {
			return nil, createErr
		}
		if len(addrTRXBalanceArray) > 0 {
			trxBalance = common.StringNumToBigIntWithExp(addrTRXBalanceArray[0].Balance, decoder.wm.Decimal())
		}

		decoder.wm.Log.Debugf("balance: %v", addrBalance.Balance.Balance)
		decoder.wm.Log.Debugf("fees: %v", fee.Fee)
		decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:    sumRawTx.Coin,
			Account: sumRawTx.Account,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.StringFixed(tokenDecimals),
			},
			Required: 1,
			RawHex:   rawHex,
		}

		createErr = decoder.createRawTransaction(
			wrapper,
			rawTx,
			&AddrBalance{
				Address:      addrBalance.Balance.Address,
				TokenBalance: addrBalance_BI,
				TronBalance:  trxBalance},
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

//createRawTransaction
func (decoder *TransactionDecoder) createRawTransaction(
	wrapper openwallet.WalletDAI,
	rawTx *openwallet.RawTransaction,
	addrBalance *AddrBalance,
	feeInfo *txFeeInfo,
	callData string) error {

	var (
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		keySignList      = make([]*openwallet.KeySignature, 0)
		amountStr        string
		destination      string
		rawHex           string
	)

	decimals := int32(0)
	if rawTx.Coin.IsContract {
		decimals = int32(rawTx.Coin.Contract.Decimals)
	} else {
		decimals = decoder.wm.Decimal()
	}
	//isContract := rawTx.Coin.IsContract
	//contractAddress := rawTx.Coin.Contract.Address
	//tokenCoin := rawTx.Coin.Contract.Token
	//tokenDecimals := int(rawTx.Coin.Contract.Decimals)
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

	txFrom = []string{fmt.Sprintf("%s:%s", addrBalance.Address, amountStr)}
	txTo = []string{fmt.Sprintf("%s:%s", destination, amountStr)}

	addr, err := wrapper.GetAddress(addrBalance.Address)
	if err != nil {
		return err
	}

	rawHex = rawTx.RawHex

	txHashBytes, err := getTxHash1(rawHex)
	if err != nil {
		decoder.wm.Log.Info("get Tx hash failed;unexpected error:%v", err)
		return err
	}
	txHash := hex.EncodeToString(txHashBytes)

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Address: addr,
		Message: txHash,
	}
	keySignList = append(keySignList, &signature)

	feesDec, _ := decimal.NewFromString(rawTx.Fees)
	accountTotalSent = accountTotalSent.Add(feesDec)
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	//rawTx.RawHex = rawHex
	rawTx.Signatures[rawTx.Account.AccountID] = keySignList
	rawTx.FeeRate = feeInfo.GasPrice.String()
	rawTx.Fees = feeInfo.Fee.String()
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.StringFixed(decimals)
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}
