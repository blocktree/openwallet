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

package bitcoin

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcdrivers/btcTransaction"
	"github.com/blocktree/go-owcdrivers/omniTransaction"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
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
	if rawTx.Coin.IsContract {
		return decoder.CreateOmniRawTransaction(wrapper, rawTx)
	} else {
		return decoder.CreateBTCRawTransaction(wrapper, rawTx)
	}
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if rawTx.Coin.IsContract {
		return decoder.SignOmniRawTransaction(wrapper, rawTx)
	} else {
		return decoder.SignBTCRawTransaction(wrapper, rawTx)
	}
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if rawTx.Coin.IsContract {
		return decoder.VerifyOmniRawTransaction(wrapper, rawTx)
	} else {
		return decoder.VerifyBTCRawTransaction(wrapper, rawTx)
	}
}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//先加载是否有配置文件
	//err := decoder.wm.LoadConfig()
	//if err != nil {
	//	return err
	//}

	if len(rawTx.RawHex) == 0 {
		return fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return fmt.Errorf("transaction is not completed validation")
	}

	txid, err := decoder.wm.SendRawTransaction(rawTx.RawHex)
	if err != nil {
		return err
	}

	rawTx.TxID = txid
	rawTx.IsSubmit = true

	return nil
}

////////////////////////// BTC implement //////////////////////////

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateBTCRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		vins      = make([]btcTransaction.Vin, 0)
		vouts     = make([]btcTransaction.Vout, 0)
		txUnlocks = make([]btcTransaction.TxUnlock, 0)

		usedUTXO []*Unspent
		balance  = decimal.New(0, 0)
		//totalSend  = amounts
		totalSend    = decimal.New(0, 0)
		actualFees   = decimal.New(0, 0)
		feesRate     = decimal.New(0, 0)
		accountID    = rawTx.Account.AccountID
		destinations = make([]string, 0)
	)

	address, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)
	if err != nil {
		return err
	}

	if len(address) == 0 {
		return fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range address {
		searchAddrs = append(searchAddrs, address.Address)
	}
	//log.Debug(searchAddrs)
	//查找账户的utxo
	unspents, err := decoder.wm.ListUnspent(0, searchAddrs...)
	if err != nil {
		return err
	}

	if len(unspents) == 0 {
		return fmt.Errorf("[%s] balance is not enough", accountID)
	}

	if len(rawTx.To) == 0 {
		return errors.New("Receiver addresses is empty!")
	}

	//计算总发送金额
	for addr, amount := range rawTx.To {
		deamount, _ := decimal.NewFromString(amount)
		totalSend = totalSend.Add(deamount)
		destinations = append(destinations, addr)
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

	if len(rawTx.FeeRate) == 0 {
		feesRate, err = decoder.wm.EstimateFeeRate()
		if err != nil {
			return err
		}
	} else {
		feesRate, _ = decimal.NewFromString(rawTx.FeeRate)
	}

	log.Info("Calculating wallet unspent record to build transaction...")
	computeTotalSend := totalSend
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

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if len(usedUTXO) > decoder.wm.Config.MaxTxInputs {
		errStr := fmt.Sprintf("The transaction is use max inputs over: %d", decoder.wm.Config.MaxTxInputs)
		return errors.New(errStr)
	}

	//取账户最后一个地址
	changeAddress := usedUTXO[0].Address

	changeAmount := balance.Sub(computeTotalSend).Sub(actualFees)
	rawTx.FeeRate = feesRate.StringFixed(decoder.wm.Decimal())
	rawTx.Fees = actualFees.StringFixed(decoder.wm.Decimal())

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
		in := btcTransaction.Vin{utxo.TxID, uint32(utxo.Vout)}
		vins = append(vins, in)

		txUnlock := btcTransaction.TxUnlock{LockScript: utxo.ScriptPubKey, SigType: btcTransaction.SigHashAll}
		txUnlocks = append(txUnlocks, txUnlock)
	}

	//装配输入
	for to, amount := range rawTx.To {
		deamount, _ := decimal.NewFromString(amount)
		deamount = deamount.Mul(decoder.wm.Config.CoinDecimal)
		out := btcTransaction.Vout{to, uint64(deamount.IntPart())}
		vouts = append(vouts, out)
	}

	//changeAmount := balance.Sub(totalSend).Sub(actualFees)
	if changeAmount.GreaterThan(decimal.New(0, 0)) {
		deamount := changeAmount.Mul(decoder.wm.Config.CoinDecimal)
		out := btcTransaction.Vout{changeAddress, uint64(deamount.IntPart())}
		vouts = append(vouts, out)

		//fmt.Printf("Create change address for receiving %s coin.", outputs[change])
	}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	/////////构建空交易单
	emptyTrans, err := btcTransaction.CreateEmptyRawTransaction(vins, vouts, lockTime, replaceable)

	if err != nil {
		return fmt.Errorf("create transaction failed, unexpected error: %v", err)
		//log.Error("构建空交易单失败")
	}

	////////构建用于签名的交易单哈希
	transHash, err := btcTransaction.CreateRawTransactionHashForSig(emptyTrans, txUnlocks, decoder.wm.Config.SupportSegWit)
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

	for i, txHash := range transHash {

		var unlockAddr string

		//txHash := transHash[i]

		//判断是否是多重签名
		if txHash.IsMultisig() {
			//获取地址
			//unlockAddr = txHash.GetMultiTxPubkeys() //返回hex数组
		} else {
			//获取地址
			unlockAddr = txHash.GetNormalTxAddress() //返回hex串
		}
		//获取hash值
		beSignHex := txHash.GetTxHashHex()

		log.Std.Debug("txHash[%d]: %s", i, beSignHex)
		//beSignHex := transHash[i]

		addr, err := wrapper.GetAddress(unlockAddr)
		if err != nil {
			return err
		}

		signature := openwallet.KeySignature{
			EccType: decoder.wm.Config.CurveType,
			Nonce:   "",
			Address: addr,
			Message: beSignHex,
		}

		keySigs = append(keySigs, &signature)

	}

	//TODO:多重签名要使用owner的公钥填充

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true

	return nil
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignBTCRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//先加载是否有配置文件
	//err := decoder.wm.LoadConfig()
	//if err != nil {
	//	return err
	//}

	var (
	//txUnlocks   = make([]btcTransaction.TxUnlock, 0)
	//emptyTrans  = rawTx.RawHex
	//transHash   = make([]btcTransaction.TxHash, 0)
	//sigPub      = make([]btcTransaction.SignaturePubkey, 0)
	//privateKeys = make([][]byte, 0)
	)

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
			log.Debug("privateKey:", hex.EncodeToString(keyBytes))

			//privateKeys = append(privateKeys, keyBytes)
			txHash := btcTransaction.TxHash{
				Hash: keySignature.Message,
				Normal: &btcTransaction.NormalTx{
					Address: keySignature.Address.Address,
					SigType: btcTransaction.SigHashAll,
				},
			}
			//transHash = append(transHash, txHash)

			log.Debug("hash:", txHash.GetTxHashHex())

			//签名交易
			/////////交易单哈希签名
			sigPub, err := btcTransaction.SignRawTransactionHash(txHash.GetTxHashHex(), keyBytes)
			if err != nil {
				return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
			} else {

				//for i, s := range sigPub {
				//	log.Info("第", i+1, "个签名结果")
				//	log.Info()
				//	log.Info("对应的公钥为")
				//	log.Info(hex.EncodeToString(s.Pubkey))
				//}

				//txHash.Normal.SigPub = *sigPub
			}

			keySignature.Signature = hex.EncodeToString(sigPub.Signature)
		}
	}

	log.Info("transaction hash sign success")

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	//log.Info("rawTx.Signatures 1:", rawTx.Signatures)

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyBTCRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//先加载是否有配置文件
	//err := decoder.wm.LoadConfig()
	//if err != nil {
	//	return err
	//}

	var (
		txUnlocks  = make([]btcTransaction.TxUnlock, 0)
		emptyTrans = rawTx.RawHex
		//sigPub     = make([]btcTransaction.SignaturePubkey, 0)
		transHash = make([]btcTransaction.TxHash, 0)
	)

	//TODO:待支持多重签名

	for accountID, keySignatures := range rawTx.Signatures {
		log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			pubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			signaturePubkey := btcTransaction.SignaturePubkey{
				Signature: signature,
				Pubkey:    pubkey,
			}

			//sigPub = append(sigPub, signaturePubkey)

			txHash := btcTransaction.TxHash{
				Hash: keySignature.Message,
				Normal: &btcTransaction.NormalTx{
					Address: keySignature.Address.Address,
					SigType: btcTransaction.SigHashAll,
					SigPub:  signaturePubkey,
				},
			}

			transHash = append(transHash, txHash)

			log.Debug("Signature:", keySignature.Signature)
			log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	txBytes, err := hex.DecodeString(emptyTrans)
	if err != nil {
		return errors.New("Invalid transaction hex data!")
	}

	trx, err := btcTransaction.DecodeRawTransaction(txBytes, decoder.wm.Config.SupportSegWit)
	if err != nil {
		return errors.New("Invalid transaction data! ")
	}

	for _, vin := range trx.Vins {

		utxo, err := decoder.wm.GetTxOut(vin.GetTxID(), uint64(vin.GetVout()))
		if err != nil {
			return err
		}

		txUnlock := btcTransaction.TxUnlock{
			LockScript: utxo.ScriptPubKey,
			SigType:    btcTransaction.SigHashAll}
		txUnlocks = append(txUnlocks, txUnlock)

	}

	//log.Debug(emptyTrans)

	////////填充签名结果到空交易单
	//  传入TxUnlock结构体的原因是： 解锁向脚本支付的UTXO时需要对应地址的赎回脚本， 当前案例的对应字段置为 "" 即可
	signedTrans, err := btcTransaction.InsertSignatureIntoEmptyTransaction(emptyTrans, transHash, txUnlocks, decoder.wm.Config.SupportSegWit)
	if err != nil {
		return fmt.Errorf("transaction compose signatures failed")
	}
	//else {
	//	//	fmt.Println("拼接后的交易单")
	//	//	fmt.Println(signedTrans)
	//	//}

	/////////验证交易单
	//验证时，对于公钥哈希地址，需要将对应的锁定脚本传入TxUnlock结构体
	pass := btcTransaction.VerifyRawTransaction(signedTrans, txUnlocks, decoder.wm.Config.SupportSegWit)
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

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	rate, err := decoder.wm.EstimateFeeRate()
	if err != nil {
		return "", "", err
	}

	return rate.StringFixed(decoder.wm.Decimal()), "K", nil
}

////////////////////////// omnicore implement //////////////////////////

//CreateOmniRawTransaction 创建Omni交易单
func (decoder *TransactionDecoder) CreateOmniRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		vins      = make([]omniTransaction.Vin, 0)
		vouts     = make([]omniTransaction.Vout, 0)
		txUnlocks = make([]omniTransaction.TxUnlock, 0)

		toAddress string
		toAmount  decimal.Decimal = decimal.Zero

		availableUTXO     = make([]*Unspent, 0)
		usedUTXO          = make([]*Unspent, 0)
		totalTokenBalance = decimal.Zero
		useTokenBalance   = decimal.Zero
		useTokenAddress   = ""
		missUtxoAddress   = ""
		missToken  = make([]string, 0)
		balance    = decimal.New(0, 0)
		actualFees = decimal.New(0, 0)
		feesRate   = decimal.New(0, 0)
		accountID  = rawTx.Account.AccountID
	)

	if !decoder.wm.Config.OmniSupport {
		return fmt.Errorf("%s is not support omnicore transfer", decoder.wm.Symbol())
	}

	if len(rawTx.Coin.Contract.Address) == 0 {
		return fmt.Errorf("contract address is empty")
	}

	//Omni代币编号
	propertyID := common.NewString(rawTx.Coin.Contract.Address).UInt64()
	tokenCoin := rawTx.Coin.Contract.Token
	tokenDecimals := int32(rawTx.Coin.Contract.Decimals)
	//转账最低成本
	transferCost, _ := decimal.NewFromString(decoder.wm.Config.OmniTransferCost)

	address, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)
	if err != nil {
		return err
	}

	if len(address) == 0 {
		return fmt.Errorf("[%s] have not addresses", accountID)
	}

	if len(rawTx.To) == 0 {
		return errors.New("Receiver addresses is empty!")
	}

	//omni限制只能发送目标一个
	if len(rawTx.To) > 1 {
		return fmt.Errorf("ommni transfer not support multiple receiver address")
	}

	//选择一个输出地址
	for to, amount := range rawTx.To {
		toAddress = to
		toAmount, _ = decimal.NewFromString(amount)
	}

	/*

		1. 遍历所有地址，获取token余额。
			1）累计所有地址token余额。
			2）没有token余额的地址，加入到没有token余额数组，其地址有utxo可用于手续费使用。
			3）判断是否可用Token余额超过可发送数量，是则继续遍历。
			4）记录最大的可余额，记录可用token余额地址。
			5）查询有token余额的地址的utxo，没有utxo，记录缺少utxo地址，记录可用utxo。
		2. 遍历所有地址后，全部Token余额不足， 返回错误Token余额不足。
		3. 全部Token余额足够，可用余额不足，返回地址的最大Token余额不足。
		4. 全部Token余额足够，可用余额足够，地址的没有utxo，返回错误。
		5. 先排序可用utxo，再把没有token余额的地址所有utxo查出来，加入到可用utxo数组最后。
		6. 最后可用的utxo不足，返回错误主链币不足。
		7. 可用token和可用utxo足够，构建交易单。


	*/

	for _, address := range address {

		//查找地址token余额
		tokenBalance, checkErr := decoder.wm.GetOmniBalance(propertyID, address.Address)
		if checkErr != nil {
			return checkErr
		}

		//合计总token余额
		totalTokenBalance = totalTokenBalance.Add(tokenBalance)

		//没有余额的地址可以作为手续费使用
		if tokenBalance.LessThanOrEqual(decimal.Zero) {
			missToken = append(missToken, address.Address)
			continue
		} else {

			//判断是否可用Token余额超过可发送数量，则停止遍历
			if useTokenBalance.GreaterThanOrEqual(toAmount) && len(missUtxoAddress) == 0 {
				continue
			}

			//totalTokenBalance = totalTokenBalance.Add(tokenBalance)

			//查找账户的utxo
			unspents, tokenErr := decoder.wm.ListUnspent(0, address.Address)
			if tokenErr != nil {
				return err
			}

			//取最大余额
			//if tokenBalance.GreaterThanOrEqual(toAmount) {
			useTokenBalance = tokenBalance
			useTokenAddress = address.Address
			//}

			//记录缺少utxo的地址
			if len(unspents) == 0 {
				missUtxoAddress = address.Address
				//log.Debug("missUtxoAddress:", missUtxoAddress)
			} else {
				missUtxoAddress = ""
			}
			//else {
				//useTokenBalance = tokenBalance
				//useTokenBalance = useTokenBalance.Add(tokenBalance)
			//}

			availableUTXO = unspents
			//availableUTXO = append(availableUTXO, unspents...)

		}

	}

	//遍历所有地址后，全部Token余额不足， 返回错误Token余额不足
	if totalTokenBalance.LessThan(toAmount) {
		return fmt.Errorf("account[%s] omni[%s] total balance: %s is not enough! ", accountID, tokenCoin, totalTokenBalance.StringFixed(tokenDecimals))
	}

	//单个地址的可用Token余额不足够
	if useTokenBalance.LessThan(toAmount) {
		return fmt.Errorf("account[%s] omni[%s] total balance is enough, but the available balance: %s of address[%s] is not enough! ", accountID, tokenCoin, useTokenBalance.StringFixed(tokenDecimals), useTokenAddress)
	} else {
		//可用Token余额足够，但没有utxo，返回错误及有Token余额的地址没有可用主链币
		if len(missUtxoAddress) > 0 {
			return fmt.Errorf("account[%s] omni[%s] total balance is enough, but the utxo of address[%s] is empty! ", accountID, tokenCoin, missUtxoAddress)
		}
	}

	//获取utxo，按小到大排序
	sort.Sort(UnspentSort{availableUTXO, func(a, b *Unspent) int {

		if a.Amount > b.Amount {
			return 1
		} else {
			return -1
		}
	}})

	//查找账户没有token余额的utxo，可用于手续费
	if len(missToken) > 0 {
		missTokenUnspents, err := decoder.wm.ListUnspent(0, missToken...)
		if err != nil {
			return err
		}

		availableUTXO = append(availableUTXO, missTokenUnspents...)
	}

	//获取手续费率
	if len(rawTx.FeeRate) == 0 {
		feesRate, err = decoder.wm.EstimateFeeRate()
		if err != nil {
			return err
		}
	} else {
		feesRate, _ = decimal.NewFromString(rawTx.FeeRate)
	}

	log.Info("Calculating wallet unspent record to build transaction...")
	computeTotalSend := transferCost
	//循环的计算余额是否足够支付发送数额+手续费
	for {

		usedUTXO = make([]*Unspent, 0)
		balance = decimal.Zero

		//计算一个可用于支付的余额
		for _, u := range availableUTXO {

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
			return fmt.Errorf("The [%s] available utxo balance: %s is not enough! ", decoder.wm.Symbol(), balance.StringFixed(decoder.wm.Decimal()))
		}

		//计算手续费，输出地址有2个，一个是发送，一个是找零，一个是op_reture
		fees, err := decoder.wm.EstimateFee(int64(len(usedUTXO)), int64(3), feesRate)
		if err != nil {
			return err
		}

		//如果要手续费有发送支付，得计算加入手续费后，计算余额是否足够
		//总共要发送的
		computeTotalSend = transferCost.Add(fees)
		if computeTotalSend.GreaterThan(balance) {
			continue
		}
		computeTotalSend = transferCost

		actualFees = fees

		break

	}

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if len(usedUTXO) > decoder.wm.Config.MaxTxInputs {
		errStr := fmt.Sprintf("The transaction is use max inputs over: %d", decoder.wm.Config.MaxTxInputs)
		return errors.New(errStr)
	}

	//取账户最后一个地址
	changeAddress := usedUTXO[0].Address

	changeAmount := balance.Sub(computeTotalSend).Sub(actualFees)
	rawTx.FeeRate = feesRate.StringFixed(decoder.wm.Decimal())
	rawTx.Fees = actualFees.StringFixed(decoder.wm.Decimal())

	log.Std.Notice("-----------------------------------------------")
	log.Std.Notice("From Account: %s", accountID)
	log.Std.Notice("To Address: %s", toAddress)
	log.Std.Notice("Amount %s: %v", tokenCoin, toAmount.StringFixed(tokenDecimals))
	log.Std.Notice("Use %s: %v", decoder.wm.Symbol(), balance.StringFixed(decoder.wm.Decimal()))
	log.Std.Notice("Fees %s: %v", decoder.wm.Symbol(), actualFees.StringFixed(decoder.wm.Decimal()))
	log.Std.Notice("Receive %s: %v", decoder.wm.Symbol(), computeTotalSend.StringFixed(decoder.wm.Decimal()))
	log.Std.Notice("Change %s: %v", decoder.wm.Symbol(), changeAmount.StringFixed(decoder.wm.Decimal()))
	log.Std.Notice("Change Address: %v", changeAddress)
	log.Std.Notice("-----------------------------------------------")

	//装配输入
	for _, utxo := range usedUTXO {
		in := omniTransaction.Vin{utxo.TxID, uint32(utxo.Vout)}
		vins = append(vins, in)

		txUnlock := omniTransaction.TxUnlock{LockScript: utxo.ScriptPubKey, SigType: btcTransaction.SigHashAll}
		txUnlocks = append(txUnlocks, txUnlock)
	}

	omniAmount := toAmount.Shift(tokenDecimals)

	omniDetail := omniTransaction.OmniStruct{
		TxType:     omniTransaction.SimpleSend,
		PropertyId: uint32(propertyID),
		Amount:     uint64(omniAmount.IntPart()),
		Ecosystem:  0,
		Address:    "",
		Memo:       "",
	}

	//装配输入
	deamount := computeTotalSend.Shift(decoder.wm.Decimal())
	out := omniTransaction.Vout{toAddress, uint64(deamount.IntPart())}
	vouts = append(vouts, out)

	//changeAmount := balance.Sub(totalSend).Sub(actualFees)
	if changeAmount.GreaterThan(decimal.Zero) {
		deamount := changeAmount.Shift(decoder.wm.Decimal())
		out := omniTransaction.Vout{changeAddress, uint64(deamount.IntPart())}
		vouts = append(vouts, out)

		//fmt.Printf("Create change address for receiving %s coin.", outputs[change])
	}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	/////////构建空交易单
	emptyTrans, err := omniTransaction.CreateEmptyRawTransaction(vins, vouts, omniDetail, lockTime, replaceable)

	if err != nil {
		return fmt.Errorf("create transaction failed, unexpected error: %v", err)
		//log.Error("构建空交易单失败")
	}

	////////构建用于签名的交易单哈希
	transHash, err := omniTransaction.CreateRawTransactionHashForSig(emptyTrans, txUnlocks)
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

	for i, txHash := range transHash {

		var unlockAddr string

		//txHash := transHash[i]

		//判断是否是多重签名
		if txHash.IsMultisig() {
			//获取地址
			//unlockAddr = txHash.GetMultiTxPubkeys() //返回hex数组
		} else {
			//获取地址
			unlockAddr = txHash.GetNormalTxAddress() //返回hex串
		}
		//获取hash值
		beSignHex := txHash.GetTxHashHex()

		log.Std.Debug("txHash[%d]: %s", i, beSignHex)
		//beSignHex := transHash[i]

		addr, err := wrapper.GetAddress(unlockAddr)
		if err != nil {
			return err
		}

		signature := openwallet.KeySignature{
			EccType: decoder.wm.Config.CurveType,
			Nonce:   "",
			Address: addr,
			Message: beSignHex,
		}

		keySigs = append(keySigs, &signature)

	}

	//TODO:多重签名要使用owner的公钥填充

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true

	return nil
}

//SignOmniRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignOmniRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//先加载是否有配置文件
	//err := decoder.wm.LoadConfig()
	//if err != nil {
	//	return err
	//}

	var (
	//txUnlocks   = make([]btcTransaction.TxUnlock, 0)
	//emptyTrans  = rawTx.RawHex
	//transHash   = make([]btcTransaction.TxHash, 0)
	//sigPub      = make([]btcTransaction.SignaturePubkey, 0)
	//privateKeys = make([][]byte, 0)
	)

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
			log.Debug("privateKey:", hex.EncodeToString(keyBytes))

			//privateKeys = append(privateKeys, keyBytes)
			txHash := omniTransaction.TxHash{
				Hash: keySignature.Message,
				Normal: &omniTransaction.NormalTx{
					Address: keySignature.Address.Address,
					SigType: btcTransaction.SigHashAll,
				},
			}
			//transHash = append(transHash, txHash)

			log.Debug("hash:", txHash.GetTxHashHex())

			//签名交易
			/////////交易单哈希签名
			sigPub, err := omniTransaction.SignRawTransactionHash(txHash.GetTxHashHex(), keyBytes)
			if err != nil {
				return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
			} else {

				//for i, s := range sigPub {
				//	log.Info("第", i+1, "个签名结果")
				//	log.Info()
				//	log.Info("对应的公钥为")
				//	log.Info(hex.EncodeToString(s.Pubkey))
				//}

				//txHash.Normal.SigPub = *sigPub
			}

			keySignature.Signature = hex.EncodeToString(sigPub.Signature)
		}
	}

	log.Info("transaction hash sign success")

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	//log.Info("rawTx.Signatures 1:", rawTx.Signatures)

	return nil
}

//VerifyOmniRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyOmniRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//先加载是否有配置文件
	//err := decoder.wm.LoadConfig()
	//if err != nil {
	//	return err
	//}

	var (
		txUnlocks  = make([]omniTransaction.TxUnlock, 0)
		emptyTrans = rawTx.RawHex
		//sigPub     = make([]btcTransaction.SignaturePubkey, 0)
		transHash = make([]omniTransaction.TxHash, 0)
	)

	//TODO:待支持多重签名

	for accountID, keySignatures := range rawTx.Signatures {
		log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			pubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			signaturePubkey := omniTransaction.SignaturePubkey{
				Signature: signature,
				Pubkey:    pubkey,
			}

			//sigPub = append(sigPub, signaturePubkey)

			txHash := omniTransaction.TxHash{
				Hash: keySignature.Message,
				Normal: &omniTransaction.NormalTx{
					Address: keySignature.Address.Address,
					SigType: btcTransaction.SigHashAll,
					SigPub:  signaturePubkey,
				},
			}

			transHash = append(transHash, txHash)

			log.Debug("Signature:", keySignature.Signature)
			log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	txBytes, err := hex.DecodeString(emptyTrans)
	if err != nil {
		return errors.New("Invalid transaction hex data!")
	}

	trx, err := omniTransaction.DecodeRawTransaction(txBytes, decoder.wm.Config.SupportSegWit)
	if err != nil {
		return errors.New("Invalid transaction data! ")
	}

	for _, vin := range trx.Vins {

		utxo, err := decoder.wm.GetTxOut(vin.GetTxID(), uint64(vin.GetVout()))
		if err != nil {
			return err
		}

		txUnlock := omniTransaction.TxUnlock{
			LockScript: utxo.ScriptPubKey,
			SigType:    btcTransaction.SigHashAll}
		txUnlocks = append(txUnlocks, txUnlock)

	}

	//log.Debug(emptyTrans)

	////////填充签名结果到空交易单
	//  传入TxUnlock结构体的原因是： 解锁向脚本支付的UTXO时需要对应地址的赎回脚本， 当前案例的对应字段置为 "" 即可
	signedTrans, err := omniTransaction.InsertSignatureIntoEmptyTransaction(emptyTrans, transHash, txUnlocks)
	if err != nil {
		return fmt.Errorf("transaction compose signatures failed")
	}
	//else {
	//	//	fmt.Println("拼接后的交易单")
	//	//	fmt.Println(signedTrans)
	//	//}

	/////////验证交易单
	//验证时，对于公钥哈希地址，需要将对应的锁定脚本传入TxUnlock结构体
	pass := omniTransaction.VerifyRawTransaction(signedTrans, txUnlocks)
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
