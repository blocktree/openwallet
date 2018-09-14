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
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/btcsuite/btcutil/hdkeychain"
	ethKStore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	maxAddresNum = 10000
)

const (
	ERC20TOKEN_DB = "erc20Token.db"
)

const (
	WALLET_NOT_EXIST_ERR        = "The wallet whose name is given not exist!"
	BACKUP_FILE_TYPE_ADDRESS    = 0
	BACKUP_FILE_TYPE_WALLET_KEY = 1
	BACKUP_FILE_TYPE_WALLET_DB  = 2
)

var (
	// 节点客户端
	client *Client
	//秘钥存取
	storage *keystore.HDKeystore
)

func init() {
	storage = keystore.NewHDKeystore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	client = &Client{BaseURL: serverAPI, Debug: true}
}

//CreateNewWallet 创建钱包
func CreateNewWallet(name, password string) (*Wallet, string, error) {

	//检查钱包名是否存在
	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, "", errors.New(fmt.Sprintf("get wallet keys failed, err = %v", err))
	}

	for _, w := range wallets {
		if w.Alias == name {
			return nil, "", errors.New("The wallet's alias is duplicated!")
		}
	}

	//fmt.Printf("Verify password in bitcoin-core wallet...\n")
	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		return nil, "", err
	}

	extSeed, err := keystore.GetExtendSeed(seed, MasterKey)
	if err != nil {
		return nil, "", err
	}

	key, keyFile, err := keystore.StoreHDKeyWithSeed(keyDir, name, password, extSeed, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, "", err
	}

	w := Wallet{WalletID: key.RootId, Alias: key.Alias}

	return &w, keyFile, nil
}

func GetWalletKey(fileWitoutProfix string) (*Wallet, error) {

	keyfile := fileWitoutProfix + ".key"
	//dbfile := fileWitoutProfix + ".db"
	finfo, err := os.Stat(keyfile)
	if err != nil {
		openwLogger.Log.Errorf("stat file [%v] failed, err = %v", keyfile, err)
		return nil, err
	}

	/*if strings.Index(finfo.Name(), ".key") != (len(finfo.Name()) - 5) {
		openwLogger.Log.Errorf("file name error")
		return nil, errors.New("verify key file name error")
	}*/
	var key struct {
		Alias  string `json:"alias"`
		RootId string `json:"rootid"`
	}
	buf := new(bufio.Reader)

	fd, err := os.Open(keyfile)
	defer fd.Close()
	if err != nil {
		openwLogger.Log.Errorf("get wallet key, open db failed, err = %v", err)
		return nil, err
	}

	buf.Reset(fd)
	// Parse the address.
	key.Alias = ""
	key.RootId = ""
	err = json.NewDecoder(buf).Decode(&key)
	if err != nil {
		openwLogger.Log.Errorf("decode key file error, err = %v", err)
		return nil, err
	}

	return &Wallet{WalletID: key.RootId, Alias: key.Alias, KeyFile: finfo.Name()}, nil
}

func GetWalletKeys(dir string) ([]*Wallet, error) {
	var key struct {
		Alias  string `json:"alias"`
		RootId string `json:"rootid"`
	}
	buf := new(bufio.Reader)
	wallets := make([]*Wallet, 0)

	//加载文件，实例化钱包
	readWallet := func(path string) *Wallet {

		fd, err := os.Open(path)
		defer fd.Close()
		if err != nil {
			return nil
		}

		buf.Reset(fd)
		// Parse the address.
		key.Alias = ""
		key.RootId = ""
		err = json.NewDecoder(buf).Decode(&key)
		if err != nil {
			return nil
		}

		return &Wallet{WalletID: key.RootId, Alias: key.Alias}
	}

	//扫描key目录的所有钱包
	absPath, _ := filepath.Abs(dir)
	file.MkdirAll(absPath)
	files, err := ioutil.ReadDir(absPath)
	if err != nil {
		return wallets, err
	}

	for _, fi := range files {
		// Skip any non-key files from the folder
		if skipKeyFile(fi) {
			continue
		}
		if fi.IsDir() {
			continue
		}

		path := filepath.Join(keyDir, fi.Name())

		w := readWallet(path)
		w.KeyFile = fi.Name()
		fmt.Println("absolute path:", absPath)
		wallets = append(wallets, w)

	}

	return wallets, nil
}

func skipKeyFile(fi os.FileInfo) bool {
	// Skip editor backups and UNIX-style hidden files.
	if strings.HasSuffix(fi.Name(), "~") || strings.HasPrefix(fi.Name(), ".") {
		return true
	}
	// Skip misc special files, directories (yes, symlinks too).
	if fi.IsDir() || fi.Mode()&os.ModeType != 0 {
		return true
	}

	return false
}

func SaveERC20TokenConfig(config *ERC20Token) error {
	db, err := OpenDB(dbPath, ERC20TOKEN_DB)
	defer db.Close()
	if err != nil {
		openwLogger.Log.Errorf("open db for path [%v] failed, err = %v", dbPath+"/"+ERC20TOKEN_DB, err)
		return err
	}
	err = db.Save(config)
	if err != nil {
		openwLogger.Log.Errorf("save db for path [%v] failed, err = %v", dbPath+"/"+ERC20TOKEN_DB, err)
		return err
	}
	return nil
}

func GetERC20TokenList() ([]ERC20Token, error) {
	db, err := OpenDB(dbPath, ERC20TOKEN_DB)
	defer db.Close()
	if err != nil {
		openwLogger.Log.Errorf("open db for path [%v] failed, err = %v", dbPath+"/"+ERC20TOKEN_DB, err)
		return nil, err
	}
	tokens := make([]ERC20Token, 0)
	err = db.All(&tokens)
	if err != nil {
		openwLogger.Log.Errorf("query token list in db failed, err= %v", err)
		return nil, err
	}
	return tokens, nil
}

func ERC20GetWalletList(erc20Token *ERC20Token) ([]*Wallet, error) {
	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	for i, _ := range wallets {
		wallets[i].erc20Token = &ERC20Token{}
		*wallets[i].erc20Token = *erc20Token
		tokenBanlance, err := ERC20GetWalletBalance(wallets[i])
		if err != nil {

			openwLogger.Log.Errorf(fmt.Sprintf("find wallet balance failed, err=%v\n", err))
			return nil, err
		}

		wallets[i].erc20Token.balance = tokenBanlance
	}
	return wallets, nil
}

//GetWalletList 获取钱包列表
func GetWalletList() ([]*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for i, _ := range wallets {
		//fmt.Println("loop to wallet balance")
		balance, err := GetWalletBalance(wallets[i])
		if err != nil {

			openwLogger.Log.Errorf(fmt.Sprintf("find wallet balance failed, err=%v\n", err))
			return nil, err
		}
		wallets[i].balance = balance
	}

	return wallets, nil
}

//AddWalletInSummary 添加汇总钱包账户
func AddWalletInSummary(wid string, wallet *Wallet) {
	walletsInSum[wid] = wallet
}

//阻塞式的执行外部shell命令的函数,等待执行完毕并返回标准输出
func exec_shell(s string) (string, error) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("/bin/bash", "-c", s)

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()

	return out.String(), err
}

func BackupWalletToDefaultPath(wallet *Wallet, password string) (string, error) {
	newBackupDir := filepath.Join(backupDir, wallet.FileName()+"-"+common.TimeFormat(TIME_POSTFIX))
	return BackupWallet(newBackupDir, wallet, password)
}

func BackupWallet(newBackupDir string, wallet *Wallet, password string) (string, error) {
	/*w, err := GetWalletInfo(wallet.WalletID)
	if err != nil {
		return "", err
	}*/

	err := UnlockWallet(wallet, password)
	if err != nil {
		openwLogger.Log.Errorf("unlock wallet failed, err=%v", err)
		return "", err
	}

	addressMap := make(map[string]int)
	files := make([]string, 0)

	//创建备份文件夹
	//newBackupDir := filepath.Join(backupDir, w.FileName()+"-"+common.TimeFormat("20060102150405"))
	file.MkdirAll(newBackupDir)

	addrs, err := GetAddressesByWallet(wallet)
	if err != nil {
		openwLogger.Log.Errorf("get addresses by wallet failed, err = %v", err)
		return "", err
	}

	//搜索出绑定钱包的地址
	for _, addr := range addrs {
		address := addr.Address
		address = strings.Trim(address, " ")
		address = strings.ToLower(address)
		addressMap[address] = 1
	}

	/*for k, v := range addressMap {
		fmt.Println("address[", k, "], exist[", v, "]")
	}*/

	rd, err := ioutil.ReadDir(EthereumKeyPath)
	if err != nil {
		openwLogger.Log.Errorf("open ethereum key path [%v] failed, err=%v", EthereumKeyPath, err)
		return "", err
	}

	//fmt.Println("rd length:", len(rd))
	for _, fi := range rd {
		if skipKeyFile(fi) {
			continue
		}

		//fmt.Println("file name:", fi.Name())
		parts := strings.Split(fi.Name(), "--")
		l := len(parts)
		if l == 0 {
			continue
		}

		theAddr := "0x" + parts[l-1]
		//fmt.Println("loop addr:", theAddr)
		if _, exist := addressMap[theAddr]; exist {
			files = append(files, fi.Name())
		} /*else {
			fmt.Println("address[", theAddr, "], exist[", addressMap[theAddr], "]")
		}*/
	}

	/*for _, keyfile := range files {
		cmd := "cp " + EthereumKeyPath + "/" + keyfile + " " + newBackupDir
		_, err = exec_shell(cmd)
		if err != nil {
			openwLogger.Log.Errorf("backup key faile failed, err = ", err)
			return "", err
		}
	}*/

	//fmt.Println("file list length:", len(files))

	//备份该钱包下的所有地址
	for _, keyfile := range files {
		err := file.Copy(EthereumKeyPath+"/"+keyfile, newBackupDir+"/")
		if err != nil {
			openwLogger.Log.Errorf("backup key faile failed, err = %v", err)
			return "", err
		}
	}

	//备份钱包key文件
	file.Copy(filepath.Join(keyDir, wallet.FileName()+".key"), newBackupDir)

	//备份地址数据库
	file.Copy(filepath.Join(dbPath, wallet.FileName()+".db"), newBackupDir)

	return newBackupDir, nil
}

func GetWalletInfo(walletID string) (*Wallet, error) {
	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			balance, err := GetWalletBalance(w)
			if err != nil {
				return nil, err
			}
			w.balance = balance
			return w, nil
		}

	}

	return nil, errors.New(WALLET_NOT_EXIST_ERR)
}

func CreateNewPrivateKey(parentKey *keystore.HDKey, timestamp, index uint64) (*ethKStore.Key, *Address, error) {

	derivedPath := fmt.Sprintf("%s/%d/%d", parentKey.RootPath, timestamp, index)
	//fmt.Printf("derivedPath = %s\n", derivedPath)
	childKey, err := parentKey.DerivedKeyWithPath(derivedPath)

	privateKey, err := childKey.ECPrivKey()
	if err != nil {
		return nil, nil, err
	}

	key := ecdsa.PrivateKey(*privateKey)

	/*cfg := chaincfg.MainNetParams
	if isTestNet {
		cfg = chaincfg.TestNet3Params
	}

	wif, err := btcutil.NewWIF(privateKey, &cfg, true)
	if err != nil {
		return "", nil, err
	}

	address, err := childKey.Address(&cfg)
	if err != nil {
		return "", nil, err
	}*/

	keyCombo := ethKStore.NewKeyFromECDSA(&key)

	addr := Address{
		Address: keyCombo.Address.String(), //address.String(),
		Account: parentKey.RootId,
		HDPath:  derivedPath,
		//	Balance:   "0",
		CreatedAt: time.Now(),
	}

	return keyCombo, &addr, err
}

func verifyBackupWallet(wallet *Wallet, keyPath string, password string) error {
	s := keystore.NewHDKeystore(keyPath, keystore.StandardScryptN, keystore.StandardScryptP)
	_, err := wallet.HDKey(password, s)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get HDkey from path[%v], err=%v\n", keyPath, err))
		return err
	}
	return nil
}

func UnlockWallet(wallet *Wallet, password string) error {
	_, err := wallet.HDKey(password, storage)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get HDkey, err=%v\n", err))
		return err
	}
	return nil
}

func CreateBatchAddress(name, password string, count uint64) error {
	//读取钱包
	w, err := GetWalletInfo(name)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get wallet info, err=%v\n", err))
		return err
	}

	//验证钱包
	keyroot, err := w.HDKey(password, storage)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get HDkey, err=%v\n", err))
		return err
	}

	timestamp := uint64(time.Now().Unix())

	db, err := w.OpenDB()
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("open db, err=%v\n", err))
		return err
	}
	defer db.Close()

	ethKeyStore := ethKStore.NewKeyStore(EthereumKeyPath, ethKStore.StandardScryptN, ethKStore.StandardScryptP)

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	errcount := uint64(0)
	for i := uint64(0); i < count && errcount < count; {
		// 生成地址
		keyCombo, address, err := CreateNewPrivateKey(keyroot, timestamp, i)
		if err != nil {
			log.Printf("Create new privKey failed unexpected error: %v\n", err)
			errcount++
			continue
		}
		_, err = ethKeyStore.NewAccountForWalletBT(keyCombo, password)
		if err != nil {
			openwLogger.Log.Errorf("NewAccountForWalletBT failed, err = %v", err)
			errcount++
			continue
		}
		//ethKeyStore.StoreNewKeyForWalletBT(ethKeyStore, keyCombo, DefaultPasswordForEthKey)

		err = tx.Save(address)
		if err != nil {
			log.Printf("save address for wallet failed, err=%v\n", err)
			errcount++
			continue
		}
		i++
	}

	return tx.Commit()
}

type AddrVec struct {
	addrs []*Address
}

func (this *AddrVec) Len() int {
	return len(this.addrs)
}

func (this *AddrVec) Swap(i, j int) {
	this.addrs[i], this.addrs[j] = this.addrs[j], this.addrs[i]
}

func (this *AddrVec) Less(i, j int) bool {
	if this.addrs[i].balance.Cmp(this.addrs[j].balance) < 0 {
		return true
	}
	return false
}

type TokenAddrVec struct {
	addrs []*Address
}

func (this *TokenAddrVec) Len() int {
	return len(this.addrs)
}

func (this *TokenAddrVec) Swap(i, j int) {
	this.addrs[i], this.addrs[j] = this.addrs[j], this.addrs[i]
}

func (this *TokenAddrVec) Less(i, j int) bool {
	if this.addrs[i].tokenBalance.Cmp(this.addrs[j].tokenBalance) < 0 {
		return true
	}
	return false
}

type txFeeInfo struct {
	GasLimit *big.Int
	GasPrice *big.Int
	Fee      *big.Int
}

func GetTransactionFeeEstimated(from string, to string, value *big.Int, data string) (*txFeeInfo, error) {
	gasLimit, err := ethGetGasEstimated(makeGasEstimatePara(from, to, value, data))
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas limit failed, err = %v\n", err))
		return nil, err
	}

	gasPrice, err := ethGetGasPrice()
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas price failed, err = %v\n", err))
		return nil, err
	}

	fee := new(big.Int)
	fee.Mul(gasLimit, gasPrice)

	feeInfo := &txFeeInfo{
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Fee:      fee,
	}
	return feeInfo, nil
}

func GetERC20TokenTransactionFeeEstimated(from string, to string, data string) (*txFeeInfo, error) {
	/*gasLimit, err := ethGetGasEstimated(makeERC20TokenTransGasEstimatePara(from, to, data))
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas limit failed, err = %v\n", err))
		return nil, err
	}

	gasPrice, err := ethGetGasPrice()
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas price failed, err = %v\n", err))
		return nil, err
	}

	fee := new(big.Int)
	fee.Mul(gasLimit, gasPrice)

	feeInfo := &txFeeInfo{
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Fee:      fee,
	}
	return feeInfo, nil*/
	return GetTransactionFeeEstimated(from, to, nil, data)
}

func GetSimpleTransactionFeeEstimated(from string, to string, amount *big.Int) (*txFeeInfo, error) {
	/*gasLimit, err := ethGetGasEstimated(makeSimpleTransGasEstimatedPara(from, to, amount))
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas limit failed, err = %v\n", err))
		return nil, err
	}

	gasPrice, err := ethGetGasPrice()
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas price failed, err = %v\n", err))
		return nil, err
	}

	fee := new(big.Int)
	fee.Mul(gasLimit, gasPrice)

	feeInfo := &txFeeInfo{
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Fee:      fee,
	}
	return feeInfo, nil*/
	return GetTransactionFeeEstimated(from, to, amount, "")
}

func ERC20SendTransaction(wallet *Wallet, to string, amount *big.Int, password string, feesInSender bool) ([]string, error) {
	var txIds []string

	err := UnlockWallet(wallet, password)
	if err != nil {
		openwLogger.Log.Errorf("unlock wallet [%v]. failed, err=%v", wallet.WalletID, err)
		return nil, err
	}

	addrs, err := ERC20GetAddressesByWallet(wallet)
	if err != nil {
		openwLogger.Log.Errorf("failed to get addresses from db, err = %v", err)
		return nil, err
	}

	sort.Sort(&TokenAddrVec{addrs: addrs})
	//检查下地址排序是否正确, 仅用于测试
	for _, theAddr := range addrs {
		fmt.Println("theAddr[", theAddr.Address, "]:", theAddr.tokenBalance)
	}

	for i := len(addrs) - 1; i >= 0 && amount.Cmp(big.NewInt(0)) > 0; i-- {
		var fee *txFeeInfo
		var amountToSend big.Int
		fmt.Println("amount remained:", amount.String())
		//空的token账户直接跳过
		//if addrs[i].tokenBalance.Cmp(big.NewInt(0)) == 0 {
		//	openwLogger.Log.Infof("skip the address[%v] with 0 balance. ", addrs[i].Address)
		//	continue
		//}

		if addrs[i].tokenBalance.Cmp(amount) >= 0 {
			amountToSend = *amount

		} else {
			amountToSend = *addrs[i].tokenBalance
		}

		dataPara, err := makeERC20TokenTransData(wallet.erc20Token.Address, to, &amountToSend)
		if err != nil {
			openwLogger.Log.Errorf("make token transaction data failed, err=%v", err)
			return nil, err
		}
		fee, err = GetERC20TokenTransactionFeeEstimated(addrs[i].Address, wallet.erc20Token.Address, dataPara)
		if err != nil {
			openwLogger.Log.Errorf("get erc token transaction fee estimated failed, err = %v", err)
			continue
		}

		if addrs[i].balance.Cmp(fee.Fee) < 0 {
			openwLogger.Log.Errorf("address[%v] cannot afford a token transfer with a fee [%v]", addrs[i].Address, fee.Fee)
			continue
		}

		txid, err := SendTransactionToAddr(makeERC20TokenTransactionPara(addrs[i], wallet.erc20Token.Address, dataPara, password, fee))
		if err != nil {
			openwLogger.Log.Errorf("SendTransactionToAddr failed, err=%v", err)
			if txid == "" {
				continue //txIds = append(txIds, txid)
			}
		}

		txIds = append(txIds, txid)
		amount.Sub(amount, &amountToSend)
	}

	return txIds, nil
}

func SendTransaction(wallet *Wallet, to string, amount *big.Int, password string, feesInSender bool) ([]string, error) {
	var txIds []string

	err := UnlockWallet(wallet, password)
	if err != nil {
		openwLogger.Log.Errorf("unlock wallet [%v]. failed, err=%v", wallet.WalletID, err)
		return nil, err
	}

	addrs, err := GetAddressesByWallet(wallet)
	if err != nil {
		openwLogger.Log.Errorf("failed to get addresses from db, err = %v", err)
		return nil, err
	}

	sort.Sort(&AddrVec{addrs: addrs})
	//检查下地址排序是否正确, 仅用于测试
	for _, theAddr := range addrs {
		fmt.Println("theAddr[", theAddr.Address, "]:", theAddr.balance)
	}
	//amountLeft := *amount
	for i := len(addrs) - 1; i >= 0 && amount.Cmp(big.NewInt(0)) > 0; i-- {
		var amountToSend big.Int
		var fee *txFeeInfo

		fmt.Println("amount remained:", amount.String())
		//空账户直接跳过
		//if addrs[i].balance.Cmp(big.NewInt(0)) == 0 {
		//	openwLogger.Log.Infof("skip the address[%v] with 0 balance. ", addrs[i].Address)
		//	continue
		//}

		//如果该地址的余额足够支付转账
		if addrs[i].balance.Cmp(amount) >= 0 {
			amountToSend = *amount
			fee, err = GetSimpleTransactionFeeEstimated(addrs[i].Address, to, &amountToSend)
			if err != nil {
				openwLogger.Log.Errorf("%v", err)
				continue
			}

			balanceLeft := *addrs[i].balance
			balanceLeft.Sub(&balanceLeft, fee.Fee)

			//灰尘账户, 余额不足以发起一次transaction
			//fmt.Println("amount to send ignore fee:", amountToSend.String())
			if balanceLeft.Cmp(big.NewInt(0)) < 0 {
				errinfo := fmt.Sprintf("[%v] is a dust address, will skip. ", addrs[i].Address)
				openwLogger.Log.Errorf(errinfo)
				continue
			}

			//如果改地址的余额除去手续费后, 不足以支付转账, set 转账金额 = 账户余额 - 手续费
			if balanceLeft.Cmp(&amountToSend) < 0 {
				amountToSend = balanceLeft
				//fmt.Println("amount to send plus fee:", amountToSend.String())
			}

		} else {
			amountToSend = *addrs[i].balance
			fee, err = GetSimpleTransactionFeeEstimated(addrs[i].Address, to, &amountToSend)
			if err != nil {
				openwLogger.Log.Errorf("%v", err)
				continue
			}

			//灰尘账户, 余额不足以发起一次transaction
			if amountToSend.Cmp(fee.Fee) <= 0 {
				errinfo := fmt.Sprintf("[%v] is a dust address, will skip. ", addrs[i].Address)
				openwLogger.Log.Errorf(errinfo)
				continue
			}

			//fmt.Println("amount to send without fee, ", amountToSend.String(), " , fee:", fee.Fee.String())
			amountToSend.Sub(&amountToSend, fee.Fee)
			//fmt.Println("amount to send applied fee, ", amountToSend.String())
		}

		txid, err := SendTransactionToAddr(makeSimpleTransactionPara(addrs[i], to, &amountToSend, password, fee))
		if err != nil {
			openwLogger.Log.Errorf("SendTransactionToAddr failed, err=%v", err)
			if txid == "" {
				continue //txIds = append(txIds, txid)
			}
		}

		txIds = append(txIds, txid)
		amount.Sub(amount, &amountToSend)
	}

	return txIds, nil
}

func removeOxFromHex(value string) string {
	var result string
	if strings.Index(value, "0x") != -1 {
		result = common.Substr(value, 2, len(value))
	}
	return result
}

func ConvertToBigInt(value string, base int) (*big.Int, error) {
	bigvalue := new(big.Int)
	var success bool
	if base == 16 {
		//		if strings.Index(value, "0x") != -1 {
		//			value = common.Substr(value, 2, len(value))
		//		}
		value = removeOxFromHex(value)
	}

	_, success = bigvalue.SetString(value, base)
	if !success {
		errInfo := fmt.Sprintf("convert value [%v] to bigint failed, check the value and base passed through\n", value)
		openwLogger.Log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return bigvalue, nil
}

func UnlockAddr(address string, password string, secs int) error {
	params := []interface{}{
		address,
		password,
		secs,
	}

	result, err := client.Call("personal_unlockAccount", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("unlock address [%v] faield, err = %v \n", address, err))
		return err
	}

	if result.Type != gjson.True {
		openwLogger.Log.Errorf(fmt.Sprintf("unlock address [%v] failed", address))
		return errors.New("unlock failed")
	}

	return nil
}

func LockAddr(address string) error {
	params := []interface{}{
		address,
	}

	result, err := client.Call("personal_lockAccount", 1, params)
	if err != nil {
		errInfo := fmt.Sprintf("lock address [%v] faield, err = %v \n", address, err)
		openwLogger.Log.Errorf(errInfo)
		return err
	}

	if result.Type != gjson.True {
		errInfo := fmt.Sprintf("lock address [%v] failed", address)
		openwLogger.Log.Errorf(errInfo)
		return errors.New(errInfo)
	}

	return nil
}

func createRawTransaction(from string, to string, value *big.Int, data string) ([]byte, error) {
	fee, err := GetTransactionFeeEstimated(from, to, value, data)
	if err != nil {
		openwLogger.Log.Errorf("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", from, to, err)
		return nil, err
	}

	nonce, err := GetNonceForAddress(from)
	if err != nil {
		openwLogger.Log.Errorf("GetNonceForAddress from[%v] failed, err=%v", from, err)
		return nil, err
	}

	signer := types.NewEIP155Signer(big.NewInt(chainID))

	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(to),
		value, fee.GasLimit.Uint64(), fee.GasPrice, []byte(data))
	msg := signer.Hash(tx)
	return msg[:], nil
}

func verifyTransaction(nonce uint64, to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) error {
	return nil
}

func ethGetGasPrice() (*big.Int, error) {
	params := []interface{}{}
	result, err := client.Call("eth_gasPrice", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas price failed, err = %v \n", err))
		return big.NewInt(0), err
	}

	if result.Type != gjson.String {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas price failed, response is %v\n", err))
		return big.NewInt(0), err
	}

	gasLimit, err := ConvertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert estimated gas[%v] format to bigint failed, err = %v\n", result.String(), err)
		openwLogger.Log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return gasLimit, nil
}

func ERC20GetWalletBalance(wallet *Wallet) (*big.Int, error) {
	addrs, err := ERC20GetAddressesByWallet(wallet)
	if err != nil {
		openwLogger.Log.Errorf("get address by wallet failed, err = %v", err)
		return big.NewInt(0), nil
	}

	balanceTotal := new(big.Int)
	for _, addr := range addrs {
		fmt.Printf("addr[%v] : %v\n", addr.Address, addr.tokenBalance)
		balanceTotal = balanceTotal.Add(balanceTotal, addr.tokenBalance)
	}

	return balanceTotal, nil
}

//金额的单位是wei
func GetWalletBalance(wallet *Wallet) (*big.Int, error) {
	addrs, err := GetAddressesByWallet(wallet)
	if err != nil {
		openwLogger.Log.Errorf("get address by wallet failed, err = %v", err)
		return big.NewInt(0), nil
	}

	balanceTotal := new(big.Int)
	for _, addr := range addrs {
		/*balance, err := GetAddrBalance(addr.Address)
		if err != nil {
			errinfo := fmt.Sprintf("get balance of addr[%v] failed, err=%v", addr.Address, err)
			return balanceTotal, errors.New(errinfo)
		}*/
		openwLogger.Log.Debugf("addr[%v] : %v\n", addr.Address, addr.balance)
		balanceTotal = balanceTotal.Add(balanceTotal, addr.balance)
	}

	return balanceTotal, nil
}

func ERC20GetAddressesByWallet(wallet *Wallet) ([]*Address, error) {
	addrs := make([]*Address, 0)
	db, err := wallet.OpenDB()
	if err != nil {
		return addrs, err
	}
	defer db.Close()

	err = db.Find("Account", wallet.WalletID, &addrs)
	if err != nil && strings.Index(err.Error(), "not found") == -1 {
		return addrs, err
	}

	for i, _ := range addrs {
		tokenBalance, err := ERC20GetAddressBalance(addrs[i].Address, wallet.erc20Token.Address)
		if err != nil {
			openwLogger.Log.Errorf("get address[%v] erc20 token balance failed, err=%v", addrs[i].Address, err)
			return addrs, err
		}

		balance, err := GetAddrBalance(addrs[i].Address)
		if err != nil {
			errinfo := fmt.Sprintf("get balance of addr[%v] failed, err=%v", addrs[i].Address, err)
			return addrs, errors.New(errinfo)
		}
		addrs[i].tokenBalance = tokenBalance
		addrs[i].balance = balance
	}
	return addrs, nil
}

func GetAddressesByWallet(wallet *Wallet) ([]*Address, error) {
	addrs := make([]*Address, 0)
	db, err := wallet.OpenDB()
	if err != nil {
		return addrs, err
	}
	defer db.Close()

	err = db.Find("Account", wallet.WalletID, &addrs)
	if err != nil && strings.Index(err.Error(), "not found") == -1 {
		return addrs, err
	}

	for i, _ := range addrs {
		balance, err := GetAddrBalance(addrs[i].Address)
		if err != nil {
			errinfo := fmt.Sprintf("get balance of addr[%v] failed, err=%v", addrs[i].Address, err)
			return addrs, errors.New(errinfo)
		}

		addrs[i].balance = balance
	}

	return addrs, nil
}

func ERC20SummaryWallets() {
	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
	//读取参与汇总的钱包
	for _, wallet := range walletsInSum {
		tokenBalance, err := ERC20GetWalletBalance(wallet)
		if err != nil {
			openwLogger.Log.Errorf(fmt.Sprintf("get wallet[%v] ERC20 token balance failed, err = %v", wallet.WalletID, err))
			continue
		}

		if tokenBalance.Cmp(threshold) > 0 {
			log.Printf("Summary account[%s]balance = %v \n", wallet.WalletID, tokenBalance)
			log.Printf("Summary account[%s]Start Send Transaction\n", wallet.WalletID)

			txId, err := ERC20SendTransaction(wallet, sumAddress, tokenBalance, wallet.Password, true)
			if err != nil {
				log.Printf("Summary account[%s]unexpected error: %v\n", wallet.WalletID, err)
				continue
			} else {
				log.Printf("Summary account[%s]successfully，Received Address[%s], TXID：%s\n", wallet.WalletID, sumAddress, txId)
			}
		}
	}
}

func SummaryWallets() {
	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
	//读取参与汇总的钱包
	for _, wallet := range walletsInSum {
		balance, err := GetWalletBalance(wallet)
		if err != nil {
			openwLogger.Log.Errorf(fmt.Sprintf("get wallet[%v] balance failed, err = %v", wallet.WalletID, err))
			continue
		}

		if balance.Cmp(threshold) > 0 {
			log.Printf("Summary account[%s]balance = %v \n", wallet.WalletID, balance)
			log.Printf("Summary account[%s]Start Send Transaction\n", wallet.WalletID)

			txId, err := SendTransaction(wallet, sumAddress, balance, wallet.Password, true)
			if err != nil {
				log.Printf("Summary account[%s]unexpected error: %v\n", wallet.WalletID, err)
				continue
			} else {
				log.Printf("Summary account[%s]successfully，Received Address[%s], TXID：%s\n", wallet.WalletID, sumAddress, txId)
			}
		}
	}

	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}

//RestoreWallet 恢复钱包
func RestoreWallet(keyFile string, password string) error {
	fmt.Printf("Validating key file... \n")

	finfo, err := os.Stat(keyFile)
	if err != nil || !finfo.IsDir() {
		errinfo := fmt.Sprintf("stat file[%v] failed, err = %v\n", keyFile, err)
		openwLogger.Log.Errorf(errinfo)
		return err
	}
	/*parts := strings.Split(keyFile, "\\") //filepath.SplitList(keyFile)
	l := len(parts)
	if l == 0 {
		errinfo := fmt.Sprintf("wrong keyFile[%v] passed through...", keyFile)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}
	*/
	dirName := finfo.Name()

	fmt.Println("dirName:", dirName)
	parts := strings.Split(dirName, "-")
	if len(parts) != 3 {
		errinfo := fmt.Sprintf("invalid directory name[%v] ", dirName)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	_, err = time.ParseInLocation(TIME_POSTFIX, parts[2], time.Local)
	if err != nil {
		errinfo := fmt.Sprintf("check directory name[%v] time format failed ", dirName)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	walletId := parts[1]
	//检查备份路径下key文件的密码
	walletKeyBackupPath := keyFile + "/" + parts[0] + "-" + walletId
	walletBackup, err := GetWalletKey(walletKeyBackupPath)
	if err != nil {
		openwLogger.Log.Errorf("parse the key file [%v] failed, err= %v.", walletKeyBackupPath, err)
		return err
	}
	err = verifyBackupWallet(walletBackup, keyFile, password)
	if err != nil {
		openwLogger.Log.Errorf("verify the backup wallet [%v] password failed, err= %v.", walletKeyBackupPath, err)
		return err
	}

	walletexist, err := GetWalletInfo(walletId)
	if err != nil && err.Error() != WALLET_NOT_EXIST_ERR {
		errinfo := fmt.Sprintf("get wallet [%v] info failed, err = %v ", walletId, err)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	} else if err == nil {
		err = UnlockWallet(walletexist, password)
		if err != nil {
			openwLogger.Log.Errorf("unlock the existing wallet [%v] password failed, err= %v.", walletKeyBackupPath, err)
			return err
		}

		newBackupDir := filepath.Join(backupDir+"/restore", walletexist.FileName()+"-"+common.TimeFormat(TIME_POSTFIX))
		_, err := BackupWallet(newBackupDir, walletexist, password)
		if err != nil {
			errinfo := fmt.Sprintf("backup wallet[%v] before restore failed,err = %v ", walletexist.WalletID, err)
			openwLogger.Log.Errorf(errinfo)
			return errors.New(errinfo)
		}
	} else {

	}

	files, err := ioutil.ReadDir(keyFile)
	if err != nil {
		errinfo := fmt.Sprintf("open directory [%v] failed, err = %v ", keyFile, err)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	filesMap := make(map[string]int)
	for _, fi := range files {
		// Skip any non-key files from the folder
		if skipKeyFile(fi) {
			continue
		}

		//		fmt.Println("filename:", fi.Name())
		if strings.Index(fi.Name(), "--") != -1 && strings.Index(fi.Name(), "UTC") != -1 {
			parts = strings.Split(fi.Name(), "--")
			if len(parts) == 0 {
				//				fmt.Println("1. skipped filename:", fi.Name())
				continue
			}
			if len(parts[len(parts)-1]) != len("50068fd632c1a6e6c5bd407b4ccf8861a589e776") {
				//				fmt.Println("2. skipped filename:", fi.Name())
				continue
			}
			filesMap[fi.Name()] = BACKUP_FILE_TYPE_ADDRESS
		} else if strings.Index(fi.Name(), ".key") != -1 && strings.Index(fi.Name(), "-") != -1 {
			filesMap[fi.Name()] = BACKUP_FILE_TYPE_WALLET_KEY
			//			fmt.Println("key filename:", fi.Name())
		} else if strings.Index(fi.Name(), ".db") != -1 && strings.Index(fi.Name(), "-") != -1 {
			filesMap[fi.Name()] = BACKUP_FILE_TYPE_WALLET_DB
			//			fmt.Println("db filename:", fi.Name())
		} /*else {
			fmt.Println("skipped filename:", fi.Name())
			continue
		}*/
	}

	for filename, filetype := range filesMap {
		src := keyFile + "/" + filename
		var dst string
		//		fmt.Println("src:", src)
		if filetype == BACKUP_FILE_TYPE_ADDRESS {
			dst = EthereumKeyPath + "/"
		} else if filetype == BACKUP_FILE_TYPE_WALLET_DB {
			dst = dbPath + "/"
			//			fmt.Println("db file:", filename)
		} else if filetype == BACKUP_FILE_TYPE_WALLET_KEY {
			dst = keyDir + "/"
			//			fmt.Println("key file:", filename)
		} else {
			continue
		}

		err = file.Copy(src, dst)
		if err != nil {
			errinfo := fmt.Sprintf("copy file from [%v] to [%v] failed, err = %v", src, dst, err)
			openwLogger.Log.Errorf(errinfo)
			return errors.New(errinfo)
		}
	}

	return nil
}

func GetNonceForAddress(address string) (uint64, error) {
	txpool, err := ethGetTxPoolContent()
	if err != nil {
		openwLogger.Log.Errorf("ethGetTxPoolContent failed, err = %v", err)
		return 0, err
	}

	txCount := txpool.GetPendingTxCountForAddr(address)
	openwLogger.Log.Infof("address[%v] has %v tx in pending queue of txpool.", address, txCount)
	for txCount > 0 {
		time.Sleep(time.Second * 1)
		txpool, err = ethGetTxPoolContent()
		if err != nil {
			openwLogger.Log.Errorf("ethGetTxPoolContent failed, err = %v", err)
			return 0, err
		}

		txCount = txpool.GetPendingTxCountForAddr(address)
		openwLogger.Log.Infof("address[%v] has %v tx in pending queue of txpool.", address, txCount)
	}

	nonce, err := ethGetTransactionCount(address)
	if err != nil {
		openwLogger.Log.Errorf("ethGetTransactionCount failed, err=%v", err)
		return 0, err
	}
	return nonce, nil
}
