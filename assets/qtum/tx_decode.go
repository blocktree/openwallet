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

package qtum

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
	"github.com/blocktree/OpenWallet/assets/qtum/btcLikeTxDriver"
	"time"
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

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//先加载是否有配置文件
	//err := decoder.wm.loadConfig()
	//if err != nil {
	//	return err
	//}

	var (
		vins      = make([]btcLikeTxDriver.Vin, 0)
		vouts     = make([]btcLikeTxDriver.Vout, 0)
		txUnlocks = make([]btcLikeTxDriver.TxUnlock, 0)

		usedUTXO []*Unspent
		balance  = decimal.New(0, 0)
		//totalSend  = amounts
		totalSend    = decimal.New(0, 0)
		actualFees   = decimal.New(0, 0)
		feesRate     = decimal.New(0, 0)
		accountID    = rawTx.Account.AccountID
		destinations = make([]string, 0)
		DEFAULT_GAS_LIMIT = "250000"
		DEFAULT_GAS_PRICE = decimal.New(4, -7)

		coinDecimal = int32(0)
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
	)

	isTestNet := decoder.wm.config.isTestNet

	if rawTx.Coin.IsContract {
		coinDecimal = int32(rawTx.Coin.Contract.Decimals)
	} else {
		coinDecimal = int32(decoder.wm.Decimal())
	}

	address, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)
	if err != nil {
		return err
	}

	if len(address) == 0 {
		return fmt.Errorf("[%s] account: %s has not addresses", decoder.wm.Symbol(), accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range address {
		searchAddrs = append(searchAddrs, address.Address)
	}
	log.Debug(searchAddrs)

	//log.Info("Calculating wallet unspent record to build transaction...")
	//查找账户的utxo
	unspents, err := decoder.wm.ListUnspent(0, searchAddrs...)
	if err != nil {
		return err
	}



	if len(rawTx.To) == 0 {
		return errors.New("Receiver addresses is empty!")
	}

	//获取utxo，按小到大排序
	sort.Sort(UnspentSort{unspents, func(a, b *Unspent) int {

		if a.Amount > b.Amount {
			return 1
		} else {
			return -1
		}
	}})

	//totalBalance, _ := decimal.NewFromString(decoder.wm.GetWalletBalance(w.WalletID))
	//if totalBalance.LessThanOrEqual(totalSend) {
	//	return "", errors.New("The wallet's balance is not enough!")
	//}

	//创建找零地址
	//changeAddrs, err := wrapper.CreateAddress(accountID, 1, decoder.wm.Decoder, true, decoder.wm.Config.IsTestNet)
	////changeAddr, err := decoder.wm.CreateChangeAddress(walletID, key)
	//if err != nil {
	//	return err
	//}
	//
	//changeAddress := changeAddrs[0]

	//计算总发送金额
	for addr, amount := range rawTx.To {
		deamount, _ := decimal.NewFromString(amount)
		totalSend = totalSend.Add(deamount)
		destinations = append(destinations, addr)

		//计算账户的实际转账amount
		addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", addr)
		if findErr != nil || len(addresses) == 0 {
			accountTotalSent = accountTotalSent.Add(deamount)
		}
	}

	computeTotalSend := totalSend

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	var emptyTrans string

	//fmt.Printf("IsContract: %v\n", rawTx.Coin.IsContract)

	if rawTx.Coin.IsContract {

		var (
			to             string
		    amount         string
		    txInTotal      uint64
			deAmount       decimal.Decimal
			sendAmount     decimal.Decimal
			totalQtum      decimal.Decimal
			unspent        decimal.Decimal
			change         uint64
			contractFees   decimal.Decimal
			gasPrice       string
		)


		trimContractAddr := strings.TrimPrefix(rawTx.Coin.Contract.Address, "0x")

		if len(rawTx.FeeRate) == 0 {
			feesRate, err = decoder.wm.EstimateFeeRate()
			if err != nil {
				return err
			}
		} else {
			feesRate, _ = decimal.NewFromString(rawTx.FeeRate)
		}

		log.Debugf("feesRate:%v",feesRate)

		usedTokenUTXO := make([]*Unspent, 0)
		unspent = decimal.New(0, 0)

		computeTotalfee := feesRate
		//log.Debugf("computeTotalfee:%v",computeTotalfee)
		log.Info("Calculating wallet unspent record to build transaction...")
		//循环的计算余额是否足够支付发送数额+手续费
		for {

			usedUTXO = make([]*Unspent, 0)
			//balance = decimal.New(0, 0)
			totalQtum = decimal.New(0, 0)
			contractFees = decimal.New(0, 0)

			//计算一个可用于支付的余额
			for _, u := range unspents {

				if u.Spendable {

					usedTokenUTXO = make([]*Unspent, 0)
					//unspent = decimal.New(0, 0)
					unspent, err = decoder.wm.GetQRC20Balance(rawTx.Coin.Contract, u.Address, isTestNet)
					if err != nil {
						log.Errorf("GetTokenUnspentByAddress failed unexpected error: %v\n", err)
					}

					if unspent.GreaterThanOrEqual(totalSend) {
						usedTokenUTXO = append(usedTokenUTXO, u)
						ua, _ := decimal.NewFromString(u.Amount)
						totalQtum = totalQtum.Add(ua)
						usedUTXO = append(usedUTXO, u)

						//选择一个地址作为发送
						txFrom = []string{fmt.Sprintf("%s:%s", u.Address, totalSend.StringFixed(coinDecimal))}

						break
					}
				}
			}

			if unspent.LessThan(totalSend) {
				return fmt.Errorf("The token balance is not enough! There must be enough balance in one address.")
			}

			//计算用于支付手续费的UTXO
			for _, u := range unspents {

				if u.Spendable && u.TxID != usedTokenUTXO[0].TxID{

					if totalQtum.GreaterThanOrEqual(computeTotalfee){
						break
					}
					ua, _ := decimal.NewFromString(u.Amount)
					totalQtum = totalQtum.Add(ua)
					usedUTXO = append(usedUTXO, u)


				}
			}

			if totalQtum.LessThan(computeTotalfee) {
				return fmt.Errorf("The fees(Qtum): %s is not enough! ", totalQtum.StringFixed(decoder.wm.Decimal()))
			}

			//计算手续费，找零地址有2个，一个是发送，一个是新创建的
			fees, err := decoder.wm.EstimateFee(int64(len(usedUTXO)), int64(len(destinations)+1), feesRate)
			if err != nil {
				return err
			}

			//合约手续费在普通交易基础上加上0.1个qtum, 小数点8位
			//contractFees = fees.Add(decimal.New(1, -1)).Mul(decoder.wm.config.CoinDecimal)
			contractFees = fees.Add(decimal.New(1, -1))

			//如果要手续费有发送支付，得计算加入手续费后，计算余额是否足够
			//总共要发送的
			computeTotalfee = contractFees
			log.Debugf("computeTotalfee:%v, totalQtum:%v",computeTotalfee,totalQtum)
			if computeTotalfee.GreaterThan(totalQtum) {
				continue
			}

			actualFees = contractFees.Mul(decoder.wm.config.CoinDecimal)

			rawTx.Fees = contractFees.StringFixed(decoder.wm.Decimal())
			rawTx.FeeRate = feesRate.StringFixed(decoder.wm.Decimal())

			break

		}

		//UTXO如果大于设定限制，则分拆成多笔交易单发送
		if len(usedUTXO) > decoder.wm.config.maxTxInputs {
			errStr := fmt.Sprintf("The transaction is use max inputs over: %d", decoder.wm.config.maxTxInputs)
			return errors.New(errStr)
		}

		//装配输入
		for _, utxo := range usedUTXO {
			in := btcLikeTxDriver.Vin{utxo.TxID, uint32(utxo.Vout)}
			vins = append(vins, in)
			txUnlock := btcLikeTxDriver.TxUnlock{LockScript: utxo.ScriptPubKey, Address: utxo.Address}
			txUnlocks = append(txUnlocks, txUnlock)
			tempAmount, _ := decimal.NewFromString(utxo.Amount)
			tempAmount = tempAmount.Mul(decoder.wm.config.CoinDecimal)
			amount2 := uint64(tempAmount.IntPart())
			txInTotal += amount2
		}

		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
		//fees, err := decoder.wm.EstimateFee(int64(len(usedUTXO)), int64(len(destinations)+1), feesRate)
		//if err != nil {
		//	return err
		//}

		if txInTotal >= uint64(actualFees.IntPart()) {
			change = txInTotal - uint64(actualFees.IntPart())
		}else {
			return fmt.Errorf("The fees(Qtum): %s is not enough! ", totalQtum.StringFixed(decoder.wm.Decimal()))
		}

		//装配输出
		for to, amount = range rawTx.To {
			deAmount, _ = decimal.NewFromString(amount)
			sendAmount = deAmount.Mul(decimal.New(1, coinDecimal))
			//deamount = deamount.Mul(decoder.wm.config.CoinDecimal)
			out := btcLikeTxDriver.Vout{usedUTXO[0].Address, change}
			vouts = append(vouts, out)

			txTo = []string{fmt.Sprintf("%s:%s", to, amount)}
		}

		if len(vouts)!= 1 {
			return errors.New("error: the number of change addresses must be equal to one. ")
		}

		//gasPrice
		//gasPriceDec, _ := decimal.NewFromString(rawTx.FeeRate)
		//if rawTx.FeeRate == "" || gasPriceDec.LessThan(DEFAULT_GAS_PRICE) {
		//	gasPriceDec = DEFAULT_GAS_PRICE
		//}
		SotashiGasPriceDec := DEFAULT_GAS_PRICE.Mul(decoder.wm.config.CoinDecimal)
		gasPrice = SotashiGasPriceDec.String()

		//装配合约
		vcontract := btcLikeTxDriver.Vcontract{trimContractAddr, to, sendAmount, DEFAULT_GAS_LIMIT, gasPrice, 0}

		//构建空合约交易单
		emptyTrans, err = btcLikeTxDriver.CreateQRC20TokenEmptyRawTransaction(vins, vcontract, vouts, lockTime, replaceable, isTestNet)
		if err != nil {
			return err
			//log.Error("构建空交易单失败")
		}

	}else {

		if len(rawTx.FeeRate) == 0 {
			feesRate, err = decoder.wm.EstimateFeeRate()
			if err != nil {
				return err
			}
		} else {
			feesRate, _ = decimal.NewFromString(rawTx.FeeRate)
			relayfee := decimal.NewFromFloat(0.004)
			if feesRate.LessThan(relayfee) {
				feesRate, err = decoder.wm.EstimateFeeRate()
				if err != nil {
					return err
				}
			}
		}

		log.Info("Calculating wallet unspent record to build transaction...")
		//循环的计算余额是否足够支付发送数额+手续费
		for {

			usedUTXO = make([]*Unspent, 0)
			balance = decimal.New(0, 0)

			//计算一个可用于支付的余额
			for _, u := range unspents {

				if u.Spendable {
					ua, _ := decimal.NewFromString(u.Amount)
					balance = balance.Add(ua)
					usedUTXO = append(usedUTXO, u)
					if balance.GreaterThanOrEqual(computeTotalSend) {
						break
					}
				}
			}

			if balance.LessThan(computeTotalSend) {
				return fmt.Errorf("The balance: %s is not enough! ", balance.StringFixed(decoder.wm.Decimal()))
			}

			//计算手续费，找零地址有2个，一个是发送，一个是新创建的
			fees, err := decoder.wm.EstimateFee(int64(len(usedUTXO)), int64(len(destinations)+1), feesRate)
			if err != nil {
				return err
			}

			//log.Debugf("fees:%v",fees)
			//如果要手续费有发送支付，得计算加入手续费后，计算余额是否足够
			//总共要发送的
			computeTotalSend = totalSend.Add(fees)
			if computeTotalSend.GreaterThan(balance) {
				continue
			}
			computeTotalSend = totalSend

			actualFees = fees

			break

		}

		changeAmount := balance.Sub(computeTotalSend).Sub(actualFees)
		rawTx.Fees = actualFees.StringFixed(decoder.wm.Decimal())
		rawTx.FeeRate = feesRate.StringFixed(decoder.wm.Decimal())
		//UTXO如果大于设定限制，则分拆成多笔交易单发送
		if len(usedUTXO) > decoder.wm.config.maxTxInputs {
			errStr := fmt.Sprintf("The transaction is use max inputs over: %d", decoder.wm.config.maxTxInputs)
			return errors.New(errStr)
		}


		//取账户最后一个地址
		changeAddress := usedUTXO[0].Address


		log.Std.Notice("-----------------------------------------------")
		log.Std.Notice("From Account: %s", accountID)
		log.Std.Notice("To Address: %s", strings.Join(destinations, ", "))
		log.Std.Notice("Use: %v", balance.StringFixed(8))
		log.Std.Notice("Fees: %v", actualFees.StringFixed(8))
		log.Std.Notice("Receive: %v", computeTotalSend.StringFixed(8))
		log.Std.Notice("Change: %v", changeAmount.StringFixed(8))
		log.Std.Notice("Change Address: %v", changeAddress)
		log.Std.Notice("-----------------------------------------------")

		//装配输入
		for _, utxo := range usedUTXO {
			in := btcLikeTxDriver.Vin{utxo.TxID, uint32(utxo.Vout)}
			vins = append(vins, in)

			txUnlock := btcLikeTxDriver.TxUnlock{LockScript: utxo.ScriptPubKey, Address: utxo.Address}
			txUnlocks = append(txUnlocks, txUnlock)

			txFrom = append(txFrom, fmt.Sprintf("%s:%s", utxo.Address, utxo.Amount))
		}

		//装配输出
		for to, amount := range rawTx.To {
			deamount, _ := decimal.NewFromString(amount)
			deamount = deamount.Mul(decoder.wm.config.CoinDecimal)
			out := btcLikeTxDriver.Vout{to, uint64(deamount.IntPart())}
			vouts = append(vouts, out)

			txTo = append(txTo, fmt.Sprintf("%s:%s", to, amount))
		}

		//changeAmount := balance.Sub(totalSend).Sub(actualFees)
		if changeAmount.GreaterThan(decimal.New(0, 0)) {
			deamount := changeAmount.Mul(decoder.wm.config.CoinDecimal)
			out := btcLikeTxDriver.Vout{changeAddress, uint64(deamount.IntPart())}
			vouts = append(vouts, out)

			txTo = append(txTo, fmt.Sprintf("%s:%s", changeAddress, changeAmount.StringFixed(decoder.wm.Decimal())))
			//fmt.Printf("Create change address for receiving %s coin.", outputs[change])
		}
		/////////构建空交易单
		emptyTrans, err = btcLikeTxDriver.CreateEmptyRawTransaction(vins, vouts, lockTime, replaceable, isTestNet)
		if err != nil {
			return fmt.Errorf("create transaction failed, unexpected error: %v", err)
			//log.Error("构建空交易单失败")
		}

		accountTotalSent = accountTotalSent.Add(actualFees)
	}

	////////构建用于签名的交易单哈希
	transHash, err := btcLikeTxDriver.CreateRawTransactionHashForSig(emptyTrans, txUnlocks)
	if err != nil {
		return fmt.Errorf("create transaction hash for sig failed, unexpected error: %v", err)
		//log.Error("获取待签名交易单哈希失败")
	}

	rawTx.RawHex = emptyTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	//装配签名
	keySigs := make([]*openwallet.KeySignature, 0)

	for i, unlock := range txUnlocks {

		beSignHex := transHash[i]

		addr, err := wrapper.GetAddress(unlock.Address)
		if err != nil {
			return err
		}

		signature := openwallet.KeySignature{
			EccType: decoder.wm.config.CurveType,
			Nonce:   "",
			Address: addr,
			Message: beSignHex,
		}

		keySigs = append(keySigs, &signature)

	}

	//TODO:多重签名要使用owner的公钥填充

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true
	rawTx.TxAmount = "-" + accountTotalSent.StringFixed(coinDecimal)
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//先加载是否有配置文件
	//err := decoder.wm.loadConfig()
	//if err != nil {
	//	return err
	//}

	var (
		txUnlocks  = make([]btcLikeTxDriver.TxUnlock, 0)
		emptyTrans = rawTx.RawHex
		transHash  = make([]string, 0)
		sigPub     = make([]btcLikeTxDriver.SignaturePubkey, 0)
		privateKeys = make([][]byte, 0)
	)

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		return err
	}

	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	if keySignatures != nil {
		for _, keySignature := range keySignatures {

			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
			keyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				return err
			}
			privateKeys = append(privateKeys, keyBytes)
			transHash = append(transHash, keySignature.Message)
		}
	}

	txBytes, err := hex.DecodeString(emptyTrans)
	if err != nil {
		return errors.New("Invalid transaction hex data!")
	}

	trx, err := btcLikeTxDriver.DecodeRawTransaction(txBytes)
	if err != nil {
		return errors.New("Invalid transaction data! ")
	}

	for i, vin := range trx.Vins {

		utxo, err := decoder.wm.GetTxOut(vin.GetTxID(), uint64(vin.GetVout()))
		if err != nil {
			return err
		}

		keyBytes := privateKeys[i]

		txUnlock := btcLikeTxDriver.TxUnlock{
			LockScript: utxo.ScriptPubKey,
			PrivateKey: keyBytes,
		}
		txUnlocks = append(txUnlocks, txUnlock)


	}

	//log.Debug("transHash len:", len(transHash))
	//log.Debug("txUnlocks len:", len(txUnlocks))

	/////////交易单哈希签名
	sigPub, err = btcLikeTxDriver.SignRawTransactionHash(transHash, txUnlocks)
	if err != nil {
		return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
	} else {
		log.Info("transaction hash sign success")
		//for i, s := range sigPub {
		//	log.Info("第", i+1, "个签名结果")
		//	log.Info()
		//	log.Info("对应的公钥为")
		//	log.Info(hex.EncodeToString(s.Pubkey))
		//}
	}

	if len(sigPub) != len(keySignatures) {
		return fmt.Errorf("sign raw transaction fail, program error. ")
	}

	for i, keySignature := range keySignatures {
		keySignature.Signature = hex.EncodeToString(sigPub[i].Signature)
		//log.Debug("keySignature.Signature:",i, "=", keySignature.Signature)
	}

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	//log.Info("rawTx.Signatures 1:", rawTx.Signatures)

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//先加载是否有配置文件
	//err := decoder.wm.loadConfig()
	//if err != nil {
	//	return err
	//}

	var (
		txUnlocks  = make([]btcLikeTxDriver.TxUnlock, 0)
		emptyTrans = rawTx.RawHex
		sigPub     = make([]btcLikeTxDriver.SignaturePubkey, 0)
	)

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	//TODO:待支持多重签名

	for accountID, keySignatures := range rawTx.Signatures {
		log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			pubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			signaturePubkey := btcLikeTxDriver.SignaturePubkey{
				Signature: signature,
				Pubkey:    pubkey,
			}

			sigPub = append(sigPub, signaturePubkey)

			log.Debug("Signature:", keySignature.Signature)
			log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	txBytes, err := hex.DecodeString(emptyTrans)
	if err != nil {
		return errors.New("Invalid transaction hex data!")
	}

	trx, err := btcLikeTxDriver.DecodeRawTransaction(txBytes)
	if err != nil {
		return errors.New("Invalid transaction data! ")
	}

	for _, vin := range trx.Vins {

		utxo, err := decoder.wm.GetTxOut(vin.GetTxID(), uint64(vin.GetVout()))
		if err != nil {
			return err
		}

		txUnlock := btcLikeTxDriver.TxUnlock{LockScript: utxo.ScriptPubKey}
		txUnlocks = append(txUnlocks, txUnlock)

	}

	//log.Debug(emptyTrans)

	////////填充签名结果到空交易单
	//  传入TxUnlock结构体的原因是： 解锁向脚本支付的UTXO时需要对应地址的赎回脚本， 当前案例的对应字段置为 "" 即可
	signedTrans, err := btcLikeTxDriver.InsertSignatureIntoEmptyTransaction(emptyTrans, sigPub, txUnlocks)
	if err != nil {
		return fmt.Errorf("transaction compose signatures failed")
	}
	//else {
	//	//	fmt.Println("拼接后的交易单")
	//	//	fmt.Println(signedTrans)
	//	//}

	/////////验证交易单
	//验证时，对于公钥哈希地址，需要将对应的锁定脚本传入TxUnlock结构体
	pass := btcLikeTxDriver.VerifyRawTransaction(signedTrans, txUnlocks)
	if pass {
		log.Debug("transaction verify passed")
		rawTx.IsCompleted = true
		rawTx.RawHex = signedTrans
	} else {
		log.Debug("transaction verify failed")
		rawTx.IsCompleted = false
	}

	return nil
}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	//先加载是否有配置文件
	//err := decoder.wm.loadConfig()
	//if err != nil {
	//	return err
	//}

	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	txid, err := decoder.wm.SendRawTransaction(rawTx.RawHex)
	if err != nil {
		return nil, err
	}

	decimals := int32(0)
	fees := "0"
	if rawTx.Coin.IsContract {
		decimals = int32(rawTx.Coin.Contract.Decimals)
		fees = "0"
	} else {
		decimals = int32(decoder.wm.Decimal())
		fees = rawTx.Fees
	}

	rawTx.TxID = txid
	rawTx.IsSubmit = true

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	return tx, nil
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	rate, err := decoder.wm.EstimateFeeRate()
	if err != nil {
		return "", "", err
	}

	return rate.StringFixed(decoder.wm.Decimal()), "K", nil
}

