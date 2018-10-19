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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	//"log"
	"math/big"
	"strings"

	"github.com/asdine/storm"

	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/timer"
	"github.com/bndr/gotabulate"
	"github.com/shopspring/decimal"
)

const (
	TRANS_AMOUNT_UNIT_LIST = `
	1: wei
	2: Kwei
	3: Mwei
	4: GWei
	5: microether
	6: milliether
	7: ether
	`
	TRANS_AMOUNT_UNIT_WEI          = 1
	TRANS_AMOUNT_UNIT_K_WEI        = 2
	TRANS_AMOUNT_UNIT_M_WEI        = 3
	TRANS_AMOUNT_UNIT_G_WEI        = 4
	TRANS_AMOUNT_UNIT_MICRO_ETHER  = 5
	TRANS_AMOUNT_UNIT_MILLIE_ETHER = 6
	TRNAS_AMOUNT_UNIT_ETHER        = 7
)

func ConvertEthStringToWei(amount string) (*big.Int, error) {
	log.Debug("amount:", amount)
	vDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		log.Error("convert from string to decimal failed, err=", err)
		return nil, err
	}

	ETH, _ := decimal.NewFromString(strings.Replace("1,000,000,000,000,000,000", ",", "", -1))
	vDecimal = vDecimal.Mul(ETH)
	rst := new(big.Int)
	if _, valid := rst.SetString(vDecimal.String(), 10); !valid {
		log.Error("conver to big.int failed")
		return nil, errors.New("conver to big.int failed")
	}
	return rst, nil
}

func ConverWeiStringToEthDecimal(amount string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		log.Error("convert string to deciaml failed, err=", err)
		return d, err
	}

	ETH, _ := decimal.NewFromString(strings.Replace("1,000,000,000,000,000,000", ",", "", -1))
	d = d.Div(ETH)
	return d, nil
}

func toHexBigIntForEtherTrans(value string, base int, unit int64) (*big.Int, error) {
	amount, err := ConvertToBigInt(value, base)
	if err != nil {
		openwLogger.Log.Errorf("format transaction value failed, err = %v", err)
		return big.NewInt(0), err
	}

	switch unit {
	case TRANS_AMOUNT_UNIT_WEI:
	case TRANS_AMOUNT_UNIT_K_WEI:
		amount.Mul(amount, big.NewInt(1000))
	case TRANS_AMOUNT_UNIT_M_WEI:
		amount.Mul(amount, big.NewInt(1000*1000))
	case TRANS_AMOUNT_UNIT_G_WEI:
		amount.Mul(amount, big.NewInt(1000*1000*1000))
	case TRANS_AMOUNT_UNIT_MICRO_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000))
	case TRANS_AMOUNT_UNIT_MILLIE_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000*1000))
	case TRNAS_AMOUNT_UNIT_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000*1000*1000))
	default:
		return big.NewInt(0), errors.New("wrong unit inputed")
	}

	return amount, nil
}

/*//初始化配置流程
func (this *WalletManager) InitConfigFlow() error {
	file := filepath.Join(this.GetConfig().ConfigFilePath, this.GetConfig().ConfigFileName)
	fmt.Printf("You can run 'vim %s' to edit wallet's config.\n", file)
	return nil
}

//查看配置信息
func (this *WalletManager) ShowConfig() error {
	cfg := this.GetConfig()
	cfgstr, _ := json.MarshalIndent(cfg, "", " ")
	fmt.Printf("-----------------------------------------------------------\n")
	fmt.Println(cfgstr)
	fmt.Printf("-----------------------------------------------------------\n")
	return nil
}*/

//创建钱包流程
func (this *WalletManager) CreateWalletFlow() error {
	//先加载是否有配置文件
	err := this.loadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入钱包名字
	name, err := console.InputText("Enter wallet's name: ", true)

	// 等待用户输入密码
	password, err := console.InputPassword(true, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err = %v", err)
		return err
	}

	_, keyFile, err := this.CreateWallet(name, password)
	if err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("Wallet create successfully, key path: %s\n", keyFile)

	return nil
}

/*
type ERC20Token struct {
	Address  string `json:"address" storm:"id"`
	SymbolID   string `json:"symbol" storm:"index"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	Valid    int    `json:"valid"`
	balance  *big.Int
}
*/

func printTokenAvailable(list []ERC20Token) {
	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		tableInfo = append(tableInfo, []interface{}{
			i, w.Symbol, w.Name, w.Address, w.Name,
		})
	}
	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "SymbolID", "Name", "Contract Address", "Decimals"})

	//打印信息
	fmt.Println(t.Render("simple"))
}

func printTokenWalletList(list []*Wallet) {
	tableInfo := make([][]interface{}, 0)

	for i, w := range list {

		tableInfo = append(tableInfo, []interface{}{
			i, w.WalletID, w.Alias, w.erc20Token.Symbol, w.erc20Token.balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "SymbolID", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))
}

//打印钱包列表
func printWalletList(list []*Wallet, showBalance bool) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		if showBalance {
			balance, _ := ConverWeiStringToEthDecimal(w.balance.String())
			tableInfo = append(tableInfo, []interface{}{
				i, w.WalletID, w.Alias, balance,
			})
		} else {
			tableInfo = append(tableInfo, []interface{}{
				i, w.WalletID, w.Alias,
			})
		}

	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	if showBalance {
		t.SetHeaders([]string{"No.", "ID", "Name", "Balance"})
	} else {
		t.SetHeaders([]string{"No.", "ID", "Name"})
	}

	//打印信息
	fmt.Println(t.Render("simple"))

}

//创建地址流程
func (this *WalletManager) CreateAddressFlow() error {
	//先加载是否有配置文件
	err := this.loadConfig()
	if err != nil {
		return err
	}

	//查询所有钱包信息
	wallets, err := this.GetLocalWalletList(this.GetConfig().KeyDir, this.GetConfig().DbPath, false)
	if err != nil {
		fmt.Printf("The node did not create any wallet!\n")
		return err
	}

	//打印钱包
	printWalletList(wallets, false)

	fmt.Printf("[Please select a wallet account to create address] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(wallets) {
		return errors.New("Input number is out of index! ")
	}

	account := wallets[num]

	// 输入地址数量
	count, err := console.InputNumber("Enter the number of addresses you want: ", false)
	if err != nil {
		return err
	}

	if count > maxAddresNum {
		return errors.New(fmt.Sprintf("The number of addresses can not exceed %d\n", maxAddresNum))
	}

	//输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err = %v", err)
		return err
	}

	err = this.UnlockWallet(account, password)
	if err != nil {
		openwLogger.Log.Errorf("unlock wallet [%v] failed, err = %v", account.WalletID, err)
		return err
	}

	log.Info("Start batch creation ")
	log.Info("-------------------------------------------------")

	filepath, err := this.CreateBatchAddress2(account.WalletID, password, count)
	if err != nil {
		return err
	}

	log.Info("-------------------------------------------------")
	log.Info("all ", count, " addresses have created, file path:", filepath)

	return nil
}

func (this *WalletManager) ERC20TokenSummaryFollow() error {
	endRunning := make(chan bool, 1)
	//先加载是否有配置文件
	err := this.loadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(this.GetConfig().SumAddress) == 0 {
		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", this.SymbolID))
	}

	ercTokens, err := this.GetERC20TokenList()
	if err != nil {
		openwLogger.Log.Errorf("find tokens failed, err = %v", err)
		return err
	}

	if len(ercTokens) == 0 {
		openwLogger.Log.Errorf("no token available, config the tokens first.")
		return err
	}

	printTokenAvailable(ercTokens)

	//选择token
	tokenId, err := console.InputNumber("Enter Token No. : ", true)
	if err != nil {
		return err
	}

	if int(tokenId) >= len(ercTokens) {
		return errors.New("Input Token No. is out of index! ")
	}

	token := ercTokens[tokenId]
	fmt.Println("token[", token.Symbol, "] is chosen. ")

	wallets, err := this.ERC20GetWalletList(&token)
	if err != nil {
		return err
	}

	//打印钱包列表
	printTokenWalletList(wallets)

	fmt.Printf("[Please select the wallet to summary, and enter the numbers split by ','." +
		" For example: 0,1,2,3] \n")

	// 等待用户输入钱包名字
	nums, err := console.InputText("Enter the No. group: ", true)
	if err != nil {
		return err
	}

	//分隔数组
	array := strings.Split(nums, ",")

	for _, numIput := range array {
		if common.IsNumberString(numIput) {
			numInt := common.NewString(numIput).Int()
			if numInt < len(wallets) {
				w := wallets[numInt]
				fmt.Printf("Register summary wallet [%s]-[%s]\n", w.Alias, w.WalletID)

				password, err := console.InputPassword(false, 8)
				if err != nil {
					openwLogger.Log.Errorf("input wallet password failed, err=%v", err)
					return err
				}

				err = this.UnlockWallet(w, password)
				if err != nil {
					openwLogger.Log.Errorf("unlock wallet [%v] failed, err = %v", w.WalletID, err)
					return err
				}
				w.Password = password
				this.AddWalletInSummary(w.WalletID, w)
			} else {
				return errors.New("The input No. out of index! ")
			}
		} else {
			return errors.New("The input No. is not numeric! ")
		}
	}

	if len(this.WalletInSumOld) == 0 {
		return errors.New("Not summary wallets to register! ")
	}

	fmt.Printf("The timer for summary has started. Execute by every %v seconds.\n", this.GetConfig().CycleSeconds)

	//启动钱包汇总程序
	//sumTimer := timer.NewTask(cycleSeconds, ERC20SummaryWallets)
	//sumTimer.Start()
	go this.ERC20SummaryWallets()

	<-endRunning
	return nil
}

//汇总钱包流程
func (this *WalletManager) SummaryFollow() error {
	endRunning := make(chan bool, 1)
	err := this.loadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(this.GetConfig().SumAddress) == 0 {
		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", this.SymbolID))
	}

	wallets, err := this.GetLocalWalletList(this.GetConfig().KeyDir, this.GetConfig().DbPath, true)
	if err != nil {
		return err
	}

	//打印钱包列表
	printWalletList(wallets, true)

	fmt.Printf("[Please select the wallet to summary, and enter the numbers split by ','." +
		" For example: 0,1,2,3] \n")

	// 等待用户输入钱包名字
	nums, err := console.InputText("Enter the No. group: ", true)
	if err != nil {
		return err
	}

	//分隔数组
	array := strings.Split(nums, ",")

	for _, numIput := range array {
		if common.IsNumberString(numIput) {
			numInt := common.NewString(numIput).Int()
			if numInt < len(wallets) {
				w := wallets[numInt]
				fmt.Printf("Register summary wallet [%s]-[%s]\n", w.Alias, w.WalletID)

				password, err := console.InputPassword(false, 8)
				if err != nil {
					openwLogger.Log.Errorf("input wallet password failed, err=%v", err)
					return err
				}

				err = this.UnlockWallet(w, password)
				if err != nil {
					openwLogger.Log.Errorf("unlock wallet [%v] failed, err = %v", w.WalletID, err)
					return err
				}
				w.Password = password
				this.AddWalletInSummary(w.WalletID, w)
			} else {
				return errors.New("The input No. out of index! ")
			}

		} else {
			return errors.New("The input No. is not numeric! ")
		}
	}

	if len(this.WalletInSumOld) == 0 {
		return errors.New("Not summary wallets to register! ")
	}

	fmt.Printf("The timer for summary has started. Execute by every %v seconds.\n", this.GetConfig().CycleSeconds)

	//启动钱包汇总程序
	sumTimer := timer.NewTask(time.Second*time.Duration(this.GetConfig().CycleSeconds), this.SummaryWallets)
	sumTimer.Start()
	//go SummaryWallets()

	<-endRunning
	return nil
}

//查看钱包列表，显示信息
func (this *WalletManager) GetWalletList() error {
	err := this.loadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(this.GetConfig().SumAddress) == 0 {
		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", this.SymbolID))
	}

	wallets, err := this.GetLocalWalletList(this.GetConfig().KeyDir, this.GetConfig().DbPath, true)
	if err != nil {
		return err
	}

	//打印钱包列表
	printWalletList(wallets, true)
	return nil
}

func (this *WalletManager) ConfigERC20Token() error {
	//先加载是否有配置文件
	err := this.loadConfig()
	if err != nil {
		return err
	}

	ercTokens, err := this.GetERC20TokenList()
	if err != nil {
		openwLogger.Log.Errorf("find tokens failed, err = %v", err)
		return err
	}

	if len(ercTokens) == 0 {
		openwLogger.Log.Errorf("no token available, config the tokens first.")
	}

	printTokenAvailable(ercTokens)

	tokenName, err := console.InputText("Enter Token Name. : ", true)
	if err != nil {
		return err
	}

	tokenSymbol, err := console.InputText("Enter Token SymbolID. : ", true)
	if err != nil {
		return err
	}

	tokenAddress, err := console.InputText("Enter Token Address. : ", true)
	if err != nil {
		return err
	}

	tokenDecimal, err := console.InputNumber("Enter Token Decimals. :", false)
	if err != nil {
		return err
	}

	tosave, err := console.InputText("Save Token Config [Y/N]. :", true)
	if err != nil {
		return err
	}
	tosave = strings.ToLower(tosave)
	if tosave != "y" {
		fmt.Println("give up token config. ")
		return nil
	}

	tokenConfig := &ERC20Token{
		Name:     tokenName,
		Symbol:   tokenSymbol,
		Address:  tokenAddress,
		Decimals: int(tokenDecimal),
	}

	err = this.SaveERC20TokenConfig(tokenConfig)
	if err != nil {
		openwLogger.Log.Errorf("save token config failed, err = %v", err)
		return err
	}

	ercTokens, err = this.GetERC20TokenList()
	if err != nil {
		openwLogger.Log.Errorf("find tokens failed, err = %v", err)
		return err
	}

	if len(ercTokens) == 0 {
		openwLogger.Log.Errorf("no token available, config the tokens first.")
		return err
	}

	printTokenAvailable(ercTokens)

	return nil
}

func (this *WalletManager) ERC20TokenTransferFlow() error {
	//先加载是否有配置文件
	err := this.loadConfig()
	if err != nil {
		return err
	}

	ercTokens, err := this.GetERC20TokenList()
	if err != nil {
		openwLogger.Log.Errorf("find tokens failed, err = %v", err)
		return err
	}

	if len(ercTokens) == 0 {
		openwLogger.Log.Errorf("no token available, config the tokens first.")
		return err
	}

	printTokenAvailable(ercTokens)

	//选择钱包
	tokenId, err := console.InputNumber("Enter Token No. : ", true)
	if err != nil {
		return err
	}

	if int(tokenId) >= len(ercTokens) {
		return errors.New("Input Token No. is out of index! ")
	}

	token := ercTokens[tokenId]
	fmt.Println("token[", token.Symbol, "] is chosen. ")

	list, err := this.ERC20GetWalletList(&token)
	if err != nil {
		return err
	}

	printTokenWalletList(list)

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(list) {
		return errors.New("Input number is out of index! ")
	}

	wallet := list[num]
	fmt.Println("wallet[", wallet.Alias, "] is chosen. ")

	// 等待用户输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err=%v", err)
		return err
	}

	// 等待用户输入发送数量
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	value, err := console.InputRealNumber("Enter amount to send : ", true)
	if err != nil {
		return err
	}

	amount, err := ConvertToBigInt(value, 10)
	if err != nil {
		openwLogger.Log.Errorf("convert to big int failed, err = %v", err)
		return err
	}

	fmt.Println("amount input:", amount.String())

	if wallet.erc20Token.balance.Cmp(amount) < 0 {
		return errors.New("Input amount is greater than balance! ")
	}

	//建立交易单
	txID, err := this.ERC20SendTransaction2(wallet,
		receiver, amount, password, true)
	if err != nil {
		return err
	}

	fmt.Printf("Send transaction successfully, TXID：%s\n", txID)

	return nil
}

//发送交易
func (this *WalletManager) TransferFlow() error {
	//先加载是否有配置文件
	err := this.loadConfig()
	if err != nil {
		return err
	}

	list, err := this.GetLocalWalletList(this.GetConfig().KeyDir, this.GetConfig().DbPath, true)
	if err != nil {
		return err
	}

	//打印钱包列表
	printWalletList(list, true)

	fmt.Printf("[Please select a wallet to send transaction] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(list) {
		return errors.New("Input number is out of index! ")
	}

	wallet := list[num]
	fmt.Println("wallet[", wallet.Alias, "] is chosen. ")

	// 等待用户输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err=%v", err)
		return err
	}

	// 等待用户输入发送数量
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	fmt.Println("receiver: ", receiver)

	//fmt.Println("Choose the unit for the transaction:")
	//fmt.Println(TRANS_AMOUNT_UNIT_LIST)
	//unit, err := console.InputNumber("Index of the unit: ", true)
	//if err != nil {
	//	return err
	//}

	amount, err := console.InputRealNumber("Enter amount to send : ", true)
	if err != nil {
		return err
	}

	amountInt, err := ConvertEthStringToWei(amount) //toHexBigIntForEtherTrans(amount, 10, int64(unit))
	if err != nil {
		openwLogger.Log.Errorf("wrong amount inputed. ")
		return err
	}

	amountDecimal, err := ConverWeiStringToEthDecimal(amountInt.String())
	if err != nil {
		return err
	}

	fmt.Println("amount input:", amountDecimal)

	if wallet.balance.Cmp(amountInt) < 0 {
		return errors.New("Input amount is greater than balance! ")
	}

	//建立交易单
	txID, err := this.SendTransaction2(wallet,
		receiver, amountInt, password, true)
	if err != nil {
		return err
	}

	fmt.Printf("Send transaction successfully, TXID：%s\n", txID)

	return nil
}

//备份钱包流程
func (this *WalletManager) BackupWalletFlow() error {
	//先加载是否有配置文件
	err := this.loadConfig()
	if err != nil {
		return err
	}

	wallets, err := this.GetLocalWalletList(this.GetConfig().KeyDir, this.GetConfig().DbPath, true)
	if err != nil {
		openwLogger.Log.Errorf("get wallet list failed, err = ", err)
		return err
	}

	//打印钱包列表
	printWalletList(wallets, true)

	fmt.Printf("[Please select a wallet to backup] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(wallets) {
		return errors.New("Input number is out of index! ")
	}

	wallet := wallets[num]

	// 等待用户输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err = %v", err)
		return err
	}

	backupPath, err := this.BackupWalletToDefaultPath(wallet, password)
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Wallet backup file path: %s\n", backupPath)

	return nil
}

func (this *WalletManager) GetLocalBlockHeight() (uint64, error) {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for get local block height failed, err=%v", err)
		return 0, err
	}
	defer db.Close()
	var blockHeight uint64
	err = db.Get(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &blockHeight)
	if err != nil {
		openwLogger.Log.Errorf("get block height from db failed, err=%v", err)
		return 0, err
	}
	// blockHeight, err := ConvertToUint64(blockHeightStr, 16) //ConvertToBigInt(blockHeightStr, 16)
	// if err != nil {
	// 	openwLogger.Log.Errorf("convert block height string failed, err=%v", err)
	// 	return 0, err
	// }
	return blockHeight, nil
}

func (this *WalletManager) SaveLocalBlockScanned(blockHeight uint64, blockHash string) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for update local block height failed, err=%v", err)
		return err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		openwLogger.Log.Errorf("start transaction for save block scanned failed, err=%v", err)
		return err
	}
	defer tx.Rollback()

	//blockHeightStr := "0x" + strconv.FormatUint(blockHeight, 16) //blockHeight.Text(16)
	err = tx.Set(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &blockHeight)
	if err != nil {
		openwLogger.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	err = tx.Set(BLOCK_CHAIN_BUCKET, BLOCK_HASH_KEY, &blockHash)
	if err != nil {
		openwLogger.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	tx.Commit()
	return nil
}

func (this *WalletManager) UpdateLocalBlockHeight(blockHeight uint64) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for update local block height failed, err=%v", err)
		return err
	}
	defer db.Close()

	//blockHeightStr := "0x" + strconv.FormatUint(blockHeight, 16) //blockHeight.Text(16)
	err = db.Set(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &blockHeight)
	if err != nil {
		openwLogger.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	return nil
}

func (this *WalletManager) RecoverBlockHeader(height uint64) (*EthBlock, error) {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for save block failed, err=%v", err)
		return nil, err
	}
	defer db.Close()
	var block EthBlock

	err = db.One("BlockNumber", "0x"+strconv.FormatUint(height, 16), &block.BlockHeader)
	if err != nil {
		openwLogger.Log.Errorf("get block failed, block number=%v, err=%v", "0x"+strconv.FormatUint(height, 16), err)
		return nil, err
	}

	block.blockHeight, err = ConvertToUint64(block.BlockNumber, 16) //ConvertToBigInt(block.BlockNumber, 16)
	if err != nil {
		openwLogger.Log.Errorf("conver block height to big int failed, err= %v", err)
		return nil, err
	}
	return &block, nil
}

func (this *WalletManager) SaveBlockHeader(block *EthBlock) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()
	err = db.Save(&block.BlockHeader)
	if err != nil {
		openwLogger.Log.Errorf("save block failed, err = %v", err)
		return err
	}
	return nil
}

func (this *WalletManager) SaveBlockHeader2(block *EthBlock) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		openwLogger.Log.Errorf("start transaction for save block header failed, err=%v", err)
		return err
	}
	defer tx.Rollback()

	err = tx.Save(&block.BlockHeader)
	if err != nil {
		openwLogger.Log.Errorf("save block failed, err = %v", err)
		return err
	}

	//blockHeightStr := "0x" + strconv.FormatUint(block.blockHeight, 16) //block.blockHeight.Text(16)
	err = tx.Set(BLOCK_CHAIN_BUCKET, BLOCK_HEIGHT_KEY, &block.blockHeight)
	if err != nil {
		openwLogger.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	err = tx.Set(BLOCK_CHAIN_BUCKET, BLOCK_HASH_KEY, &block.BlockHash)
	if err != nil {
		openwLogger.Log.Errorf("update block height failed, err= %v", err)
		return err
	}

	tx.Commit()
	return nil
}

/*func (this *WalletManager) SaveTransaction(tx *BlockTransaction) error {
	db, err := OpenDB(DbPath, BLOCK_CHAIN_DB)
	if err != nil {
		openwLogger.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()

	err = db.Save(tx)
	if err != nil {
		openwLogger.Log.Errorf("save block transaction failed, err = %v", err)
		return err
	}
	return nil
}*/

func (this *WalletManager) RecoverUnscannedTransactions(unscannedTxs []UnscanTransaction) ([]BlockTransaction, error) {
	allTxs := make([]BlockTransaction, 0, len(unscannedTxs))
	for i, _ := range unscannedTxs {
		var tx BlockTransaction
		err := json.Unmarshal([]byte(unscannedTxs[i].TxSpec), &tx)
		if err != nil {
			openwLogger.Log.Errorf("decode json [%v] from unscanned transactions failed, err=%v", unscannedTxs[i].TxSpec, err)
			return nil, err
		}
		allTxs = append(allTxs, tx)
	}
	return allTxs, nil
}

func (this *WalletManager) GetAllUnscannedTransactions() ([]UnscanTransaction, error) {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for save block failed, err=%v", err)
		return nil, err
	}
	defer db.Close()

	var allRecords []UnscanTransaction
	err = db.All(&allRecords)
	if err != nil {
		openwLogger.Log.Errorf("get all unscanned transactions failed, err = %v", err)
		return nil, err
	}

	return allRecords, nil
}

func (this *WalletManager) DeleteUnscannedTransactions(list []UnscanTransaction) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		log.Errorf("start transaction failed, err=%v", err)
		return err
	}
	defer tx.Rollback()

	for i, _ := range list {
		err = tx.DeleteStruct(&list[i])
		if err != nil {
			log.Errorf("delete unscanned tx faled, err= %v", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error("commit failed, err=%v", err)
		return err
	}
	return nil
}

func (this *WalletManager) DeleteUnscannedTransactionByHeight(height uint64) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()

	var list []UnscanTransaction
	heightStr := "0x" + strconv.FormatUint(height, 16)
	err = db.Find("BlockNumber", heightStr, &list)
	if err != nil && err != storm.ErrNotFound {
		openwLogger.Log.Errorf("find unscanned tx failed, block height=%v, err=%v", heightStr, err)
		return err
	} else if err == storm.ErrNotFound {
		openwLogger.Log.Infof("no unscanned tx found in block [%v]", heightStr)
		return nil
	}

	for _, r := range list {
		err = db.DeleteStruct(&r)
		if err != nil {
			openwLogger.Log.Errorf("delete unscanned tx faled, block height=%v, err=%v", heightStr, err)
			return err
		}
	}
	return nil
}

func (this *WalletManager) SaveUnscannedTransaction(tx *BlockTransaction, reason string) error {
	db, err := OpenDB(this.GetConfig().DbPath, this.GetConfig().BlockchainFile)
	if err != nil {
		openwLogger.Log.Errorf("open db for save block failed, err=%v", err)
		return err
	}
	defer db.Close()

	txSpec, _ := json.Marshal(tx)

	unscannedRecord := &UnscanTransaction{
		TxID:        tx.Hash,
		BlockNumber: tx.BlockNumber,
		BlockHash:   tx.BlockHash,
		TxSpec:      string(txSpec),
		Reason:      reason,
	}
	err = db.Save(unscannedRecord)
	if err != nil {
		openwLogger.Log.Errorf("save unscanned record failed, err=%v", err)
		return err
	}
	return nil
}

//恢复钱包
func (this *WalletManager) RestoreWalletFlow() error {
	//先加载是否有配置文件
	err := this.loadConfig()
	if err != nil {
		return err
	}

	//输入恢复文件路径
	keyPath, err := console.InputText("Enter backup key file path: ", true)
	if err != nil {
		return err
	}

	// 等待用户输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err = %v", err)
		return err
	}

	fmt.Printf("Wallet restoring, please wait a moment...\n")
	err = this.RestoreWallet2(keyPath, password)
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Restore wallet successfully.\n")

	return nil
}
