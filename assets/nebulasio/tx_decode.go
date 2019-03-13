/*
 * Copyright 2018 The openwallet Authors
 * decoder file is part of the openwallet library.
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
package nebulasio

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"time"

	ow "github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/logger"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/blocktree/go-owcrypt"
	"github.com/bytom/common"
	"github.com/shopspring/decimal"
)

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

func CheckRawTransaction(rawTx *openwallet.RawTransaction) error {
	//账户模型原始账单只有一个To
	if len(rawTx.To) != 1 {
		openwLogger.Log.Errorf("noly one To address can be set.")
		return errors.New("noly one to address can be set.")
	}
	return nil
}

type AddrBalance struct {
	Address      string
	Balance      *big.Int
	TokenBalance *big.Int
	Index        int
}

type txFeeInfo struct {
	GasUse   *big.Int
	GasPrice *big.Int
	Fee      *big.Int
}


//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return decoder.CreateSimpleRawTransaction(wrapper, rawTx)
	}
	//wjq return decoder.CreateNRC20TokenRawTransaction(wrapper, rawTx)
	return nil
}



//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	if sumRawTx.Coin.IsContract {
		return nil, fmt.Errorf("can not support token summary transaction")
	} else {
		return decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	}
}


//SendRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	if !rawTx.Coin.IsContract {
		return decoder.SubmitSimpleRawTransaction(wrapper, rawTx)
	}

	return nil, fmt.Errorf("Contract is not supported. ")
	//wjq return decoder.SubmitErc20TokenRawTransaction(wrapper, rawTx)
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {

	rate := decoder.wm.EstimateFeeRate()
	rate_decimal := decimal.RequireFromString(rate).Div(coinDecimal)

	return rate_decimal.StringFixed(decoder.wm.Decimal()), "Gas", nil
}


func (decoder *TransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountID       = rawTx.Account.AccountID
		findAddrBalance *AddrBalance
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

	amount, _ := ConvertNasStringToWei(amountStr)

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
		addrBalance_BI, _ := ConvertNasStringToWei(addrBalance.Balance)

		//计算手续费
		feeInfo, err = decoder.wm.Getestimatefee(addrBalance.Address, to, amount)
		if err != nil {
			decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, to, err)
			continue
		}

		if rawTx.FeeRate != "" {
			feeInfo.GasPrice, _ = ConvertNasStringToWei(rawTx.FeeRate)
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


//CreateSimpleSummaryRawTransaction 创建ETH汇总交易
func (decoder *TransactionDecoder) CreateSimpleSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {

	var (
		rawTxArray         = make([]*openwallet.RawTransaction, 0)
		accountID          = sumRawTx.Account.AccountID
		minTransfer, _     = ConvertNasStringToWei(sumRawTx.MinTransfer)
		retainedBalance, _ = ConvertNasStringToWei(sumRawTx.RetainedBalance)
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
		addrBalance_BI, _ := ConvertNasStringToWei(addrBalance.Balance)

		if addrBalance_BI.Cmp(minTransfer) < 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount_BI := new(big.Int)
		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)

		//this.wm.Log.Debug("sumAmount:", sumAmount)
		//计算手续费
		fee, createErr := decoder.wm.Getestimatefee(addrBalance.Address, sumRawTx.SummaryAddress, sumAmount_BI)
		if createErr != nil {
			decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, sumRawTx.SummaryAddress, createErr)
			return nil, createErr
		}

		if sumRawTx.FeeRate != "" {
			fee.GasPrice, createErr = ConvertNasStringToWei(sumRawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
			if createErr != nil {
				decoder.wm.Log.Std.Error("fee rate passed through error, err=%v", createErr)
				return nil, createErr
			}
			fee.CalcFee()
		}

		//减去手续费
		sumAmount_BI.Sub(sumAmount_BI, fee.Fee)
		if sumAmount_BI.Cmp(big.NewInt(0)) <= 0 {
			continue
		}

		sumAmount, _ := ConverWeiStringToNasDecimal(sumAmount_BI.String())
		fees, _ := ConverWeiStringToNasDecimal(fee.Fee.String())

		decoder.wm.Log.Debugf("balance: %v", addrBalance.Balance)
		decoder.wm.Log.Debugf("fees: %v", fees)
		decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:    sumRawTx.Coin,
			Account: sumRawTx.Account,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.StringFixed(decoder.wm.Decimal()),
			},
			Required: 1,
		}

		createErr = decoder.createRawTransaction(
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

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//check交易交易单基本字段
	err := CheckRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		log.Error("get HDKey from wallet wrapper failed, err=%v", err)
		return err
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	if len(rawTx.Signatures[rawTx.Account.AccountID]) != 1 {
		log.Error("signature failed in account[%v].", rawTx.Account.AccountID)
		return errors.New("signature failed in account.")
	}

	keySignature := rawTx.Signatures[rawTx.Account.AccountID][0]
	fromAddr := keySignature.Address

	childKey, _ := key.DerivedKeyWithPath(fromAddr.HDPath, decoder.wm.Config.CurveType)
	PrivateKey, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		log.Error("get private key bytes, err=", err)
		return err
	}

	tx_hash := common.FromHex(keySignature.Message) //TX.Hash

	signed, err := SignRawTransaction(PrivateKey, tx_hash)
	if err != nil {
		log.Error("signature error !")
		return nil
	}

	//TX.Sign = signed

	keySignature.Signature = hex.EncodeToString(signed)

	log.Debug("** pri:", hex.EncodeToString(PrivateKey))
	log.Debug("** tx_hash:", keySignature.Message)
	log.Debug("** Signature:", keySignature.Signature)

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//check交易交易单基本字段
	err := CheckRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	accountSig, exist := rawTx.Signatures[rawTx.Account.AccountID]
	if !exist {
		decoder.wm.Log.Std.Error("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	if len(accountSig) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	message := common.FromHex(accountSig[0].Message)     //TX.Hash
	signature := common.FromHex(accountSig[0].Signature) //TX.Sign
	publicKey := common.FromHex(accountSig[0].Address.PublicKey)
	//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
	verify_result := VerifyRawTransaction(publicKey, message, signature)
	if verify_result == owcrypt.SUCCESS {
		log.Debug("transaction verify passed")
		rawTx.IsCompleted = true

	} else {
		log.Debug("transaction verify failed")
		rawTx.IsCompleted = false
	}

	return nil
}

func (decoder *TransactionDecoder) SubmitSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	//check交易交易单基本字段
	err := CheckRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return nil, err
	}
	if len(rawTx.Signatures) != 1 {
		openwLogger.Log.Errorf("len of signatures error. ")
		return nil, errors.New("len of signatures error. ")
	}

	accSignatures, exist := rawTx.Signatures[rawTx.Account.AccountID]
	if !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return nil, errors.New("wallet signature not found ")
	}

	if len(accSignatures) == 0 {
		openwLogger.Log.Errorf("wallet[%v] signature is empty ", rawTx.Account.AccountID)
		return nil, errors.New("wallet signature not found ")
	}

	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	keySignature := accSignatures[0]

	trx, err := DecodeRawHexToTransaction(rawTx.RawHex)
	if err != nil {
		return nil, err
	}

	trx.Sign = common.FromHex(keySignature.Signature)

	submitRawHex, err := EncodeTransaction(trx)
	if err != nil {
		return nil, err
	}

	txid, err := decoder.wm.SubmitRawTransaction(submitRawHex)
	if err != nil {
		return nil, err
	} else {
		//广播成功后记录nonce值到本地
		//fmt.Printf("Submit Success , Save nonce To AddressExtParam!\n")
		wrapper.SetAddressExtParam(rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address, decoder.wm.FullName(), rawTx.Signatures[rawTx.Account.AccountID][0].Nonce)
	}
	rawTx.TxID = txid
	rawTx.IsSubmit = true

	decimals := int32(0)
	if rawTx.Coin.IsContract {
		decimals = int32(rawTx.Coin.Contract.Decimals)
	} else {
		decimals = int32(decoder.wm.Decimal())
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

	//fmt.Printf("rawTx=%+v\n", rawTx)

	return tx, nil
}


//createRawTransaction
func (decoder *TransactionDecoder) createRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction, addrBalance *AddrBalance, estimatefee *txFeeInfo, callData string) error {

	var (
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		keySignList      = make([]*openwallet.KeySignature, 0)
		amountStr        string
		destination      string
		TX *SubmitTransaction
	)

	isContract := rawTx.Coin.IsContract
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

	amount, _ := ConvertNasStringToWei(amountStr)

	//最终费率
	gasprice, err := ConverWeiStringToNasDecimal(estimatefee.GasPrice.String())
	if err != nil {
		log.Error("convert wei string to gas price failed, err=", err)
		return err
	}
	//最终手续费
	fee, err := ConverWeiStringToNasDecimal(estimatefee.Fee.String())
	if err != nil {
		log.Error("convert wei string to gas price failed, err=", err)
		return err
	}

	var nonce uint64
	//获取db记录的nonce并确认nonce值
	nonce_db, err := wrapper.GetAddressExtParam(addrBalance.Address, decoder.wm.FullName())
	if err != nil {
		return err
	}
	//判断nonce_db是否为空,为空则说明当前nonce是0
	if nonce_db == nil {
		nonce = 0
	} else {
		nonce = ow.NewString(nonce_db).UInt64()
	}

	nonce = decoder.wm.ConfirmTxdecodeNonce(addrBalance.Address, nonce)

	if isContract {

		return fmt.Errorf("can not support token transfer")
		////构建合约交易
		//amount, _ := ConvertFloatStringToBigInt(amountStr, tokenDecimals)
		//if addrBalance.TokenBalance.Cmp(amount) < 0 {
		//	return fmt.Errorf("the token balance: %s is not enough", amountStr)
		//}
		//
		//if addrBalance.Balance.Cmp(fee.Fee) < 0 {
		//	coinBalance, _ := ConverWeiStringToEthDecimal(addrBalance.Balance.String())
		//	return fmt.Errorf("the [%s] balance: %s is not enough to call smart contract", rawTx.Coin.Symbol, coinBalance)
		//}
		//
		//tx = types.NewTransaction(nonce, ethcommon.HexToAddress(rawTx.Coin.Contract.Address),
		//	big.NewInt(0), fee.GasLimit.Uint64(), fee.GasPrice, common.FromHex(callData))
	} else {

		//构建交易单
		TX, err = decoder.wm.CreateRawTransaction(addrBalance.Address, destination, Gaslimit, gasprice.Mul(coinDecimal).String(), amount.String(), nonce)
		if err != nil {
			return err
		}
	}

	rawHex, err := EncodeToTransactionRawHex(TX)
	if err != nil {
		return err
	}

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Nonce:   strconv.FormatUint(nonce, 10),
		Address: addr,
		Message: hex.EncodeToString(TX.Hash[:]),
	}
	keySignList = append(keySignList, &signature)

	feesDec, _ := decimal.NewFromString(rawTx.Fees)
	accountTotalSent = accountTotalSent.Add(feesDec)
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.RawHex = rawHex
	rawTx.Signatures[rawTx.Account.AccountID] = keySignList
	rawTx.FeeRate = gasprice.String()
	rawTx.Fees = fee.String()
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.StringFixed(decoder.wm.Decimal())
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}
