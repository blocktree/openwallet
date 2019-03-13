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

package qtum

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blocktree/openwallet/assets/qtum/btcLikeTxDriver"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
	"time"
)

var (
	DEFAULT_GAS_LIMIT = "250000"
	DEFAULT_GAS_PRICE = decimal.New(4, -7)
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
		return decoder.CreateQRC20RawTransaction(wrapper, rawTx)
	} else {
		return decoder.CreateSimpleRawTransaction(wrapper, rawTx)
	}
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	if sumRawTx.Coin.IsContract {
		return decoder.CreateQRC20SummaryRawTransaction(wrapper, sumRawTx)
	} else {
		return decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	}
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		usedUTXO         []*Unspent
		outputAddrs      = make(map[string]string)
		balance          = decimal.New(0, 0)
		totalSend        = decimal.New(0, 0)
		actualFees       = decimal.New(0, 0)
		feesRate         = decimal.New(0, 0)
		accountID        = rawTx.Account.AccountID
		destinations     = make([]string, 0)
		accountTotalSent = decimal.Zero
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
	//decoder.wm.Log.Debug(searchAddrs)
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
		//计算账户的实际转账amount
		addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", addr)
		if findErr != nil || len(addresses) == 0 {
			amountDec, _ := decimal.NewFromString(amount)
			accountTotalSent = accountTotalSent.Add(amountDec)
		}
	}

	//获取utxo，按小到大排序
	sort.Sort(UnspentSort{unspents, func(a, b *Unspent) int {
		a_amount, _ := decimal.NewFromString(a.Amount)
		b_amount, _ := decimal.NewFromString(b.Amount)
		if a_amount.GreaterThan(b_amount) {
			return 1
		} else {
			return -1
		}
	}})

	if len(rawTx.FeeRate) == 0 {
		feesRate, err = decoder.wm.EstimateFeeRate()
		if err != nil {
			return err
		}
	} else {
		feesRate, _ = decimal.NewFromString(rawTx.FeeRate)
	}

	decoder.wm.Log.Info("Calculating wallet unspent record to build transaction...")
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
	if len(usedUTXO) > decoder.wm.config.maxTxInputs {
		errStr := fmt.Sprintf("The transaction is use max inputs over: %d", decoder.wm.config.maxTxInputs)
		return errors.New(errStr)
	}

	//取账户最后一个地址
	changeAddress := usedUTXO[0].Address

	changeAmount := balance.Sub(computeTotalSend).Sub(actualFees)
	rawTx.FeeRate = feesRate.StringFixed(decoder.wm.Decimal())
	rawTx.Fees = actualFees.StringFixed(decoder.wm.Decimal())

	decoder.wm.Log.Std.Notice("-----------------------------------------------")
	decoder.wm.Log.Std.Notice("From Account: %s", accountID)
	decoder.wm.Log.Std.Notice("To Address: %s", strings.Join(destinations, ", "))
	decoder.wm.Log.Std.Notice("Use: %v", balance.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Fees: %v", actualFees.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Receive: %v", computeTotalSend.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Change: %v", changeAmount.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Change Address: %v", changeAddress)
	decoder.wm.Log.Std.Notice("-----------------------------------------------")

	//装配输出
	for to, amount := range rawTx.To {
		outputAddrs[to] = amount
	}

	//changeAmount := balance.Sub(totalSend).Sub(actualFees)
	if changeAmount.GreaterThan(decimal.New(0, 0)) {
		outputAddrs[changeAddress] = changeAmount.StringFixed(decoder.wm.Decimal())
	}

	err = decoder.createSimpleRawTransaction(wrapper, rawTx, usedUTXO, outputAddrs)
	if err != nil {
		return err
	}

	return nil
}

//CreateSimpleSummaryRawTransaction 创建主币汇总交易
func (decoder *TransactionDecoder) CreateSimpleSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {

	var (
		feesRate           = decimal.New(0, 0)
		accountID          = sumRawTx.Account.AccountID
		minTransfer, _     = decimal.NewFromString(sumRawTx.MinTransfer)
		retainedBalance, _ = decimal.NewFromString(sumRawTx.RetainedBalance)
		sumAddresses       = make([]string, 0)
		rawTxArray         = make([]*openwallet.RawTransaction, 0)
		sumUnspents        []*Unspent
		outputAddrs        map[string]string
		totalInputAmount   decimal.Decimal
	)

	if minTransfer.LessThan(retainedBalance) {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	address, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit, "AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(address) == 0 {
		return nil, fmt.Errorf("[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range address {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := decoder.wm.GetBlockScanner().GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return nil, err
	}

	for _, addrBalance := range addrBalanceArray {
		decoder.wm.Log.Debugf("addrBalance: %+v", addrBalance)
		//检查余额是否超过最低转账
		addrBalance_dec, _ := decimal.NewFromString(addrBalance.Balance)
		if addrBalance_dec.GreaterThanOrEqual(minTransfer) {
			//添加到转账地址数组
			sumAddresses = append(sumAddresses, addrBalance.Address)
		}
	}

	if len(sumAddresses) == 0 {
		return nil, fmt.Errorf("all address balance is less than mini transfer")
	}

	//取得费率
	if len(sumRawTx.FeeRate) == 0 {
		feesRate, err = decoder.wm.EstimateFeeRate()
		if err != nil {
			return nil, err
		}
	} else {
		feesRate, _ = decimal.NewFromString(sumRawTx.FeeRate)
	}

	sumUnspents = make([]*Unspent, 0)
	outputAddrs = make(map[string]string, 0)
	totalInputAmount = decimal.Zero

	for i, addr := range sumAddresses {

		unspents, err := decoder.wm.ListUnspent(sumRawTx.Confirms, addr)
		if err != nil {
			return nil, err
		}

		//尽可能筹够最大input数
		if len(unspents)+len(sumUnspents) < decoder.wm.config.maxTxInputs {

			for _, u := range unspents {
				if u.Spendable {
					sumUnspents = append(sumUnspents, u)
					if retainedBalance.GreaterThan(decimal.Zero) {
						outputAddrs[addr] = retainedBalance.StringFixed(decoder.wm.Decimal())
					}
				}
			}

			//decoder.wm.Log.Debugf("sumUnspents: %+v", sumUnspents)
		}

		//如果utxo已经超过最大输入，或遍历地址完结，就可以进行构建交易单
		if i == len(sumAddresses)-1 || len(sumUnspents) >= decoder.wm.config.maxTxInputs {
			//执行构建交易单工作
			//decoder.wm.Log.Debugf("sumUnspents: %+v", sumUnspents)
			//计算手续费，构建交易单inputs，地址保留余额>0，地址需要加入输出，最后+1是汇总地址
			fees, createErr := decoder.wm.EstimateFee(int64(len(sumUnspents)), int64(len(outputAddrs)+1), feesRate)
			if createErr != nil {
				return nil, createErr
			}

			//计算这笔交易单的汇总数量
			for _, u := range sumUnspents {

				ua, _ := decimal.NewFromString(u.Amount)
				totalInputAmount = totalInputAmount.Add(ua)
			}

			/*

					汇总数量计算：

					1. 输入总数量 = 合计账户地址的所有utxo
					2. 账户地址输出总数量 = 账户地址保留余额 * 地址数
				    3. 汇总数量 = 输入总数量 - 账户地址输出总数量 - 手续费
			*/
			retainedBalanceTotal := retainedBalance.Mul(decimal.New(int64(len(outputAddrs)), 0))
			sumAmount := totalInputAmount.Sub(retainedBalanceTotal).Sub(fees)

			decoder.wm.Log.Debugf("totalInputAmount: %v", totalInputAmount)
			decoder.wm.Log.Debugf("retainedBalanceTotal: %v", retainedBalanceTotal)
			decoder.wm.Log.Debugf("fees: %v", fees)
			decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

			//最后填充汇总地址及汇总数量
			outputAddrs[sumRawTx.SummaryAddress] = sumAmount.StringFixed(decoder.wm.Decimal())

			//创建一笔交易单
			rawTx := &openwallet.RawTransaction{
				Coin:     sumRawTx.Coin,
				Account:  sumRawTx.Account,
				FeeRate:  sumRawTx.FeeRate,
				To:       outputAddrs,
				Fees:     fees.StringFixed(decoder.wm.Decimal()),
				Required: 1,
			}

			createErr = decoder.createSimpleRawTransaction(wrapper, rawTx, sumUnspents, outputAddrs)
			if createErr != nil {
				return nil, createErr
			}

			//创建成功，添加到队列
			rawTxArray = append(rawTxArray, rawTx)

			//清空临时变量
			sumUnspents = make([]*Unspent, 0)
			outputAddrs = make(map[string]string, 0)
			totalInputAmount = decimal.Zero

		}
	}

	return rawTxArray, nil
}

//CreateQRC20RawTransaction 创建QRC20交易单
func (decoder *TransactionDecoder) CreateQRC20RawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		outputAddrs      = make(map[string]string)
		tokenOutputAddrs = make(map[string]string)

		toAddress string
		toAmount  = decimal.Zero

		availableUTXO     = make([]*Unspent, 0)
		usedUTXO          = make([]*Unspent, 0)
		totalTokenBalance = decimal.Zero
		useTokenBalance   = decimal.Zero
		useTokenAddress   = ""
		missUtxoAddress   = ""
		missToken         = make([]string, 0)
		balance           = decimal.New(0, 0)
		actualFees        = decimal.New(0, 0)
		feesRate          = decimal.New(0, 0)
		accountID         = rawTx.Account.AccountID

		//accountTotalSent = decimal.Zero
		//txFrom           = make([]string, 0)
		//txTo             = make([]string, 0)
	)

	if len(rawTx.Coin.Contract.Address) == 0 {
		return fmt.Errorf("contract address is empty")
	}

	tokenCoin := rawTx.Coin.Contract.Token
	tokenDecimals := int32(rawTx.Coin.Contract.Decimals)

	//TODO: 应该通过配置文件设置
	//合约手续费在普通交易基础上加0.1个qtum, 小数点8位
	actualFees, _ = decimal.NewFromString("0.1")

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

	//token限制只能发送目标一个
	if len(rawTx.To) > 1 {
		return fmt.Errorf("token transfer not support multiple receiver address")
	}

	//选择一个输出地址
	for to, amount := range rawTx.To {
		toAddress = to
		toAmount, _ = decimal.NewFromString(amount)

		//计算账户的实际转账amount
		//addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", to)
		//if findErr != nil || len(addresses) == 0 {
		//	accountTotalSent = accountTotalSent.Add(toAmount)
		//}
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
		tokenBalance, checkErr := decoder.wm.GetQRC20Balance(rawTx.Coin.Contract, address.Address, decoder.wm.config.isTestNet)
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
				//decoder.wm.Log.Debug("missUtxoAddress:", missUtxoAddress)
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
		return fmt.Errorf("account[%s] token[%s] total balance: %s is not enough! ", accountID, tokenCoin, totalTokenBalance.StringFixed(tokenDecimals))
	}

	//单个地址的可用Token余额不足够
	if useTokenBalance.LessThan(toAmount) {
		return fmt.Errorf("account[%s] token[%s] total balance is enough, but the available balance: %s of address[%s] is not enough! ", accountID, tokenCoin, useTokenBalance.StringFixed(tokenDecimals), useTokenAddress)
	} else {
		//可用Token余额足够，但没有utxo，返回错误及有Token余额的地址没有可用主链币
		if len(missUtxoAddress) > 0 {
			return fmt.Errorf("account[%s] token[%s] total balance is enough, but the utxo of address[%s] is empty! ", accountID, tokenCoin, missUtxoAddress)
		}
	}

	//选择一个地址作为发送
	//txFrom = []string{fmt.Sprintf("%s:%s", availableUTXO[0].Address, toAmount.StringFixed(tokenDecimals))}

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

	decoder.wm.Log.Info("Calculating wallet unspent record to build transaction...")

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
				if balance.GreaterThanOrEqual(actualFees) {
					break
				}
			}
		}

		if balance.LessThan(actualFees) {
			return fmt.Errorf("The [%s] available utxo balance: %s is not enough! ", decoder.wm.Symbol(), balance.StringFixed(decoder.wm.Decimal()))
		}

		//计算手续费，输出地址有2个，一个是发送，一个是找零，一个是OP_CALL
		fees, err := decoder.wm.EstimateFee(int64(len(usedUTXO)), int64(2), feesRate)
		if err != nil {
			return err
		}

		//如果要手续费有发送支付，得计算加入手续费后，计算余额是否足够
		//总共要发送的
		actualFees = actualFees.Add(fees)
		if actualFees.GreaterThan(balance) {
			continue
		}

		break

	}

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if len(usedUTXO) > decoder.wm.config.maxTxInputs {
		errStr := fmt.Sprintf("The transaction is use max inputs over: %d", decoder.wm.config.maxTxInputs)
		return errors.New(errStr)
	}

	//取账户最后一个地址
	changeAddress := usedUTXO[0].Address

	changeAmount := balance.Sub(actualFees)
	rawTx.FeeRate = feesRate.StringFixed(decoder.wm.Decimal())
	rawTx.Fees = actualFees.StringFixed(decoder.wm.Decimal())

	decoder.wm.Log.Std.Notice("-----------------------------------------------")
	decoder.wm.Log.Std.Notice("From Account: %s", accountID)
	decoder.wm.Log.Std.Notice("To Address: %s", toAddress)
	decoder.wm.Log.Std.Notice("Amount %s: %v", tokenCoin, toAmount.StringFixed(tokenDecimals))
	decoder.wm.Log.Std.Notice("Use %s: %v", decoder.wm.Symbol(), balance.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Fees %s: %v", decoder.wm.Symbol(), actualFees.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Change %s: %v", decoder.wm.Symbol(), changeAmount.StringFixed(decoder.wm.Decimal()))
	decoder.wm.Log.Std.Notice("Change Address: %v", changeAddress)
	decoder.wm.Log.Std.Notice("-----------------------------------------------")

	//changeAmount := balance.Sub(totalSend).Sub(actualFees)
	if changeAmount.GreaterThan(decimal.Zero) {
		outputAddrs[changeAddress] = changeAmount.StringFixed(decoder.wm.Decimal())
	}

	tokenOutputAddrs[toAddress] = toAmount.StringFixed(tokenDecimals)

	err = decoder.createQRC2ORawTransaction(wrapper, rawTx, usedUTXO, outputAddrs, tokenOutputAddrs)
	if err != nil {
		return err
	}

	return nil
}

//CreateQRC20SummaryRawTransaction 创建QRC20汇总交易
func (decoder *TransactionDecoder) CreateQRC20SummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {

	var (
		feesRate           = decimal.New(0, 0)
		accountID          = sumRawTx.Account.AccountID
		minTransfer, _     = decimal.NewFromString(sumRawTx.MinTransfer)
		retainedBalance, _ = decimal.NewFromString(sumRawTx.RetainedBalance)
		rawTxArray         = make([]*openwallet.RawTransaction, 0)
		sumUnspents        []*Unspent
		outputAddrs        map[string]string
		tokenOutputAddrs   map[string]string
	)

	if len(sumRawTx.Coin.Contract.Address) == 0 {
		return nil, fmt.Errorf("contract address is empty")
	}

	tokenDecimals := int32(sumRawTx.Coin.Contract.Decimals)
	//TODO: 应该通过配置文件设置
	//合约手续费在普通交易基础上加0.1个qtum, 小数点8位
	transferCost, _ := decimal.NewFromString("0.1")
	coinDecimals := decoder.wm.Decimal()

	if minTransfer.LessThan(retainedBalance) {
		return nil, fmt.Errorf("mini transfer amount must be greater than address retained balance")
	}

	address, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit, "AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(address) == 0 {
		return nil, fmt.Errorf("[%s] have not addresses", accountID)
	}

	//取得费率
	if len(sumRawTx.FeeRate) == 0 {
		feesRate, err = decoder.wm.EstimateFeeRate()
		if err != nil {
			return nil, err
		}
	} else {
		feesRate, _ = decimal.NewFromString(sumRawTx.FeeRate)
	}

	/*

		1. 遍历账户所有地址。
		2. 查询地址的token余额。
		3. 查询地址是否有utxo。（地址要有足够的主币做手续费）
		4. 保留余额，检查手续费是否足够
		5. 构建token交易单。
		6. 把原始交易单加入到数组。

	*/

	for _, address := range address {

		sumUnspents = make([]*Unspent, 0)

		//清空临时变量
		outputAddrs = make(map[string]string, 0)
		tokenOutputAddrs = make(map[string]string, 0)
		//decoder.wm.Log.Debug("address.Address:", address.Address)
		//查找地址token余额
		tokenBalance, createErr := decoder.wm.GetQRC20Balance(sumRawTx.Coin.Contract, address.Address, decoder.wm.config.isTestNet)
		if createErr != nil {
			continue
		}

		//decoder.wm.Log.Debug("tokenBalance:", tokenBalance)
		//查询地址的utxo
		unspents, createErr := decoder.wm.ListUnspent(sumRawTx.Confirms, address.Address)
		if createErr != nil {
			continue
		}
		if tokenBalance.LessThan(minTransfer) || len(unspents) == 0 {
			continue
		}

		//合计地址主币余额
		addrBalance := decimal.Zero
		for _, u := range unspents {

			if u.Spendable {
				ua, _ := decimal.NewFromString(u.Amount)
				addrBalance = addrBalance.Add(ua)
				sumUnspents = append(sumUnspents, u)
			}

		}
		//decoder.wm.Log.Debug("addrBalance:", addrBalance)
		//计算手续费，构建交易单inputs，输出2个，1个为目标地址，1个为OP_CALL
		fees, createErr := decoder.wm.EstimateFee(int64(len(sumUnspents)), 2, feesRate)
		if createErr != nil {
			return nil, createErr
		}

		//地址的主币余额要必须大于最低转账成本+手续费
		if addrBalance.LessThan(transferCost.Add(fees)) {
			continue
		}

		//主币输出第一个为汇总地址及最低转账成本
		//outputAddrs[sumRawTx.SummaryAddress] = transferCost.StringFixed(coinDecimals)

		//计算找零 = 地址余额 - 手续费 - 汇总地址的最低转账成本
		changeAmount := addrBalance.Sub(fees).Sub(transferCost)
		if changeAmount.GreaterThan(decimal.Zero) {
			//主币输出第二个地址为找零地址，找零主币
			outputAddrs[address.Address] = changeAmount.StringFixed(coinDecimals)
		}

		//计算汇总数量
		sumTokenAmount := tokenBalance.Sub(retainedBalance)
		//token输出汇总地址及汇总数量
		tokenOutputAddrs[sumRawTx.SummaryAddress] = sumTokenAmount.StringFixed(tokenDecimals)

		decoder.wm.Log.Debugf("tokenBalance: %v", tokenBalance)
		decoder.wm.Log.Debugf("addressBalance: %v", addrBalance)
		decoder.wm.Log.Debugf("fees: %v", fees)
		decoder.wm.Log.Debugf("changeAmount: %v", changeAmount)
		decoder.wm.Log.Debugf("sumTokenAmount: %v", sumTokenAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:     sumRawTx.Coin,
			Account:  sumRawTx.Account,
			FeeRate:  sumRawTx.FeeRate,
			To:       tokenOutputAddrs,
			Fees:     fees.StringFixed(decoder.wm.Decimal()),
			Required: 1,
		}

		createErr = decoder.createQRC2ORawTransaction(wrapper, rawTx, sumUnspents, outputAddrs, tokenOutputAddrs)
		if createErr != nil {
			return nil, createErr
		}

		//创建成功，添加到队列
		rawTxArray = append(rawTxArray, rawTx)

	}

	return rawTxArray, nil
}

//CreateQRC20RawTransaction 创建QRC20交易单
func (decoder *TransactionDecoder) CreateRawTransaction__deprecated(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

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
		totalSend         = decimal.New(0, 0)
		actualFees        = decimal.New(0, 0)
		feesRate          = decimal.New(0, 0)
		accountID         = rawTx.Account.AccountID
		destinations      = make([]string, 0)
		DEFAULT_GAS_LIMIT = "250000"
		DEFAULT_GAS_PRICE = decimal.New(4, -7)

		coinDecimal      = int32(0)
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
	decoder.wm.Log.Debug(searchAddrs)

	//decoder.wm.Log.Info("Calculating wallet unspent record to build transaction...")
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
			to           string
			amount       string
			txInTotal    uint64
			deAmount     decimal.Decimal
			sendAmount   decimal.Decimal
			totalQtum    decimal.Decimal
			unspent      decimal.Decimal
			change       uint64
			contractFees decimal.Decimal
			gasPrice     string
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

		decoder.wm.Log.Debugf("feesRate:%v", feesRate)

		usedTokenUTXO := make([]*Unspent, 0)
		unspent = decimal.New(0, 0)

		computeTotalfee := feesRate
		//decoder.wm.Log.Debugf("computeTotalfee:%v",computeTotalfee)
		decoder.wm.Log.Info("Calculating wallet unspent record to build transaction...")
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
						decoder.wm.Log.Errorf("GetTokenUnspentByAddress failed unexpected error: %v\n", err)
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

				if u.Spendable && u.TxID != usedTokenUTXO[0].TxID {

					if totalQtum.GreaterThanOrEqual(computeTotalfee) {
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

			//计算手续费，找零地址有2个，一个是发送，一个是OP_CALL
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
			decoder.wm.Log.Debugf("computeTotalfee:%v, totalQtum:%v", computeTotalfee, totalQtum)
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
		} else {
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

		if len(vouts) != 1 {
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
			//decoder.wm.Log.Error("构建空交易单失败")
		}

	} else {

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

		decoder.wm.Log.Info("Calculating wallet unspent record to build transaction...")
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

			//decoder.wm.Log.Debugf("fees:%v",fees)
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

		decoder.wm.Log.Std.Notice("-----------------------------------------------")
		decoder.wm.Log.Std.Notice("From Account: %s", accountID)
		decoder.wm.Log.Std.Notice("To Address: %s", strings.Join(destinations, ", "))
		decoder.wm.Log.Std.Notice("Use: %v", balance.StringFixed(8))
		decoder.wm.Log.Std.Notice("Fees: %v", actualFees.StringFixed(8))
		decoder.wm.Log.Std.Notice("Receive: %v", computeTotalSend.StringFixed(8))
		decoder.wm.Log.Std.Notice("Change: %v", changeAmount.StringFixed(8))
		decoder.wm.Log.Std.Notice("Change Address: %v", changeAddress)
		decoder.wm.Log.Std.Notice("-----------------------------------------------")

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
			//decoder.wm.Log.Error("构建空交易单失败")
		}

		accountTotalSent = accountTotalSent.Add(actualFees)
	}

	////////构建用于签名的交易单哈希
	transHash, err := btcLikeTxDriver.CreateRawTransactionHashForSig(emptyTrans, txUnlocks)
	if err != nil {
		return fmt.Errorf("create transaction hash for sig failed, unexpected error: %v", err)
		//decoder.wm.Log.Error("获取待签名交易单哈希失败")
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
		txUnlocks   = make([]btcLikeTxDriver.TxUnlock, 0)
		emptyTrans  = rawTx.RawHex
		transHash   = make([]string, 0)
		sigPub      = make([]btcLikeTxDriver.SignaturePubkey, 0)
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

	//decoder.wm.Log.Debug("transHash len:", len(transHash))
	//decoder.wm.Log.Debug("txUnlocks len:", len(txUnlocks))

	/////////交易单哈希签名
	sigPub, err = btcLikeTxDriver.SignRawTransactionHash(transHash, txUnlocks)
	if err != nil {
		return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
	} else {
		decoder.wm.Log.Info("transaction hash sign success")
		//for i, s := range sigPub {
		//	decoder.wm.Log.Info("第", i+1, "个签名结果")
		//	decoder.wm.Log.Info()
		//	decoder.wm.Log.Info("对应的公钥为")
		//	decoder.wm.Log.Info(hex.EncodeToString(s.Pubkey))
		//}
	}

	if len(sigPub) != len(keySignatures) {
		return fmt.Errorf("sign raw transaction fail, program error. ")
	}

	for i, keySignature := range keySignatures {
		keySignature.Signature = hex.EncodeToString(sigPub[i].Signature)
		//decoder.wm.Log.Debug("keySignature.Signature:",i, "=", keySignature.Signature)
	}

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	//decoder.wm.Log.Info("rawTx.Signatures 1:", rawTx.Signatures)

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
		decoder.wm.Log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			pubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			signaturePubkey := btcLikeTxDriver.SignaturePubkey{
				Signature: signature,
				Pubkey:    pubkey,
			}

			sigPub = append(sigPub, signaturePubkey)

			decoder.wm.Log.Debug("Signature:", keySignature.Signature)
			decoder.wm.Log.Debug("PublicKey:", keySignature.Address.PublicKey)
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

	//decoder.wm.Log.Debug(emptyTrans)

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
		decoder.wm.Log.Debug("transaction verify passed")
		rawTx.IsCompleted = true
		rawTx.RawHex = signedTrans
	} else {
		decoder.wm.Log.Debug("transaction verify failed")
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

//createSimpleRawTransaction 创建原始交易单
func (decoder *TransactionDecoder) createSimpleRawTransaction(
	wrapper openwallet.WalletDAI,
	rawTx *openwallet.RawTransaction,
	usedUTXO []*Unspent,
	to map[string]string,
) error {

	var (
		err              error
		vins             = make([]btcLikeTxDriver.Vin, 0)
		vouts            = make([]btcLikeTxDriver.Vout, 0)
		txUnlocks        = make([]btcLikeTxDriver.TxUnlock, 0)
		totalSend        = decimal.New(0, 0)
		destinations     = make([]string, 0)
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		accountID        = rawTx.Account.AccountID
	)

	if len(usedUTXO) == 0 {
		return fmt.Errorf("utxo is empty")
	}

	if len(to) == 0 {
		return fmt.Errorf("Receiver addresses is empty! ")
	}

	//计算总发送金额
	for addr, amount := range to {
		deamount, _ := decimal.NewFromString(amount)
		totalSend = totalSend.Add(deamount)
		destinations = append(destinations, addr)
		//计算账户的实际转账amount
		addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", accountID, "Address", addr)
		if findErr != nil || len(addresses) == 0 {
			amountDec, _ := decimal.NewFromString(amount)
			accountTotalSent = accountTotalSent.Add(amountDec)
		}
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

		txFrom = append(txFrom, fmt.Sprintf("%s:%s", utxo.Address, utxo.Amount))
	}

	//装配输入
	for to, amount := range to {
		deamount, _ := decimal.NewFromString(amount)
		deamount = deamount.Shift(decoder.wm.Decimal())
		out := btcLikeTxDriver.Vout{to, uint64(deamount.IntPart())}
		vouts = append(vouts, out)

		txTo = append(txTo, fmt.Sprintf("%s:%s", to, amount))
	}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	/////////构建空交易单
	emptyTrans, err := btcLikeTxDriver.CreateEmptyRawTransaction(vins, vouts, lockTime, replaceable, decoder.wm.config.isTestNet)

	if err != nil {
		return fmt.Errorf("create transaction failed, unexpected error: %v", err)
		//decoder.wm.Log.Error("构建空交易单失败")
	}

	////////构建用于签名的交易单哈希
	transHash, err := btcLikeTxDriver.CreateRawTransactionHashForSig(emptyTrans, txUnlocks)
	if err != nil {
		return fmt.Errorf("create transaction hash for sig failed, unexpected error: %v", err)
		//decoder.wm.Log.Error("获取待签名交易单哈希失败")
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

	feesDec, _ := decimal.NewFromString(rawTx.Fees)
	accountTotalSent = accountTotalSent.Add(feesDec)
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.StringFixed(decoder.wm.Decimal())
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}

//createQRC20RawTransaction 创建QRC20原始交易单
func (decoder *TransactionDecoder) createQRC2ORawTransaction(
	wrapper openwallet.WalletDAI,
	rawTx *openwallet.RawTransaction,
	usedUTXO []*Unspent,
	coinTo map[string]string,
	tokenTo map[string]string,
) error {

	var (
		err              error
		to               string
		vins             = make([]btcLikeTxDriver.Vin, 0)
		vouts            = make([]btcLikeTxDriver.Vout, 0)
		txUnlocks        = make([]btcLikeTxDriver.TxUnlock, 0)
		accountTotalSent = decimal.Zero
		toAmount         = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		accountID        = rawTx.Account.AccountID
	)

	if len(usedUTXO) == 0 {
		return fmt.Errorf("utxo is empty")
	}

	if len(coinTo) == 0 {
		return fmt.Errorf("Receiver addresses is empty! ")
	}

	if len(tokenTo) == 0 {
		return fmt.Errorf("Receiver addresses is empty! ")
	}

	//合约地址
	contractAddr := strings.TrimPrefix(rawTx.Coin.Contract.Address, "0x")
	tokenDecimals := int32(rawTx.Coin.Contract.Decimals)

	//记录输入输出明细
	for addr, amount := range tokenTo {
		//选择utxo的第一个地址作为发送放
		txFrom = []string{fmt.Sprintf("%s:%s", usedUTXO[0].Address, amount)}
		//接收方的地址和数量
		txTo = []string{fmt.Sprintf("%s:%s", addr, amount)}
		to = addr
		toAmount, _ = decimal.NewFromString(amount)
		//计算账户的实际转账amount
		addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", accountID, "Address", addr)
		if findErr != nil || len(addresses) == 0 {
			accountTotalSent = accountTotalSent.Add(toAmount)
		}
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

		//txFrom = append(txFrom, fmt.Sprintf("%s:%s", utxo.Address, utxo.Amount))
	}

	//装配输入
	for outAddr, amount := range coinTo {
		deamount, _ := decimal.NewFromString(amount)
		deamount = deamount.Mul(decimal.New(1, tokenDecimals))
		out := btcLikeTxDriver.Vout{outAddr, uint64(deamount.IntPart())}
		vouts = append(vouts, out)

		//txTo = append(txTo, fmt.Sprintf("%s:%s", to, amount))
	}

	if len(vouts) != 1 {
		return fmt.Errorf("the number of change addresses must be equal to one. ")
	}

	SotashiGasPriceDec := DEFAULT_GAS_PRICE.Shift(decoder.wm.Decimal())
	gasPrice := SotashiGasPriceDec.String()
	sendAmount := toAmount.Shift(tokenDecimals)

	//装配合约
	vcontract := btcLikeTxDriver.Vcontract{contractAddr, to, sendAmount, DEFAULT_GAS_LIMIT, gasPrice, 0}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	/////////构建空交易单
	emptyTrans, err := btcLikeTxDriver.CreateQRC20TokenEmptyRawTransaction(vins, vcontract, vouts, lockTime, replaceable, decoder.wm.config.isTestNet)

	if err != nil {
		return fmt.Errorf("create transaction failed, unexpected error: %v", err)
		//decoder.wm.Log.Error("构建空交易单失败")
	}

	////////构建用于签名的交易单哈希
	transHash, err := btcLikeTxDriver.CreateRawTransactionHashForSig(emptyTrans, txUnlocks)
	if err != nil {
		return fmt.Errorf("create transaction hash for sig failed, unexpected error: %v", err)
		//decoder.wm.Log.Error("获取待签名交易单哈希失败")
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

	//feesDec, _ := decimal.NewFromString(rawTx.Fees)
	//accountTotalSent = accountTotalSent.Add(feesDec)
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.StringFixed(tokenDecimals)
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}
