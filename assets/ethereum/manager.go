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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	//	"log"
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
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCBasedFuncs/owkeychain"
	owcrypt "github.com/blocktree/go-OWCrypt"
	ethKStore "github.com/ethereum/go-ethereum/accounts/keystore"
)

const (
	maxAddresNum = 10000
)

const (
	WALLET_NOT_EXIST_ERR        = "The wallet whose name is given not exist!"
	BACKUP_FILE_TYPE_ADDRESS    = 0
	BACKUP_FILE_TYPE_WALLET_KEY = 1
	BACKUP_FILE_TYPE_WALLET_DB  = 2
)

var (
// 节点客户端
//	client *Client

//秘钥存取
//storage *keystore.HDKeystore
//全局的manager
//	g_manager *WalletManager
)

/*func init() {
	storage = keystore.NewHDKeystore(KeyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	client = &Client{BaseURL: serverAPI, Debug: true}
}*/

type WalletManager struct {
	Storage      *hdkeystore.HDKeystore        //秘钥存取
	WalletClient *Client                       // 节点客户端
	Config       *WalletConfig                 //钱包管理配置
	WalletsInSum map[string]*openwallet.Wallet //参与汇总的钱包
	Blockscanner *ETHBlockScanner              //区块扫描器
	Decoder      openwallet.AddressDecoder     //地址编码器
	TxDecoder    openwallet.TransactionDecoder //交易单编码器
	//	RootDir        string                        //
	locker         sync.Mutex //防止并发修改和读取配置, 可能用不上
	WalletInSumOld map[string]*Wallet
	StorageOld     *keystore.HDKeystore
	ConfigPath     string
	RootPath       string
	DefaultConfig  string
	SymbolID       string
}

func (this *WalletManager) GetConfig() WalletConfig {
	this.locker.Lock()
	defer this.locker.Unlock()
	return *this.Config
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.RootPath = "data"
	wm.ConfigPath = "conf"
	wm.SymbolID = Symbol
	wm.Config = &WalletConfig{}
	wm.DefaultConfig = makeEthDefaultConfig(wm.ConfigPath)

	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = NewETHBlockScanner(&wm)
	wm.Decoder = &AddressDecoder{}
	wm.TxDecoder = NewTransactionDecoder(&wm)

	wm.WalletInSumOld = make(map[string]*Wallet)
	return &wm
}

func (this *WalletManager) CreateWallet(name string, password string) (*Wallet, string, error) {
	//检查钱包名是否存在
	wallets, err := GetWalletKeys(this.GetConfig().KeyDir)
	if err != nil {
		log.Error("GetWalletKeys failed, err=", err)
		return nil, "", errors.New(fmt.Sprintf("get wallet keys failed, err = %v", err))
	}

	for _, w := range wallets {
		if w.Alias == name {
			log.Error("duplicated alias.")
			return nil, "", errors.New("The wallet's alias is duplicated!")
		}
	}

	//生成keystore
	key, filePath, err := hdkeystore.StoreHDKey(this.GetConfig().KeyDir, name, password, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	if err != nil {
		log.Error("create hdkeystore failed, err=", err)
		return nil, "", err
	}

	// root/n' , 使用强化方案
	hdPath := fmt.Sprintf("%s/%d'", key.RootPath, 1)
	childKey, err := key.DerivedKeyWithPath(hdPath, owcrypt.ECC_CURVE_SECP256K1)
	if err != nil {
		log.Error("generate child key failed, err=", err)
		return nil, "", err
	}

	publicKey := childKey.GetPublicKey().OWEncode()

	w := Wallet{
		WalletID:  key.KeyID,
		Alias:     key.Alias,
		RootPath:  key.RootPath,
		KeyFile:   filePath,
		HdPath:    hdPath,
		PublicKey: publicKey,
	}

	db, err := w.OpenDB(this.GetConfig().DbPath)
	if err != nil {
		log.Error("open wallet db[", w.Alias, "] failed, err=")
		return nil, "", err
	}
	defer db.Close()

	err = db.Save(&w)
	if err != nil {
		log.Error("save wallet[", w.Alias, "] to db failed, err=", err)
		return nil, "", err
	}

	return &w, filePath, nil
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
	type keyType struct {
		Alias    string `json:"alias"`
		KeyId    string `json:"keyid"`
		RootPath string `json:"rootpath"`
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
		var key keyType
		err = json.NewDecoder(buf).Decode(&key)
		if err != nil {
			return nil
		}

		return &Wallet{WalletID: key.KeyId, Alias: key.Alias, RootPath: key.RootPath}
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

		path := filepath.Join(dir, fi.Name())

		w := readWallet(path)
		w.KeyFile = fi.Name()
		//fmt.Println("absolute path:", absPath)
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

func (this *WalletManager) SaveERC20TokenConfig(config *ERC20Token) error {
	db, err := OpenDB(this.GetConfig().DbPath, ERC20TOKEN_DB)
	defer db.Close()
	if err != nil {
		openwLogger.Log.Errorf("open db for path [%v] failed, err = %v", this.GetConfig().DbPath+"/"+ERC20TOKEN_DB, err)
		return err
	}
	err = db.Save(config)
	if err != nil {
		openwLogger.Log.Errorf("save db for path [%v] failed, err = %v", this.GetConfig().DbPath+"/"+ERC20TOKEN_DB, err)
		return err
	}
	return nil
}

func (this *WalletManager) GetERC20TokenList() ([]ERC20Token, error) {
	db, err := OpenDB(this.GetConfig().DbPath, ERC20TOKEN_DB)
	defer db.Close()
	if err != nil {
		openwLogger.Log.Errorf("open db for path [%v] failed, err = %v", this.GetConfig().DbPath+"/"+ERC20TOKEN_DB, err)
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

func (this *WalletManager) ERC20GetWalletList(erc20Token *ERC20Token) ([]*Wallet, error) {
	wallets, err := GetWalletKeys(this.GetConfig().KeyDir)
	if err != nil {
		return nil, err
	}

	for i, _ := range wallets {
		wallets[i].erc20Token = &ERC20Token{}
		*wallets[i].erc20Token = *erc20Token
		tokenBanlance, err := this.ERC20GetWalletBalance(wallets[i])
		if err != nil {

			openwLogger.Log.Errorf(fmt.Sprintf("find wallet balance failed, err=%v\n", err))
			return nil, err
		}

		wallets[i].erc20Token.balance = tokenBanlance
	}
	return wallets, nil
}

//GetWalletList 获取钱包列表
func (this *WalletManager) GetLocalWalletList(keyDir string, dbPath string) ([]*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for i, _ := range wallets {
		err = wallets[i].RestoreFromDb(this.GetConfig().DbPath)
		if err != nil {
			log.Error("restore wallet[", wallets[i].WalletID, "] from db failed, err=", err)
			return nil, err
		}

		balance, err := this.GetWalletBalance(dbPath, wallets[i])
		if err != nil {

			openwLogger.Log.Errorf(fmt.Sprintf("find wallet balance failed, err=%v\n", err))
			return nil, err
		}
		wallets[i].balance = balance
	}

	return wallets, nil
}

//AddWalletInSummary 添加汇总钱包账户
func (this *WalletManager) AddWalletInSummary(wid string, wallet *Wallet) {
	this.WalletInSumOld[wid] = wallet
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

func (this *WalletManager) BackupWalletToDefaultPath(wallet *Wallet, password string) (string, error) {
	newBackupDir := filepath.Join(this.GetConfig().BackupDir, wallet.FileName()+"-"+common.TimeFormat(TIME_POSTFIX))
	return this.BackupWallet2(newBackupDir, wallet, password)
}

func (this *WalletManager) BackupWallet2(newBackupDir string, wallet *Wallet,
	password string) (string, error) {
	err := this.UnlockWallet(wallet, password)
	if err != nil {
		openwLogger.Log.Errorf("unlock wallet failed, err=%v", err)
		return "", err
	}

	keyFile := filepath.Join(this.GetConfig().KeyDir, wallet.FileName()+".key")
	dbFile := filepath.Join(this.GetConfig().DbPath, wallet.FileName()+".db")

	file.MkdirAll(newBackupDir)

	//备份钱包key文件
	err = file.Copy(keyFile, newBackupDir)
	if err != nil {
		log.Error("backup key file [", keyFile, "] to ", newBackupDir, " failed, err=", err)
		return "", err
	}

	//备份地址数据库
	err = file.Copy(dbFile, newBackupDir)
	if err != nil {
		log.Error("backup db file [", dbFile, "] to ", newBackupDir, " failed, err=", err)
		return "", err
	}
	return newBackupDir, nil
}

func (this *WalletManager) GetWalletInfo(keyDir string, dbPath string, walletID string) (*Wallet, error) {
	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			err = w.RestoreFromDb(this.GetConfig().DbPath)
			if err != nil {
				log.Error("restore from db failed, err=", err)
				return nil, err
			}
			balance, err := this.GetWalletBalance(dbPath, w)
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

func verifyBackupKey(keyFile string, password string) error {
	keyjson, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return err
	}
	_, err = hdkeystore.DecryptHDKey(keyjson, password)
	if err != nil {
		return err
	}
	return nil
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

func (this *WalletManager) UnlockWallet(wallet *Wallet, password string) error {
	_, err := wallet.HDKey2(password)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get HDkey, err=%v\n", err))
		return err
	}
	return nil
}

func (this *WalletManager) CreateAddressForTest(name, password string, count uint64) (*ethKStore.Key, *Address, error) {
	//读取钱包
	w, err := this.GetWalletInfo("", "", name)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get wallet info, err=%v\n", err))
		return nil, nil, err
	}

	//验证钱包
	keyroot, err := w.HDKey(password, this.StorageOld)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get HDkey, err=%v\n", err))
		return nil, nil, err
	}

	timestamp := uint64(time.Now().Unix())

	keyCombo, address, err := CreateNewPrivateKey(keyroot, timestamp, count)
	if err != nil {
		log.Error("Create new privKey failed unexpected error:", err)
		return nil, nil, err
	}
	return keyCombo, address, nil
}

//exportAddressToFile 导出地址到文件中
func (this *WalletManager) exportAddressToFile(addrs []*Address, filePath string) error {

	var content string

	for _, a := range addrs {

		log.Std.Info("Export: %s ", a.Address)

		content = content + a.Address + "\n"
	}

	file.MkdirAll(this.GetConfig().AddressDir)
	if !file.WriteFile(filePath, []byte(content), true) {
		return errors.New("export address to file failed.")
	}
	return nil
}

func (this *WalletManager) CreateBatchAddress2(name, password string, count uint64) (string, error) {
	//读取钱包
	w, err := this.GetWalletInfo(this.GetConfig().KeyDir, this.GetConfig().DbPath, name)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get wallet info, err=%v\n", err))
		return "", err
	}

	_, err = w.HDKey2(password)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get HDkey, err=%v\n", err))
		return "", err
	}

	db, err := w.OpenDB(this.GetConfig().DbPath)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("open db, err=%v\n", err))
		return "", err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	addressIndex := w.AddressCount
	pubkey, err := owkeychain.OWDecode(w.PublicKey)
	if err != nil {
		log.Error("owkeychain.OWDecode failed, err=", err)
		return "", err
	}

	errcount := uint64(0)
	errMaximum := uint64(15)
	threadControl := make(chan int, 20)
	addressChan := make(chan *Address, 100)
	done := make(chan int, 1)

	generateAddress := func(addressIndex uint64) {
		threadControl <- 1
		var addr *Address
		defer func() {
			addressChan <- addr
			<-threadControl
		}()
		derivedPath := fmt.Sprintf("%s/%d/%d", w.HdPath, 0, addressIndex)
		start, err := pubkey.GenPublicChild(0)
		if err != nil {
			log.Error("pubkey.GenPublicChild failed, err = %v", err)
			return
		}

		derived, err := start.GenPublicChild(uint32(addressIndex))
		if err != nil {
			log.Error("start.GenPublicChild failed, err = %v", err)
			return
		}

		newKey := derived.GetPublicKeyBytes()

		newpubkey := hex.EncodeToString(newKey)
		address, err := this.Decoder.PublicKeyToAddress(newKey, false)
		if err != nil {
			log.Error("decoder.PublicKeyToAddress failed, err = %v", err)
			return
		}

		addr = &Address{
			Address:   address, //address.String(),
			Account:   w.WalletID,
			HDPath:    derivedPath,
			CreatedAt: time.Now(),
			Index:     int(addressIndex),
			PublicKey: newpubkey,
		}
	}

	var addressList []*Address
	go func() {
		for j := uint64(0); j < count; j++ {
			addr := <-addressChan
			if addr == nil {
				errcount++
				continue
			}
			err := tx.Save(addr)
			if err != nil {
				log.Error("save address to db failed, err=", err)
				errcount++
			}
			addressList = append(addressList, addr)
		}
		done <- 1
	}()

	for i := uint64(0); i < count && errcount < errMaximum; i++ {

		go generateAddress(addressIndex)
		addressIndex++

		/*derivedPath := fmt.Sprintf("%s/%d/%d", w.HdPath, 0, addressIndex)
		start, err := pubkey.GenPublicChild(0)
		if err != nil {
			log.Error("pubkey.GenPublicChild failed, err = %v", err)
			errcount++
			continue
		}
		derived, err := start.GenPublicChild(uint32(addressIndex))
		if err != nil {
			log.Error("start.GenPublicChild failed, err = %v", err)
			errcount++
			continue
		}

		newKey := derived.GetPublicKeyBytes()

		newpubkey := hex.EncodeToString(newKey)
		address, err := this.Decoder.PublicKeyToAddress(newKey, false)
		if err != nil {
			log.Error("decoder.PublicKeyToAddress failed, err = %v", err)
			errcount++
			continue
		}

		addr := Address{
			Address:   address, //address.String(),
			Account:   w.WalletID,
			HDPath:    derivedPath,
			CreatedAt: time.Now(),
			Index:     int(addressIndex),
			PublicKey: newpubkey,
		}*/
	}

	<-done

	if errcount > 0 {
		log.Error("errors ocurred exceed the maximum. ")
		return "", errors.New("errors ocurred exceed the maximum. ")
	}

	w.AddressCount = addressIndex
	err = tx.Save(w)
	if err != nil {
		log.Error("save wallet to db failed, err=", err)
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		log.Error("commit address failed, err=", err)
		return "", err
	}

	filename := "address-" + common.TimeFormat("20060102150405", time.Now()) + ".txt"
	filePath := filepath.Join(this.GetConfig().AddressDir, filename)
	this.exportAddressToFile(addressList, filePath)
	return filePath, nil
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

func (this *txFeeInfo) CalcFee() error {
	fee := new(big.Int)
	fee.Mul(this.GasLimit, this.GasPrice)
	this.Fee = fee
	return nil
}

func (this *WalletManager) GetTransactionFeeEstimated(from string, to string, value *big.Int, data string) (*txFeeInfo, error) {
	gasLimit, err := this.WalletClient.ethGetGasEstimated(makeGasEstimatePara(from, to, value, data))
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas limit failed, err = %v\n", err))
		return nil, err
	}

	gasPrice, err := this.WalletClient.ethGetGasPrice()
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get gas price failed, err = %v\n", err))
		return nil, err
	}

	//	fee := new(big.Int)
	//	fee.Mul(gasLimit, gasPrice)

	feeInfo := &txFeeInfo{
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		//		Fee:      fee,
	}

	feeInfo.CalcFee()
	return feeInfo, nil
}

func (this *WalletManager) GetERC20TokenTransactionFeeEstimated(from string, to string, data string) (*txFeeInfo, error) {
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
	return this.GetTransactionFeeEstimated(from, to, nil, data)
}

func (this *WalletManager) GetSimpleTransactionFeeEstimated(from string, to string, amount *big.Int) (*txFeeInfo, error) {
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
	return this.GetTransactionFeeEstimated(from, to, amount, "")
}

func (this *WalletManager) ERC20SendTransaction(wallet *Wallet, to string, amount *big.Int, password string, feesInSender bool) ([]string, error) {
	var txIds []string

	err := this.UnlockWallet(wallet, password)
	if err != nil {
		openwLogger.Log.Errorf("unlock wallet [%v]. failed, err=%v", wallet.WalletID, err)
		return nil, err
	}

	addrs, err := this.ERC20GetAddressesByWallet(this.GetConfig().DbPath, wallet)
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
		fee, err = this.GetERC20TokenTransactionFeeEstimated(addrs[i].Address, wallet.erc20Token.Address, dataPara)
		if err != nil {
			openwLogger.Log.Errorf("get erc token transaction fee estimated failed, err = %v", err)
			continue
		}

		if addrs[i].balance.Cmp(fee.Fee) < 0 {
			openwLogger.Log.Errorf("address[%v] cannot afford a token transfer with a fee [%v]", addrs[i].Address, fee.Fee)
			continue
		}

		txid, err := this.SendTransactionToAddr(makeERC20TokenTransactionPara(addrs[i], wallet.erc20Token.Address, dataPara, password, fee))
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

func (this *WalletManager) SendTransaction2(wallet *Wallet, to string,
	amount *big.Int, password string, feesInSender bool) ([]string, error) {
	var txIds []string

	masterKey, err := wallet.HDKey2(password)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get HDkey, err=%v\n", err))
		return nil, err
	}

	addrs, err := this.GetAddressesByWallet(this.GetConfig().DbPath, wallet)
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

		//如果该地址的余额足够支付转账
		if addrs[i].balance.Cmp(amount) >= 0 {
			amountToSend = *amount
			fee, err = this.GetSimpleTransactionFeeEstimated(addrs[i].Address, to, &amountToSend)
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
			fee, err = this.GetSimpleTransactionFeeEstimated(addrs[i].Address, to, &amountToSend)
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

		priKey, err := addrs[i].CalcPrivKey(masterKey)
		if err != nil {
			log.Error("calc private key failed, err=", err)
			continue
		}

		nonce := addrs[i].TxCount
		if !this.GetConfig().LocalNonce {
			nonce, err = this.GetNonceForAddress2(addrs[i].Address)
			if err != nil {
				log.Error("get nonce failed, err=", err)
				continue
			}
		}

		raw, err := signEthTransaction(priKey, to, &amountToSend, nonce, "", fee, this.GetConfig().ChainID)
		if err != nil {
			log.Error("signEthTransaction failed, err=", err)
			continue
		}

		//txid, err := this.SendTransactionToAddr(makeSimpleTransactionPara(addrs[i], to, &amountToSend, password, fee))
		txid, err := this.WalletClient.ethSendRawTransaction(raw)
		if err != nil {
			openwLogger.Log.Errorf("SendTransactionToAddr failed, err=%v", err)
			if txid == "" {
				if this.GetConfig().LocalNonce {
					txCount, err := this.WalletClient.ethGetTransactionCount(addrs[i].Address)
					if err != nil {
						log.Error("ethGetTransactionCount failed, err=", err)
						continue
					}

					if nonce < txCount {
						addrs[i].TxCount = txCount
						err = wallet.SaveAddress(this.GetConfig().DbPath, addrs[i])
						if err != nil {
							log.Error("update address[", addrs[i].Address, "] tx count to ", addrs[i].TxCount, " failed, err=", err)
						}
						continue
					}
				}
				return txIds, err //continue //txIds = append(txIds, txid)
			}
		}

		addrs[i].TxCount = nonce + 1
		err = wallet.SaveAddress(this.GetConfig().DbPath, addrs[i])
		if err != nil {
			log.Error("update address[", addrs[i].Address, "] tx count to ", addrs[i].TxCount, " failed, err=", err)
			continue
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

func (this *Client) UnlockAddr(address string, password string, secs int) error {
	params := []interface{}{
		address,
		password,
		secs,
	}

	result, err := this.Call("personal_unlockAccount", 1, params)
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

func (this *Client) LockAddr(address string) error {
	params := []interface{}{
		address,
	}

	result, err := this.Call("personal_lockAccount", 1, params)
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

/*func createRawTransaction(from string, to string, value *big.Int, data string) ([]byte, error) {
	fee, err := GetTransactionFeeEstimated(from, to, value, data)
	if err != nil {
		openwLogger.Log.Errorf("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", from, to, err)
		return nil, err
	}

	nonce, err := GetNonceForAddress2(from)
	if err != nil {
		openwLogger.Log.Errorf("GetNonceForAddress from[%v] failed, err=%v", from, err)
		return nil, err
	}

	signer := types.NewEIP155Signer(big.NewInt(CHAIN_ID))

	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(to),
		value, fee.GasLimit.Uint64(), fee.GasPrice, []byte(data))
	msg := signer.Hash(tx)
	return msg[:], nil
}*/

func verifyTransaction(nonce uint64, to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) error {
	return nil
}

func (this *Client) ethGetGasPrice() (*big.Int, error) {
	params := []interface{}{}
	result, err := this.Call("eth_gasPrice", 1, params)
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

func (this *WalletManager) ERC20GetWalletBalance(wallet *Wallet) (*big.Int, error) {
	addrs, err := this.ERC20GetAddressesByWallet(this.GetConfig().DbPath, wallet)
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
func (this *WalletManager) GetWalletBalance(dbPath string, wallet *Wallet) (*big.Int, error) {
	addrs, err := this.GetAddressesByWallet(dbPath, wallet)
	if err != nil {
		openwLogger.Log.Errorf("get address by wallet failed, err = %v", err)
		return big.NewInt(0), err
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

func (this *WalletManager) ERC20GetAddressesByWallet(dbPath string, wallet *Wallet) ([]*Address, error) {
	addrs := make([]*Address, 0)
	db, err := wallet.OpenDB(dbPath)
	if err != nil {
		return addrs, err
	}
	defer db.Close()

	err = db.Find("Account", wallet.WalletID, &addrs)
	if err != nil && strings.Index(err.Error(), "not found") == -1 {
		return addrs, err
	}

	for i, _ := range addrs {
		tokenBalance, err := this.WalletClient.ERC20GetAddressBalance(addrs[i].Address, wallet.erc20Token.Address)
		if err != nil {
			openwLogger.Log.Errorf("get address[%v] erc20 token balance failed, err=%v", addrs[i].Address, err)
			return addrs, err
		}

		balance, err := this.WalletClient.GetAddrBalance(addrs[i].Address)
		if err != nil {
			errinfo := fmt.Sprintf("get balance of addr[%v] failed, err=%v", addrs[i].Address, err)
			return addrs, errors.New(errinfo)
		}
		addrs[i].tokenBalance = tokenBalance
		addrs[i].balance = balance
	}
	return addrs, nil
}

func (this *WalletManager) GetAddressesByWallet(dbPath string, wallet *Wallet) ([]*Address, error) {
	addrs := make([]*Address, 0)
	db, err := wallet.OpenDB(dbPath)
	if err != nil {
		return addrs, err
	}
	defer db.Close()

	err = db.Find("Account", wallet.WalletID, &addrs)
	if err != nil && strings.Index(err.Error(), "not found") == -1 {
		return addrs, err
	}

	count := len(addrs)

	queryBalanceChan := make(chan int, 20)
	resultChan := make(chan *Address, 100)
	done := make(chan int, 1)

	queryBalance := func(theAddr *Address) {
		queryBalanceChan <- 1
		var paddr *Address
		defer func() {
			resultChan <- paddr
			<-queryBalanceChan
		}()
		balance, err := this.WalletClient.GetAddrBalance(theAddr.Address)
		if err != nil {
			errinfo := fmt.Sprintf("get balance of addr[%v] failed, err=%v", theAddr.Address, err)
			log.Error(errinfo)
			return
		}

		theAddr.balance = balance
		paddr = theAddr
	}

	var addrResults []*Address
	go func() {
		for i := 0; i < count; i++ {
			paddr := <-resultChan
			if paddr != nil {
				addrResults = append(addrResults, paddr)
			}
		}
		done <- 1
	}()

	for i, _ := range addrs {
		go queryBalance(addrs[i])
	}
	<-done

	if len(addrResults) != count {
		log.Error("get address balance failed in this wallet.")
		return make([]*Address, 0), errors.New("get address balance failed in this wallet.")
	}
	return addrResults, nil
}

func (this *WalletManager) ERC20SummaryWallets() {
	log.Info("[Summary Wallet Start]------", common.TimeFormat("2006-01-02 15:04:05"))
	//读取参与汇总的钱包
	for _, wallet := range this.WalletInSumOld {
		tokenBalance, err := this.ERC20GetWalletBalance(wallet)
		if err != nil {
			openwLogger.Log.Errorf(fmt.Sprintf("get wallet[%v] ERC20 token balance failed, err = %v", wallet.WalletID, err))
			continue
		}

		if tokenBalance.Cmp(this.GetConfig().Threshold) > 0 {
			log.Debugf("Summary account[%s]balance = %v \n", wallet.WalletID, tokenBalance)
			log.Debugf("Summary account[%s]Start Send Transaction\n", wallet.WalletID)

			txId, err := this.ERC20SendTransaction(wallet, this.GetConfig().SumAddress, tokenBalance, wallet.Password, true)
			if err != nil {
				log.Debugf("Summary account[%s]unexpected error: %v\n", wallet.WalletID, err)
				continue
			} else {
				log.Debugf("Summary account[%s]successfully，Received Address[%s], TXID：%s\n", wallet.WalletID, this.GetConfig().SumAddress, txId)
			}
		}
	}
}

func (this *WalletManager) SummaryWallets() {
	log.Debugf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
	//读取参与汇总的钱包
	for _, wallet := range this.WalletInSumOld {
		balance, err := this.GetWalletBalance(this.GetConfig().DbPath, wallet)
		if err != nil {
			openwLogger.Log.Errorf(fmt.Sprintf("get wallet[%v] balance failed, err = %v", wallet.WalletID, err))
			continue
		}

		if balance.Cmp(this.GetConfig().Threshold) > 0 {
			log.Debugf("Summary account[%s]balance = %v \n", wallet.WalletID, balance)
			log.Debugf("Summary account[%s]Start Send Transaction\n", wallet.WalletID)

			txId, err := this.SendTransaction2(wallet, this.GetConfig().SumAddress, balance, wallet.Password, true)
			if err != nil {
				log.Debugf("Summary account[%s]unexpected error: %v\n", wallet.WalletID, err)
				continue
			} else {
				log.Debugf("Summary account[%s]successfully，Received Address[%s], TXID：%s\n", wallet.WalletID, this.GetConfig().SumAddress, txId)
			}
		}
	}

	log.Debugf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}

func (this *WalletManager) RestoreWallet2(backupPath string, password string) error {
	//检查用户输入的备份路径是否存在
	finfo, err := os.Stat(backupPath)
	if err != nil || !finfo.IsDir() {
		errinfo := fmt.Sprintf("stat file[%v] failed, err = %v\n", backupPath, err)
		openwLogger.Log.Errorf(errinfo)
		return err
	}

	//检查备份文件夹的命名是否规范, 形如:peter-WAKGjuuGL6ifMqdJhZ19TfQDTafH5hsvmw-20180810171606
	backupDirName := finfo.Name()
	parts := strings.Split(backupDirName, "-")
	if len(parts) != 3 {
		errinfo := fmt.Sprintf("invalid directory name[%v] ", backupDirName)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	_, err = time.ParseInLocation(TIME_POSTFIX, parts[2], time.Local)
	if err != nil {
		errinfo := fmt.Sprintf("check directory name[%v] time format failed ", backupDirName)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	//验证被恢复钱包的密码
	walletId := parts[1]
	alias := parts[0]
	keyFileBackup := filepath.Join(backupPath, alias+"-"+walletId+".key")
	dbFileBackup := filepath.Join(backupPath, alias+"-"+walletId+".db")
	err = verifyBackupKey(keyFileBackup, password)
	if err != nil {
		log.Error("wrong password, restore process has been rejected.")
		return err
	}

	//检查要被恢复的钱包是否已经存在于当前目录, 如果存在需要先备份这个钱包
	keyFileRestore := filepath.Join(this.GetConfig().KeyDir, alias+"-"+walletId+".key")
	finfo, err = os.Stat(keyFileRestore)
	if err == nil {
		//当前钱包已经存在, 备份钱包
		walletExist := &Wallet{WalletID: walletId, Alias: alias}
		err = walletExist.RestoreFromDb(this.GetConfig().DbPath)
		if err != nil {
			log.Error("restore existing wallet from db failed, err=", err)
			return err
		}
		restorePath := filepath.Join(this.GetConfig().BackupDir, "restore")
		_, err = this.BackupWallet2(restorePath, walletExist, password)
		if err != nil {
			log.Error("backup existing wallet before restore failed, err=", err)
			return err
		}
	} else if err != nil && !os.IsNotExist(err) {
		log.Error("unexpected error, err=", err)
		return err
	}

	err = file.Copy(dbFileBackup, this.GetConfig().DbPath)
	if err != nil {
		log.Error("restore db file failed, err=", err)
		return err
	}

	err = file.Copy(keyFileBackup, this.GetConfig().KeyDir)
	if err != nil {
		log.Error("restore db file failed, err=", err)
		return err
	}

	return nil
}

func (this *WalletManager) GetNonceForAddress2(address string) (uint64, error) {
	txpool, err := this.WalletClient.ethGetTxPoolContent()
	if err != nil {
		openwLogger.Log.Errorf("ethGetTxPoolContent failed, err = %v", err)
		return 0, err
	}

	_, max, count, err := txpool.GetSequentTxNonce(address)
	if err != nil {
		log.Error("get txpool sequent tx nonce failed, err=%v", err)
		return 0, err
	}
	log.Debugf("sequent max nonce:%v", max)
	log.Debugf("sequent nonce count:%v", count)
	txCount, err := this.WalletClient.ethGetTransactionCount(address)
	if err != nil {
		log.Error("ethGetTransactionCount failed, err=", err)
		return 0, err
	}

	if count == 0 || max < txCount {
		return txCount, nil
	}
	return max + 1, nil
}

func (this *WalletManager) GetNonceForAddress(address string) (uint64, error) {
	txpool, err := this.WalletClient.ethGetTxPoolContent()
	if err != nil {
		openwLogger.Log.Errorf("ethGetTxPoolContent failed, err = %v", err)
		return 0, err
	}

	txCount := txpool.GetPendingTxCountForAddr(address)
	openwLogger.Log.Infof("address[%v] has %v tx in pending queue of txpool.", address, txCount)
	for txCount > 0 {
		time.Sleep(time.Second * 1)
		txpool, err = this.WalletClient.ethGetTxPoolContent()
		if err != nil {
			openwLogger.Log.Errorf("ethGetTxPoolContent failed, err = %v", err)
			return 0, err
		}

		txCount = txpool.GetPendingTxCountForAddr(address)
		openwLogger.Log.Infof("address[%v] has %v tx in pending queue of txpool.", address, txCount)
	}

	nonce, err := this.WalletClient.ethGetTransactionCount(address)
	if err != nil {
		openwLogger.Log.Errorf("ethGetTransactionCount failed, err=%v", err)
		return 0, err
	}
	return nonce, nil
}
