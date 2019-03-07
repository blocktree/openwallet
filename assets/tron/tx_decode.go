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
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strconv" // "sort"
	// "github.com/blocktree/OpenWallet/log"
	"strings"
	"time"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
	// "github.com/blocktree/OpenWallet/assets/qtum/btcLikeTxDriver"
	// "github.com/blocktree/OpenWallet/log"
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
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		log.Errorf("nvalid transaction hex data;unexpected err:%v", err)
		return "", fmt.Errorf("invalid transaction hex data")
	}
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		log.Errorf("invalid transaction signature hex data;unexpected err:%v", err)
		return "", fmt.Errorf("invalid signature hex data")
	}
	mergeTxBytes := append(txBytes, signatureBytes...)
	mergeTxHex := hex.EncodeToString(mergeTxBytes)
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
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
	)
	if rawTx.Coin.Symbol != Symbol {
		return fmt.Errorf("CreateRawTransaction: Symbol is not <TRX>")
	}

	if len(rawTx.To) == 0 {
		return fmt.Errorf("CreateRawTransaction: Receiver addresses is empty")
	}
	if rawTx.Account.AccountID == "" {
		return fmt.Errorf("CreateRawTransaction: AccountID is empty")
	}
	addressList, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)
	if err != nil {
		decoder.wm.Log.Info("get address list failed;unexpected error:%v", err)
		return err
	}
	if len(addressList) == 0 {
		return fmt.Errorf("[%s] account: %s has not addresses", decoder.wm.Symbol(), rawTx.Account.AccountID)
	}
	addressesBalanceList := make([]AddrBalance, 0, len(addressList))
	for i, addr := range addressList {
		balance, err := decoder.wm.Getbalance(addr.Address)
		if err != nil {
			decoder.wm.Log.Info("get balance failed;unexpected error:%v", err)
			return err
		}
		balance.Index = i
		addressesBalanceList = append(addressesBalanceList, *balance)
	}
	sort.Slice(addressesBalanceList, func(i int, j int) bool {
		return addressesBalanceList[i].TronBalance.Cmp(addressesBalanceList[j].TronBalance) >= 0
	})
	var amountStr, toAddress string
	for k, v := range rawTx.To {
		toAddress = k
		amountStr = v
		break
	}
	//计算账户的实际转账amount
	accountTotalSentAddresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", toAddress)
	if findErr != nil || len(accountTotalSentAddresses) == 0 {
		amountDec, _ := decimal.NewFromString(amountStr)
		accountTotalSent = accountTotalSent.Add(amountDec)
	}
	txTo = []string{fmt.Sprintf("%s:%s", toAddress, amountStr)}
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		decoder.wm.Log.Info("conver amount from string  to float failed;unexpected error:%v", err)
		return err
	}
	signatureMap := make(map[string][]*openwallet.KeySignature)
	keySignList := make([]*openwallet.KeySignature, 0, 1)
	amountInt64 := int64(amountFloat * 1000000)
	amount := big.NewInt(amountInt64)
	count := big.NewInt(0)
	countList := []uint64{}
	for i, _ := range addressesBalanceList {
		if addressesBalanceList[i].TronBalance.Cmp(amount) < 0 {
			count.Add(count, addressesBalanceList[i].TronBalance)
			if count.Cmp(amount) >= 0 {
				countList = append(countList, addressesBalanceList[i].TronBalance.Sub(addressesBalanceList[i].TronBalance, count.Sub(count, amount)).Uint64())
				return fmt.Errorf("The Tron of account is enough,"+
					"but cannot be sent in just one transaction!\n"+
					"the amount can be sent in "+string(len(countList))+"times with amounts:\n"+strings.Replace(strings.Trim(fmt.Sprint(countList), "[]"), " ", ",", -1), err)
			} else {
				countList = append(countList, addressesBalanceList[i].TronBalance.Uint64())
			}
			continue
		} else {
			ownerAddress := &openwallet.Address{
				AccountID:   addressList[addressesBalanceList[i].Index].AccountID,
				Address:     addressList[addressesBalanceList[i].Index].Address,
				PublicKey:   addressList[addressesBalanceList[i].Index].PublicKey,
				Alias:       addressList[addressesBalanceList[i].Index].Alias,
				Tag:         addressList[addressesBalanceList[i].Index].Tag,
				Index:       addressList[addressesBalanceList[i].Index].Index,
				HDPath:      addressList[addressesBalanceList[i].Index].HDPath,
				WatchOnly:   addressList[addressesBalanceList[i].Index].WatchOnly,
				Symbol:      addressList[addressesBalanceList[i].Index].Symbol,
				Balance:     addressList[addressesBalanceList[i].Index].Balance,
				IsMemo:      addressList[addressesBalanceList[i].Index].IsMemo,
				Memo:        addressList[addressesBalanceList[i].Index].Memo,
				CreatedTime: addressList[addressesBalanceList[i].Index].CreatedTime,
			}
			keySignList = append(keySignList, &openwallet.KeySignature{
				Address: ownerAddress,
			})
			txFrom = []string{fmt.Sprintf("%s:%s", ownerAddress.Address, amountStr)}
			break
		}

	}
	//fmt.Println("from address:=", keySignList[0].Address.Address)
	if len(keySignList) != 1 {
		return fmt.Errorf("NO enough Tron to send")
	}

	//创建空交易单
	rawHex, err := decoder.wm.CreateTransactionRef(toAddress, keySignList[0].Address.Address, amountFloat)
	if err != nil {
		return err
	}

	txHashBytes, err := getTxHash1(rawHex)
	if err != nil {
		decoder.wm.Log.Info("get Tx hash failed;unexpected error:%v", err)
		return err
	}
	txHash := hex.EncodeToString(txHashBytes)
	keySignList[0].Nonce = ""
	keySignList[0].Message = txHash
	signatureMap[rawTx.Account.AccountID] = keySignList
	//rawTx.Signatures = make(map[string][]*openwallet.KeySignature, 0)

	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.Fees = "0"
	rawTx.FeeRate = "0"
	rawTx.RawHex = rawHex
	rawTx.Signatures = signatureMap
	rawTx.IsBuilt = true
	rawTx.TxTo = txTo
	rawTx.TxFrom = txFrom
	rawTx.TxAmount = accountTotalSent.StringFixed(decoder.wm.Decimal())
	return nil
}

func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return decoder.CreateSimpleTransaction(wrapper, rawTx)
	} else {
		return nil
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
			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, 0xECC00000)
			if err != nil {
				decoder.wm.Log.Info("derived key with path failed;unexpected error:%v", err)
				return err
			}
			priKeyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				decoder.wm.Log.Info("get privatekey bytes failed;unexpected error:%v", err)
				return err
			}
			txHashBytes, err := getTxHash1(rawTx.RawHex)
			if err != nil {
				decoder.wm.Log.Info("get Tx hash failed;unexpected error:%v", err)
				return err
			}
			txHash := hex.EncodeToString(txHashBytes)
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
	verifyRet := decoder.wm.ValidSignedTransactionRef(mergeTxHex)
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
	mergeTxHex, err := InsertSignatureIntoRawTransaction(rawTx.RawHex, sig[0].Signature)
	if err != nil {
		decoder.wm.Log.Info("merge empty transaction and signature failed;unexpected error:%v", err)
		return nil, err
	}
	rawTx.RawHex = mergeTxHex
	//********广播交易单********
	txid, err := decoder.wm.BroadcastTransaction(rawTx.RawHex)
	if err != nil {
		decoder.wm.Log.Info("submit transaction failed;unexpected erroe:%v", err)
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
//func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
//	if sumRawTx.Coin.IsContract {
//		return nil, fmt.Errorf("can not support token summary transaction")
//	} else {
//		return decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
//	}
//}


//CreateSimpleSummaryRawTransaction 创建ETH汇总交易
//func (decoder *TransactionDecoder) CreateSimpleSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
//
//	var (
//		rawTxArray         = make([]*openwallet.RawTransaction, 0)
//		accountID          = sumRawTx.Account.AccountID
//		minTransfer, _     = ConvertNasStringToWei(sumRawTx.MinTransfer)
//		retainedBalance, _ = ConvertNasStringToWei(sumRawTx.RetainedBalance)
//	)
//
//	if minTransfer.Cmp(retainedBalance) < 0 {
//		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
//	}
//
//	//获取wallet
//	addresses, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit,
//		"AccountID", sumRawTx.Account.AccountID)
//	if err != nil {
//		return nil, err
//	}
//
//	if len(addresses) == 0 {
//		return nil, fmt.Errorf("[%s] have not addresses", accountID)
//	}
//
//	searchAddrs := make([]string, 0)
//	for _, address := range addresses {
//		searchAddrs = append(searchAddrs, address.Address)
//	}
//
//	addrBalanceArray, err := decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, addrBalance := range addrBalanceArray {
//
//		//检查余额是否超过最低转账
//		addrBalance_BI, _ := ConvertNasStringToWei(addrBalance.Balance)
//
//		if addrBalance_BI.Cmp(minTransfer) < 0 {
//			continue
//		}
//		//计算汇总数量 = 余额 - 保留余额
//		sumAmount_BI := new(big.Int)
//		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)
//
//		//this.wm.Log.Debug("sumAmount:", sumAmount)
//		//计算手续费
//		fee, createErr := decoder.wm.Getestimatefee(addrBalance.Address, sumRawTx.SummaryAddress, sumAmount_BI)
//		if createErr != nil {
//			decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, sumRawTx.SummaryAddress, createErr)
//			return nil, createErr
//		}
//
//		if sumRawTx.FeeRate != "" {
//			fee.GasPrice, createErr = ConvertNasStringToWei(sumRawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
//			if createErr != nil {
//				decoder.wm.Log.Std.Error("fee rate passed through error, err=%v", createErr)
//				return nil, createErr
//			}
//			fee.CalcFee()
//		}
//
//		//减去手续费
//		sumAmount_BI.Sub(sumAmount_BI, fee.Fee)
//		if sumAmount_BI.Cmp(big.NewInt(0)) <= 0 {
//			continue
//		}
//
//		sumAmount, _ := ConverWeiStringToNasDecimal(sumAmount_BI.String())
//		fees, _ := ConverWeiStringToNasDecimal(fee.Fee.String())
//
//		decoder.wm.Log.Debugf("balance: %v", addrBalance.Balance)
//		decoder.wm.Log.Debugf("fees: %v", fees)
//		decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)
//
//		//创建一笔交易单
//		rawTx := &openwallet.RawTransaction{
//			Coin:    sumRawTx.Coin,
//			Account: sumRawTx.Account,
//			To: map[string]string{
//				sumRawTx.SummaryAddress: sumAmount.StringFixed(decoder.wm.Decimal()),
//			},
//			Required: 1,
//		}
//
//		createErr = decoder.createRawTransaction(
//			wrapper,
//			rawTx,
//			&AddrBalance{Address: addrBalance.Address, Balance: addrBalance_BI},
//			fee,
//			"")
//		if createErr != nil {
//			return nil, createErr
//		}
//
//		//创建成功，添加到队列
//		rawTxArray = append(rawTxArray, rawTx)
//
//	}
//
//	return rawTxArray, nil
//}