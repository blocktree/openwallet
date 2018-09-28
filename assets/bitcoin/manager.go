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
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCBasedFuncs/owkeychain"
	"github.com/bndr/gotabulate"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/codeskyblue/go-sh"
	"github.com/shopspring/decimal"
)

const (
	maxAddresNum = 10000
)

type WalletManager struct {
	Storage      *hdkeystore.HDKeystore        //秘钥存取
	WalletClient *Client                       // 节点客户端
	Config       *WalletConfig                 //钱包管理配置
	WalletsInSum map[string]*openwallet.Wallet //参与汇总的钱包
	Blockscanner *BTCBlockScanner              //区块扫描器
	Decoder      openwallet.AddressDecoder     //地址编码器
	TxDecoder    openwallet.TransactionDecoder //交易单编码器
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey)
	storage := hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.Storage = storage
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = NewBTCBlockScanner(&wm)
	wm.Decoder = &addressDecoder{}
	wm.TxDecoder = NewTransactionDecoder(&wm)
	return &wm
}

func (wm *WalletManager) GetAddressesByAccount(walletID string) ([]string, error) {

	var (
		addresses = make([]string, 0)
	)

	request := []interface{}{
		walletID,
	}

	result, err := wm.WalletClient.Call("getaddressesbyaccount", request)
	if err != nil {
		return nil, err
	}

	array := result.Array()
	for _, a := range array {
		addresses = append(addresses, a.String())
	}

	return addresses, nil

}

//GetAddressesFromLocalDB 从本地数据库
func (wm *WalletManager) GetAddressesFromLocalDB(walletID string, offset, limit int) ([]*openwallet.Address, error) {

	wallet, err := wm.GetWalletInfo(walletID)
	if err != nil {
		return nil, err
	}

	db, err := wallet.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var addresses []*openwallet.Address
	//err = db.Find("WalletID", walletID, &addresses)
	if limit > 0 {
		err = db.Find("AccountID", walletID, &addresses, storm.Limit(limit), storm.Skip(offset))
	} else {
		err = db.Find("AccountID", walletID, &addresses, storm.Skip(offset))
	}

	if err != nil {
		return nil, err
	}

	return addresses, nil

}

//ImportPrivKey 导入私钥
func (wm *WalletManager) ImportPrivKey(wif, walletID string) error {

	request := []interface{}{
		wif,
		walletID,
		false,
	}

	_, err := wm.WalletClient.Call("importprivkey", request)

	if err != nil {
		return err
	}

	return err

}

//ImportAddress 导入地址核心钱包
func (wm *WalletManager) ImportAddress(address *openwallet.Address) error {

	request := []interface{}{
		address.Address,
		address.AccountID,
		false,
	}

	_, err := wm.WalletClient.Call("importaddress", request)

	if err != nil {
		return err
	}

	return nil

}

//ImportMulti 批量导入地址和私钥
func (wm *WalletManager) ImportMulti(addresses []*openwallet.Address, keys []string, watchOnly bool) ([]int, error) {

	/*
		[
		{
			"scriptPubKey" : { "address": "1NL9w5fP9kX2D9ToNZPxaiwFJCngNYEYJo" },
			"timestamp" : 0,
			"label" : "Personal"
		},
		{
			"scriptPubKey" : "76a9149e857da0a5b397559c78c98c9d3f7f655d19c68688ac",
			"timestamp" : 1493912405,
			"label" : "TestFailure"
		}
		]' '{ "rescan": true }'
	*/

	var (
		request     []interface{}
		imports     = make([]interface{}, 0)
		failedIndex = make([]int, 0)
	)

	if len(addresses) != len(keys) && !watchOnly {
		return nil, errors.New("Import addresses is not equal keys count!")
	}

	for i, a := range addresses {

		obj := map[string]interface{}{
			"scriptPubKey": map[string]interface{}{
				"address": a.Address,
			},
			"label":     a.AccountID,
			"timestamp": "now",
			"watchonly": watchOnly,
		}

		if !watchOnly {
			k := keys[i]
			obj["keys"] = []string{k}
		}

		imports = append(imports, obj)
	}

	request = []interface{}{
		imports,
		map[string]interface{}{
			"rescan": false,
		},
	}

	result, err := wm.WalletClient.Call("importmulti", request)
	if err != nil {
		return nil, err
	}

	for i, r := range result.Array() {
		if !r.Get("success").Bool() {
			failedIndex = append(failedIndex, i)
		}
	}

	return failedIndex, err

}

//GetCoreWalletinfo 获取核心钱包节点信息
func (wm *WalletManager) GetCoreWalletinfo() error {

	_, err := wm.WalletClient.Call("getwalletinfo", nil)

	if err != nil {
		return err
	}

	return err

}

//UnlockWallet 解锁钱包
func (wm *WalletManager) UnlockWallet(passphrase string, seconds int) error {

	request := []interface{}{
		passphrase,
		seconds,
	}

	_, err := wm.WalletClient.Call("walletpassphrase", request)
	if err != nil {
		return err
	}

	return err

}

//LockWallet 锁钱包
func (wm *WalletManager) LockWallet() error {

	_, err := wm.WalletClient.Call("walletlock", nil)
	if err != nil {
		return err
	}

	return err
}

//GetNetworkInfo 获取网络信息
func (wm *WalletManager) GetNetworkInfo() error {

	_, err := wm.WalletClient.Call("getnetworkinfo", nil)
	if err != nil {
		return err
	}

	return err
}

//KeyPoolRefill 重新填充私钥池
func (wm *WalletManager) KeyPoolRefill(keyPoolSize uint64) error {

	request := []interface{}{
		keyPoolSize,
	}

	_, err := wm.WalletClient.Call("keypoolrefill", request)
	if err != nil {
		return err
	}

	return nil
}

//CreateReceiverAddress 给指定账户创建地址
func (wm *WalletManager) CreateReceiverAddress(account string) (string, error) {

	request := []interface{}{
		account,
	}

	result, err := wm.WalletClient.Call("getnewaddress", request)
	if err != nil {
		return "", err
	}

	return result.String(), err

}

//CreateBatchAddress 批量创建地址
func (wm *WalletManager) CreateBatchAddress(name, password string, count uint64) (string, []*openwallet.Address, error) {

	var (
		synCount   uint64 = 20
		quit              = make(chan struct{})
		done              = 0 //完成标记
		shouldDone        = 0 //需要完成的总数
	)

	//读取钱包
	w, err := wm.GetWalletInfo(name)
	if err != nil {
		return "", nil, err
	}

	//加载钱包
	key, err := w.HDKey(password)
	if err != nil {
		return "", nil, err
	}

	timestamp := time.Now()
	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405", timestamp) + ".txt"
	filePath := filepath.Join(wm.Config.addressDir, filename)

	//生产通道
	producer := make(chan []*openwallet.Address)
	defer close(producer)

	//消费通道
	worker := make(chan []*openwallet.Address)
	defer close(worker)

	//保存地址过程
	saveAddressWork := func(addresses chan []*openwallet.Address, filename string, wallet *openwallet.Wallet) {

		var (
			saveErr error
		)

		for {
			//回收创建的地址
			getAddrs := <-addresses

			//批量写入数据库
			saveErr = wm.saveAddressToDB(getAddrs, wallet)
			//数据保存成功才导出文件
			if saveErr == nil {
				//导出一批地址
				wm.exportAddressToFile(getAddrs, filename)
			}

			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//解锁钱包
	err = wm.UnlockWallet(password, 3600)
	if err != nil {
		return "", nil, err
	}

	/*	开启导出的线程，监听新地址，批量导出	*/

	go saveAddressWork(worker, filePath, w)

	/*	计算synCount个线程，内部运行的次数	*/

	//每个线程内循环的数量，以synCount个线程并行处理
	runCount := count / synCount
	otherCount := count % synCount

	if runCount > 0 {

		for i := uint64(0); i < synCount; i++ {

			//开始创建地址
			log.Std.Info("Start create address thread[%d]", i)
			s := i * runCount
			e := (i + 1) * runCount
			go wm.createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		log.Std.Info("Start create address thread[REST]")
		s := count - otherCount
		e := count
		go wm.createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)

		shouldDone++
	}

	values := make([][]*openwallet.Address, 0)
	outputAddress := make([]*openwallet.Address, 0)

	//以下使用生产消费模式

	for {

		var activeWorker chan<- []*openwallet.Address
		var activeValue []*openwallet.Address

		//当数据队列有数据时，释放顶部，激活消费
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
			outputAddress = append(outputAddress, pa...)
			//log.Std.Info("completed %d", len(pa))
			//当激活消费者后，传输数据给消费者，并把顶部数据出队
		case activeWorker <- activeValue:
			//log.Std.Info("Get %d", len(activeValue))
			values = values[1:]

		case <-quit:
			//退出
			wm.LockWallet()
			log.Std.Info("All addresses have been created!")
			return filePath, outputAddress, nil
		}
	}

	wm.LockWallet()

	return filePath, outputAddress, nil
}

//CreateNewWallet 创建钱包
func (wm *WalletManager) CreateNewWallet(name, password string) (*openwallet.Wallet, string, error) {

	var (
		err     error
		wallets []*openwallet.Wallet
	)

	//检查钱包名是否存在
	wallets, err = wm.GetWallets()
	for _, w := range wallets {
		if w.Alias == name {
			return nil, "", errors.New("The wallet's alias is duplicated!")
		}
	}

	fmt.Printf("Verify password in core wallet...")

	err = wm.EncryptWallet(password)
	if err != nil {
		//钱包已经加密，解锁钱包1秒，检查密码
		err = wm.UnlockWallet(password, 1)
		if err != nil {
			return nil, "", fmt.Errorf("The wallet's password can not unlock core wallet, unexpected error: %v ", err)
		}
	} else {
		//加密钱包后，需要10秒后重启bitcoin core
		fmt.Printf("Start node server... \n")
		time.Sleep(10 * time.Second)
		wm.startNode()
	}

	fmt.Printf("Create new wallet keystore...\n")

	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		return nil, "", err
	}

	extSeed, err := hdkeystore.GetExtendSeed(seed, wm.Config.MasterKey)
	if err != nil {
		return nil, "", err
	}

	key, keyFile, err := hdkeystore.StoreHDKeyWithSeed(wm.Config.keyDir, name, password, extSeed, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	if err != nil {
		return nil, "", err
	}

	file.MkdirAll(wm.Config.dbPath)
	file.MkdirAll(wm.Config.keyDir)

	w := &openwallet.Wallet{
		WalletID: key.KeyID,
		Alias:    key.Alias,
		KeyFile:  keyFile,
		DBFile:   filepath.Join(wm.Config.dbPath, key.FileName()+".db"),
	}

	w.SaveToDB()

	//w := Wallet{WalletID: key.RootId, Alias: key.Alias}

	return w, keyFile, nil
}

//EncryptWallet 通过密码加密钱包，只在第一次加密码时才有效
func (wm *WalletManager) EncryptWallet(password string) error {

	request := []interface{}{
		password,
	}

	_, err := wm.WalletClient.Call("encryptwallet", request)
	if err != nil {
		return err
	}
	return nil
}

//GetWallets 获取钱包列表
func (wm *WalletManager) GetWallets() ([]*openwallet.Wallet, error) {

	wallets, err := openwallet.GetWalletsByKeyDir(wm.Config.keyDir)
	if err != nil {
		return nil, err
	}

	for _, w := range wallets {
		w.DBFile = filepath.Join(wm.Config.dbPath, w.FileName()+".db")
	}

	return wallets, nil

}

//GetWalletInfo 获取钱包列表
func (wm *WalletManager) GetWalletInfo(walletID string) (*openwallet.Wallet, error) {

	wallets, err := wm.GetWallets()
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			return w, nil
		}

	}

	return nil, errors.New("The wallet that your given name is not exist!")
}

//GetWalletBalance 获取钱包余额
func (wm *WalletManager) GetWalletBalance(accountID string) string {

	//request := []interface{}{
	//	name,
	//	1,
	//	true,
	//}

	wm.RebuildWalletUnspent(accountID)
	utxos, err := wm.ListUnspentFromLocalDB(accountID)
	if err != nil {
		return "0"
	}
	balance := decimal.New(0, 0)

	//批量插入到本地数据库
	//设置utxo的钱包账户
	for _, utxo := range utxos {
		if accountID == utxo.AccountID {
			amount, _ := decimal.NewFromString(utxo.Amount)
			balance = balance.Add(amount)
		}
	}

	//for _, u := range utxos {
	//	amount, _ := decimal.NewFromString(u.Amount)
	//	balance = balance.Add(amount)
	//}

	//balance, err := wm.WalletClient.Call("getbalance", request)
	//if err != nil {
	//	return "", err
	//}

	return balance.StringFixed(8)
}

//GetAddressBalance 获取地址余额
func (wm *WalletManager) GetAddressBalance(walletID, address string) string {

	wm.RebuildWalletUnspent(walletID)

	wallet, err := wm.GetWalletInfo(walletID)
	if err != nil {
		return "0"
	}

	db, err := wallet.OpenDB()
	if err != nil {
		return "0"
	}
	defer db.Close()

	var utxos []*Unspent

	err = db.Select(q.And(
		q.Eq("AccountID", walletID),
		q.Eq("Address", address),
	)).Find(&utxos)
	if err != nil {
		return "0"
	}

	balance := decimal.New(0, 0)

	for _, utxo := range utxos {
		amount, _ := decimal.NewFromString(utxo.Amount)
		balance = balance.Add(amount)
	}

	return balance.StringFixed(8)
}

//CreateNewPrivateKey 创建私钥，返回私钥wif格式字符串
func (wm *WalletManager) CreateNewPrivateKey(accountID string, key *owkeychain.ExtendedKey, derivedPath string, index uint64) (string, *openwallet.Address, error) {

	derivedPath = fmt.Sprintf("%s/%d", derivedPath, index)
	//fmt.Printf("derivedPath = %s\n", derivedPath)
	childKey, err := key.GenPrivateChild(uint32(index))
	//childKey, err := key.DerivedKeyWithPath(derivedPath, wm.Config.CurveType)
	if err != nil {
		return "", nil, err
	}

	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return "", nil, err
	}

	//privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)
	//if err != nil {
	//	return "", nil, err
	//}

	wif, err := wm.Decoder.PrivateKeyToWIF(keyBytes, wm.Config.IsTestNet)

	//cfg := chaincfg.MainNetParams
	//if wm.Config.IsTestNet {
	//	cfg = chaincfg.TestNet3Params
	//}

	//wif, err := btcutil.NewWIF(privateKey, &cfg, true)
	//if err != nil {
	//	return "", nil, err
	//}

	publicKey := childKey.GetPublicKeyBytes()

	address, err := wm.Decoder.PublicKeyToAddress(publicKey, wm.Config.IsTestNet)

	//pkHash := btcutil.Hash160(publicKey)
	//address, err :=  btcutil.NewAddressPubKeyHash(pkHash, &cfg)
	//address, err := childKey.Address(&cfg)
	if err != nil {
		return "", nil, err
	}

	addr := openwallet.Address{
		Address:   address,
		AccountID: accountID,
		HDPath:    derivedPath,
		CreatedAt: time.Now().Unix(),
		Symbol:    wm.Config.Symbol,
		Index:     index,
		WatchOnly: false,
	}

	//addr := Address{
	//	Address:   address.String(),
	//	Account:   key.RootId,
	//	HDPath:    derivedPath,
	//	CreatedAt: time.Now(),
	//}

	return wif, &addr, err
}

//CreateBatchPrivateKey
//func (wm *WalletManager) CreateBatchPrivateKey(key *hdkeystore.HDKey, count uint64) ([]string, error) {
//
//	var (
//		wifs = make([]string, 0)
//	)
//
//	start := time.Now().Unix()
//	for i := uint64(0); i < count; i++ {
//		wif, _, err := CreateNewPrivateKey(key, uint64(start), i)
//		if err != nil {
//			continue
//		}
//		wifs = append(wifs, wif)
//	}
//
//	return wifs, nil
//
//}

//BackupWalletData 备份钱包
func (wm *WalletManager) BackupWalletData(dest string) error {

	request := []interface{}{
		dest,
	}

	_, err := wm.WalletClient.Call("backupwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//BackupWallet 备份数据
func (wm *WalletManager) BackupWallet(walletID string) (string, error) {
	w, err := wm.GetWalletInfo(walletID)
	if err != nil {
		return "", err
	}

	//创建备份文件夹
	newBackupDir := filepath.Join(wm.Config.backupDir, w.FileName()+"-"+common.TimeFormat("20060102150405"))
	file.MkdirAll(newBackupDir)

	//创建临时备份文件wallet.dat
	tmpWalletDat := fmt.Sprintf("tmp-walllet-%d.dat", time.Now().Unix())
	tmpWalletDat = filepath.Join(wm.Config.WalletDataPath, tmpWalletDat)

	//1. 备份核心钱包的wallet.dat
	err = wm.BackupWalletData(tmpWalletDat)
	if err != nil {
		return "", err
	}

	//复制临时文件到备份文件夹
	file.Copy(tmpWalletDat, filepath.Join(newBackupDir, "wallet.dat"))

	//删除临时文件
	file.Delete(tmpWalletDat)

	//2. 备份种子文件
	file.Copy(w.KeyFile, newBackupDir)

	//3. 备份地址数据库
	file.Copy(w.DBFile, newBackupDir)

	return newBackupDir, nil
}

//RestoreWallet 恢复钱包
func (wm *WalletManager) RestoreWallet(keyFile, dbFile, datFile, password string) error {

	//根据流程，提供种子文件路径，wallet.dat文件的路径，钱包数据库文件的路径。
	//输入钱包密码。
	//先备份核心钱包原来的wallet.dat到临时文件夹。
	//关闭钱包节点，复制wallet.dat到钱包的data目录下。
	//启动钱包，通过GetCoreWalletinfo检查钱包是否启动了。
	//检查密码是否可以解析种子文件，是否可以解锁钱包。
	//如果密码错误，关闭钱包节点，恢复原钱包的wallet.dat。
	//重新启动钱包。
	//复制种子文件到data/btc/key/。
	//复制钱包数据库文件到data/btc/db/。

	var (
		restoreSuccess = false
		err            error
		key            *hdkeystore.HDKey
		sleepTime      = 30 * time.Second
	)

	fmt.Printf("Validating key file... \n")

	//检查密码是否可以解析种子文件，是否可以解锁钱包。
	key, err = wm.Storage.GetKey("", keyFile, password)
	if err != nil {
		return errors.New("Passowrd is incorrect!")
	}

	//钱包当前的dat文件
	curretWDFile := filepath.Join(wm.Config.WalletDataPath, "wallet.dat")

	//创建临时备份文件wallet.dat，备份
	tmpWalletDat := fmt.Sprintf("restore-walllet-%d.dat", time.Now().Unix())
	tmpWalletDat = filepath.Join(wm.Config.WalletDataPath, tmpWalletDat)

	fmt.Printf("Backup current wallet.dat file... \n")

	err = wm.BackupWalletData(tmpWalletDat)
	if err != nil {
		return err
	}

	//调试使用
	//file.Copy(curretWDFile, tmpWalletDat)

	fmt.Printf("Stop node server... \n")

	//关闭钱包节点
	wm.stopNode()
	time.Sleep(sleepTime)

	fmt.Printf("Restore wallet.dat file... \n")

	//删除当前钱包文件
	file.Delete(curretWDFile)

	//恢复备份dat到钱包数据目录
	err = file.Copy(datFile, wm.Config.WalletDataPath)
	if err != nil {
		return err
	}

	fmt.Printf("Start node server... \n")

	//重新启动钱包
	wm.startNode()
	time.Sleep(sleepTime)

	fmt.Printf("Validating wallet password... \n")

	//检查wallet.dat是否可以解锁钱包
	err = wm.UnlockWallet(password, 1)
	if err != nil {
		restoreSuccess = false
		err = errors.New("Password is incorrect!")
	} else {
		restoreSuccess = true
	}

	if restoreSuccess {
		/* 恢复成功 */

		fmt.Printf("Restore wallet key and datebase file... \n")

		//复制种子文件到data/btc/key/
		file.MkdirAll(wm.Config.keyDir)
		file.Copy(keyFile, filepath.Join(wm.Config.keyDir, key.FileName()+".key"))

		//复制钱包数据库文件到data/btc/db/
		file.MkdirAll(wm.Config.dbPath)
		file.Copy(dbFile, filepath.Join(wm.Config.dbPath, key.FileName()+".db"))

		fmt.Printf("Backup wallet has been restored. \n")

		err = nil
	} else {
		/* 恢复失败还远原来的文件 */

		fmt.Printf("Wallet unlock password is incorrect. \n")

		fmt.Printf("Stop node server... \n")

		//关闭钱包节点
		wm.stopNode()
		time.Sleep(sleepTime)

		fmt.Printf("Restore original wallet.data... \n")

		//删除当前钱包文件
		file.Delete(curretWDFile)

		file.Copy(tmpWalletDat, curretWDFile)

		fmt.Printf("Start node server... \n")

		//重新启动钱包
		wm.startNode()
		time.Sleep(sleepTime)

		fmt.Printf("Original wallet has been restored. \n")

	}

	//删除临时备份的dat文件
	file.Delete(tmpWalletDat)

	return err
}

//DumpWallet 导出钱包所有私钥文件
func (wm *WalletManager) DumpWallet(filename string) error {

	request := []interface{}{
		filename,
	}

	_, err := wm.WalletClient.Call("dumpwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//ImportWallet 导入钱包私钥文件
func (wm *WalletManager) ImportWallet(filename string) error {

	request := []interface{}{
		filename,
	}

	_, err := wm.WalletClient.Call("importwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//GetBlockChainInfo 获取钱包区块链信息
func (wm *WalletManager) GetBlockChainInfo() (*BlockchainInfo, error) {

	result, err := wm.WalletClient.Call("getblockchaininfo", nil)
	if err != nil {
		return nil, err
	}

	blockchain := NewBlockchainInfo(result)

	return blockchain, nil

}

//ListUnspent 获取未花记录
func (wm *WalletManager) ListUnspent(min uint64, addresses ...string) ([]*Unspent, error) {

	var (
		utxos = make([]*Unspent, 0)
	)

	request := []interface{}{
		min,
		9999999,
	}

	if len(addresses) > 0 {
		request = append(request, addresses)
	}

	result, err := wm.WalletClient.Call("listunspent", request)
	if err != nil {
		return nil, err
	}

	array := result.Array()
	for _, a := range array {
		utxos = append(utxos, NewUnspent(&a))
	}

	return utxos, nil

}

//RebuildWalletUnspent 批量插入未花记录到本地
func (wm *WalletManager) RebuildWalletUnspent(walletID string) error {

	var (
		wallet *openwallet.Wallet
	)

	wallets, err := wm.GetWallets()
	if err != nil {
		return err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			wallet = w
			break
		}
	}

	if wallet == nil {
		return errors.New("The wallet that your given name is not exist!")
	}

	//查找核心钱包确认数大于1的
	utxos, err := wm.ListUnspent(0)
	if err != nil {
		return err
	}

	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	//清空历史的UTXO
	db.Drop("Unspent")

	//开始事务
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//批量插入到本地数据库
	//设置utxo的钱包账户
	for _, utxo := range utxos {
		var addr openwallet.Address
		err = db.One("Address", utxo.Address, &addr)
		utxo.AccountID = addr.AccountID
		utxo.HDAddress = addr
		key := common.NewString(fmt.Sprintf("%s_%d_%s", utxo.TxID, utxo.Vout, utxo.Address)).SHA256()
		utxo.Key = key

		err = tx.Save(utxo)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

//ListUnspentFromLocalDB 查询本地数据库的未花记录
func (wm *WalletManager) ListUnspentFromLocalDB(walletID string) ([]*Unspent, error) {

	var (
		wallet *openwallet.Wallet
	)

	wallets, err := wm.GetWallets()
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			wallet = w
			break
		}
	}

	if wallet == nil {
		return nil, errors.New("The wallet that your given name is not exist!")
	}

	db, err := wallet.OpenDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var utxos []*Unspent
	err = db.Find("AccountID", walletID, &utxos)
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

//BuildTransaction 构建交易单
func (wm *WalletManager) BuildTransaction(utxos []*Unspent, to []string, change string, amount []decimal.Decimal, fees decimal.Decimal) (string, decimal.Decimal, error) {

	var (
		inputs      = make([]interface{}, 0)
		outputs     = make(map[string]interface{})
		totalAmount = decimal.New(0, 0)
		totalSend   = decimal.New(0, 0)
	)

	for _, u := range utxos {

		if u.Spendable {
			ua, _ := decimal.NewFromString(u.Amount)
			totalAmount = totalAmount.Add(ua)

			inputs = append(inputs, map[string]interface{}{
				"txid": u.TxID,
				"vout": u.Vout,
			})
		}

	}

	//计算总发送金额
	for _, amount := range amount {
		totalSend = totalSend.Add(amount)
	}

	if totalAmount.LessThan(totalSend) {
		return "", decimal.New(0, 0), errors.New("The balance is not enough!")
	}

	changeAmount := totalAmount.Sub(totalSend).Sub(fees)
	if changeAmount.GreaterThan(decimal.New(0, 0)) {
		//ca, _ := changeAmount.Float64()
		outputs[change] = changeAmount.StringFixed(8)

		fmt.Printf("Create change address for receiving %s coin.\n", outputs[change])
	}

	for i, r := range to {
		outputs[r] = amount[i].StringFixed(8)
	}

	//ta, _ := amount.Float64()
	//outputs[to] = amount.StringFixed(8)

	request := []interface{}{
		inputs,
		outputs,
	}

	rawTx, err := wm.WalletClient.Call("createrawtransaction", request)
	if err != nil {
		return "", decimal.New(0, 0), err
	}

	return rawTx.String(), changeAmount, nil
}

//SignRawTransactionInCoreWallet 钱包交易单
func (wm *WalletManager) SignRawTransactionInCoreWallet(txHex, walletID string, key *hdkeystore.HDKey, utxos []*Unspent) (string, error) {

	var (
		wifs = make([]string, 0)
	)

	//查找未花签名需要的私钥
	for _, u := range utxos {

		childKey, err := key.DerivedKeyWithPath(u.HDAddress.HDPath, wm.Config.CurveType)

		keyBytes, err := childKey.GetPrivateKeyBytes()
		if err != nil {
			return "", err
		}

		privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)

		//privateKey, err := childKey.ECPrivKey()
		//if err != nil {
		//	return "", err
		//}

		cfg := chaincfg.MainNetParams
		if wm.Config.IsTestNet {
			cfg = chaincfg.TestNet3Params
		}

		wif, err := btcutil.NewWIF(privateKey, &cfg, true)
		if err != nil {
			return "", err
		}

		wifs = append(wifs, wif.String())

	}

	request := []interface{}{
		txHex,
		utxos,
		wifs,
	}

	result, err := wm.WalletClient.Call("signrawtransaction", request)
	if err != nil {
		return "", err
	}

	return result.Get("hex").String(), nil

}

//SendRawTransaction 广播交易
func (wm *WalletManager) SendRawTransaction(txHex string) (string, error) {

	request := []interface{}{
		txHex,
	}

	result, err := wm.WalletClient.Call("sendrawtransaction", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil

}

//SendTransaction 发送交易
func (wm *WalletManager) SendTransaction(walletID, to string, amount decimal.Decimal, password string, feesInSender bool) ([]string, error) {

	var (
		usedUTXO   []*Unspent
		balance    = decimal.New(0, 0)
		totalSend  = amount
		actualFees = decimal.New(0, 0)
		sendTime   = 1
		txIDs      = make([]string, 0)
	)

	utxos, err := wm.ListUnspentFromLocalDB(walletID)
	if err != nil {
		return nil, err
	}

	//获取utxo，按小到大排序
	sort.Sort(UnspentSort{utxos, func(a, b *Unspent) int {

		if a.Amount > b.Amount {
			return 1
		} else {
			return -1
		}
	}})

	//读取钱包
	w, err := wm.GetWalletInfo(walletID)
	if err != nil {
		return nil, err
	}

	totalBalance, _ := decimal.NewFromString(wm.GetWalletBalance(w.WalletID))
	if totalBalance.LessThanOrEqual(amount) && feesInSender {
		return nil, errors.New("The wallet's balance is not enough!")
	} else if totalBalance.LessThan(amount) && !feesInSender {
		return nil, errors.New("The wallet's balance is not enough!")
	}

	//加载钱包
	key, err := w.HDKey(password)
	if err != nil {
		return nil, err
	}

	//解锁钱包
	err = wm.UnlockWallet(password, 120)
	if err != nil {
		return nil, err
	}

	//创建找零地址
	changeAddr, err := wm.CreateChangeAddress(walletID, key)
	if err != nil {
		return nil, err
	}

	feesRate, err := wm.EstimateFeeRate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Calculating wallet unspent record to build transaction...\n")

	//循环的计算余额是否足够支付发送数额+手续费
	for {

		usedUTXO = make([]*Unspent, 0)
		balance = decimal.New(0, 0)

		//计算一个可用于支付的余额
		for _, u := range utxos {

			if u.Spendable {
				ua, _ := decimal.NewFromString(u.Amount)
				balance = balance.Add(ua)
				usedUTXO = append(usedUTXO, u)
				if balance.GreaterThanOrEqual(totalSend) {
					break
				}
			}
		}

		if balance.LessThan(totalSend) {
			return nil, errors.New("The balance is not enough!")
		}

		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
		fees, err := wm.EstimateFee(int64(len(usedUTXO)), 2, feesRate)
		if err != nil {
			return nil, err
		}

		//如果要手续费有发送支付，得计算加入手续费后，计算余额是否足够
		if feesInSender {
			//总共要发送的
			totalSend = amount.Add(fees)
			if totalSend.GreaterThan(balance) {
				continue
			}
			totalSend = amount
		} else {

			if fees.GreaterThanOrEqual(amount) {
				return nil, errors.New("The sent amount is not enough for fees!")
			}

			totalSend = amount.Sub(fees)
		}

		actualFees = fees

		break

	}

	changeAmount := balance.Sub(totalSend).Sub(actualFees)

	fmt.Printf("-----------------------------------------------\n")
	fmt.Printf("From WalletID: %s\n", walletID)
	fmt.Printf("To Address: %s\n", to)
	fmt.Printf("Use: %v\n", balance.StringFixed(8))
	fmt.Printf("Fees: %v\n", actualFees.StringFixed(8))
	fmt.Printf("Receive: %v\n", totalSend.StringFixed(8))
	fmt.Printf("Change: %v\n", changeAmount.StringFixed(8))
	fmt.Printf("-----------------------------------------------\n")

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if len(usedUTXO) > wm.Config.MaxTxInputs {
		sendTime = int(math.Ceil(float64(len(usedUTXO)) / float64(wm.Config.MaxTxInputs)))
	}

	for i := 0; i < sendTime; i++ {

		var sendUxto []*Unspent
		var pieceOfSend = decimal.New(0, 0)

		s := i * wm.Config.MaxTxInputs

		//最后一个，计算余数
		if i == sendTime-1 {
			sendUxto = usedUTXO[s:]

			pieceOfSend = totalSend
		} else {
			sendUxto = usedUTXO[s : s+wm.Config.MaxTxInputs]

			for _, u := range sendUxto {
				ua, _ := decimal.NewFromString(u.Amount)
				pieceOfSend = pieceOfSend.Add(ua)
			}

		}

		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
		piecefees, err := wm.EstimateFee(int64(len(sendUxto)), 2, feesRate)

		if piecefees.LessThan(decimal.NewFromFloat(0.00001)) {
			piecefees = decimal.NewFromFloat(0.00001)
		}
		if err != nil {
			return nil, err
		}

		//解锁钱包
		err = wm.UnlockWallet(password, 120)
		if err != nil {
			return nil, err
		}

		//log.Std.Info("pieceOfSend = %s \n", pieceOfSend.StringFixed(8))
		//log.Std.Info("piecefees = %s \n", piecefees.StringFixed(8))
		//log.Std.Info("feesRate = %s \n", feesRate.StringFixed(8))

		//创建交易
		txRaw, _, err := wm.BuildTransaction(sendUxto, []string{to}, changeAddr.Address, []decimal.Decimal{pieceOfSend}, piecefees)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Build Transaction Successfully\n")

		//签名交易
		signedHex, err := wm.SignRawTransactionInCoreWallet(txRaw, walletID, key, sendUxto)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Sign Transaction Successfully\n")

		txid, err := wm.SendRawTransaction(signedHex)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Submit Transaction Successfully\n")

		txIDs = append(txIDs, txid)
		//txIDs = append(txIDs, signedHex)

		//减去已发送的
		totalSend = totalSend.Sub(pieceOfSend)

	}

	//发送成功后，删除已使用的UTXO
	wm.clearUnspends(usedUTXO, w)

	wm.LockWallet()

	return txIDs, nil
}

//SendBatchTransaction 发送批量交易
func (wm *WalletManager) SendBatchTransaction(walletID string, to []string, amounts []decimal.Decimal, password string) (string, error) {

	var (
		usedUTXO []*Unspent
		balance  = decimal.New(0, 0)
		//totalSend  = amounts
		totalSend  = decimal.New(0, 0)
		actualFees = decimal.New(0, 0)
		//sendTime   = 1
		//txIDs      = make([]string, 0)
	)

	if len(to) == 0 {
		return "", errors.New("Receiver addresses is empty!")
	}

	if len(to) != len(amounts) {
		return "", errors.New("Receiver addresses count is not equal amount count!")
	}

	//计算总发送金额
	for _, amount := range amounts {
		totalSend = totalSend.Add(amount)
	}

	utxos, err := wm.ListUnspentFromLocalDB(walletID)
	if err != nil {
		return "", err
	}

	//获取utxo，按小到大排序
	sort.Sort(UnspentSort{utxos, func(a, b *Unspent) int {

		if a.Amount > b.Amount {
			return 1
		} else {
			return -1
		}
	}})

	//读取钱包
	w, err := wm.GetWalletInfo(walletID)
	if err != nil {
		return "", err
	}

	totalBalance, _ := decimal.NewFromString(wm.GetWalletBalance(w.WalletID))
	if totalBalance.LessThanOrEqual(totalSend) {
		return "", errors.New("The wallet's balance is not enough!")
	}

	//加载钱包
	key, err := w.HDKey(password)
	if err != nil {
		return "", err
	}

	//解锁钱包
	err = wm.UnlockWallet(password, 120)
	if err != nil {
		return "", err
	}

	//创建找零地址
	changeAddr, err := wm.CreateChangeAddress(walletID, key)
	if err != nil {
		return "", err
	}

	feesRate, err := wm.EstimateFeeRate()
	if err != nil {
		return "", err
	}

	fmt.Printf("Calculating wallet unspent record to build transaction...\n")
	computeTotalSend := totalSend
	//循环的计算余额是否足够支付发送数额+手续费
	for {

		usedUTXO = make([]*Unspent, 0)
		balance = decimal.New(0, 0)

		//计算一个可用于支付的余额
		for _, u := range utxos {

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
			return "", errors.New("The balance is not enough!")
		}

		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
		fees, err := wm.EstimateFee(int64(len(usedUTXO)), int64(len(to)+1), feesRate)
		if err != nil {
			return "", err
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
	if len(usedUTXO) > wm.Config.MaxTxInputs {
		errStr := fmt.Sprintf("The transaction is use max inputs over: %d", wm.Config.MaxTxInputs)
		return "", errors.New(errStr)
	}

	changeAmount := balance.Sub(computeTotalSend).Sub(actualFees)

	fmt.Printf("-----------------------------------------------\n")
	fmt.Printf("From WalletID: %s\n", walletID)
	fmt.Printf("To Address: %s\n", strings.Join(to, ", "))
	fmt.Printf("Use: %v\n", balance.StringFixed(8))
	fmt.Printf("Fees: %v\n", actualFees.StringFixed(8))
	fmt.Printf("Receive: %v\n", computeTotalSend.StringFixed(8))
	fmt.Printf("Change: %v\n", changeAmount.StringFixed(8))
	fmt.Printf("-----------------------------------------------\n")

	//解锁钱包
	err = wm.UnlockWallet(password, 120)
	if err != nil {
		return "", err
	}

	//log.Std.Info("pieceOfSend = %s \n", pieceOfSend.StringFixed(8))
	//log.Std.Info("piecefees = %s \n", piecefees.StringFixed(8))
	//log.Std.Info("feesRate = %s \n", feesRate.StringFixed(8))

	//创建交易
	txRaw, _, err := wm.BuildTransaction(usedUTXO, to, changeAddr.Address, amounts, actualFees)
	if err != nil {
		return "", err
	}

	fmt.Printf("Build Transaction Successfully\n")

	//签名交易
	signedHex, err := wm.SignRawTransactionInCoreWallet(txRaw, walletID, key, usedUTXO)
	if err != nil {
		return "", err
	}

	fmt.Printf("Sign Transaction Successfully\n")

	txid, err := wm.SendRawTransaction(signedHex)
	if err != nil {
		return "", err
	}

	fmt.Printf("Submit Transaction Successfully\n")

	//发送成功后，删除已使用的UTXO
	wm.clearUnspends(usedUTXO, w)

	wm.LockWallet()

	return txid, nil
}

//CreateChangeAddress 创建找零地址
func (wm *WalletManager) CreateChangeAddress(walletID string, key *hdkeystore.HDKey) (*openwallet.Address, error) {

	//生产通道
	producer := make(chan []*openwallet.Address)
	defer close(producer)

	go wm.createAddressWork(key, producer, walletID, uint64(time.Now().Unix()), 0, 1)

	//回收创建的地址
	getAddrs := <-producer

	if len(getAddrs) == 0 {
		return nil, errors.New("Change address creation failed!")
	}

	//批量写入数据库
	wallet, err := wm.GetWalletInfo(walletID)
	if err != nil {
		return nil, err
	}

	err = wm.saveAddressToDB(getAddrs, wallet)
	if err != nil {
		return nil, err
	}

	return getAddrs[0], nil
}

//EstimateFee 预估手续费
func (wm *WalletManager) EstimateFee(inputs, outputs int64, feeRate decimal.Decimal) (decimal.Decimal, error) {

	var piece int64 = 1

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if inputs > int64(wm.Config.MaxTxInputs) {
		piece = int64(math.Ceil(float64(inputs) / float64(wm.Config.MaxTxInputs)))
	}

	//计算公式如下：148 * 输入数额 + 34 * 输出数额 + 10
	trx_bytes := decimal.New(inputs*148+outputs*34+piece*10, 0)
	trx_fee := trx_bytes.Div(decimal.New(1000, 0)).Mul(feeRate)

	return trx_fee, nil
}

//EstimateFeeRate 预估的没KB手续费率
func (wm *WalletManager) EstimateFeeRate() (decimal.Decimal, error) {

	defaultRate, _ := decimal.NewFromString("0.0001")

	//估算交易大小 手续费
	request := []interface{}{
		2,
	}

	result, err := wm.WalletClient.Call("estimatefee", request)
	if err != nil {
		return decimal.New(0, 0), err
	}

	feeRate, _ := decimal.NewFromString(result.String())

	if feeRate.LessThan(defaultRate) {
		feeRate = defaultRate
	}

	return feeRate, nil
}

//SummaryWallets 执行汇总流程
func (wm *WalletManager) SummaryWallets() {

	log.Std.Info("[Summary Wallet Start]------%s", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for wid, wallet := range wm.WalletsInSum {

		//重新加载utxo
		wm.RebuildWalletUnspent(wid)

		//统计钱包最新余额
		wb := wm.GetWalletBalance(wid)

		balance, _ := decimal.NewFromString(wb)
		//如果余额大于阀值，汇总的地址
		if balance.GreaterThan(wm.Config.Threshold) {

			log.Std.Info("Summary account[%s]balance = %v ", wallet.WalletID, balance)
			log.Std.Info("Summary account[%s]Start Send Transaction", wallet.WalletID)

			txID, err := wm.SendTransaction(wallet.WalletID, wm.Config.SumAddress, balance, wallet.Password, false)
			if err != nil {
				log.Std.Info("Summary account[%s]unexpected error: %v", wallet.WalletID, err)
				continue
			} else {
				log.Std.Info("Summary account[%s]successfully，Received Address[%s], TXID：%s", wallet.WalletID, wm.Config.SumAddress, txID)
			}
		} else {
			log.Std.Info("Wallet Account[%s]-[%s]Current Balance: %v，below threshold: %v", wallet.Alias, wallet.WalletID, balance, wm.Config.Threshold)
		}
	}

	log.Std.Info("[Summary Wallet end]------%s", common.TimeFormat("2006-01-02 15:04:05"))
}

//AddWalletInSummary 添加汇总钱包账户
func (wm *WalletManager) AddWalletInSummary(wid string, wallet *openwallet.Wallet) {
	wm.WalletsInSum[wid] = wallet
}

//clearUnspends 清楚已使用的UTXO
func (wm *WalletManager) clearUnspends(utxos []*Unspent, wallet *openwallet.Wallet) {
	db, err := wallet.OpenDB()
	if err != nil {
		return
	}
	defer db.Close()

	//开始事务
	tx, err := db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	for _, utxo := range utxos {
		tx.DeleteStruct(utxo)
	}

	tx.Commit()
}

//createAddressWork 创建地址过程
func (wm *WalletManager) createAddressWork(k *hdkeystore.HDKey, producer chan<- []*openwallet.Address, walletID string, index, start, end uint64) {

	runAddress := make([]*openwallet.Address, 0)
	runWIFs := make([]string, 0)

	derivedPath := fmt.Sprintf("%s/%d", k.RootPath, index)
	childKey, err := k.DerivedKeyWithPath(derivedPath, wm.Config.CurveType)
	if err != nil {
		producer <- make([]*openwallet.Address, 0)
		return
	}

	for i := start; i < end; i++ {
		// 生成地址
		wif, address, errRun := wm.CreateNewPrivateKey(k.KeyID, childKey, derivedPath, i)
		if errRun != nil {
			log.Std.Info("Create new privKey failed unexpected error: %v", errRun)
			continue
		}

		////导入私钥
		//errRun = ImportPrivKey(wif, alias)
		//if errRun != nil {
		//	log.Std.Info("Import privKey failed unexpected error: %v", errRun)
		//	continue
		//}

		runAddress = append(runAddress, address)
		runWIFs = append(runWIFs, wif)
	}

	//批量导入私钥
	failed, errRun := wm.ImportMulti(runAddress, runWIFs, true)
	if errRun != nil {
		producer <- make([]*openwallet.Address, 0)
		return
	}

	//删除导入失败的
	for _, fi := range failed {
		runAddress = append(runAddress[:fi], runAddress[fi+1:]...)
	}

	//生成完成
	producer <- runAddress
}

//generateSeed 创建种子
func (wm *WalletManager) GenerateSeed() []byte {
	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		return nil
	}

	return seed
}

//exportAddressToFile 导出地址到文件中
func (wm *WalletManager) exportAddressToFile(addrs []*openwallet.Address, filePath string) {

	var (
		content string
	)

	for _, a := range addrs {

		log.Std.Info("Export: %s ", a.Address)

		content = content + a.Address + "\n"
	}

	file.MkdirAll(wm.Config.addressDir)
	file.WriteFile(filePath, []byte(content), true)
}

//saveAddressToDB 保存地址到数据库
func (wm *WalletManager) saveAddressToDB(addrs []*openwallet.Address, wallet *openwallet.Wallet) error {
	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, a := range addrs {
		err = tx.Save(a)
		if err != nil {
			continue
		}
	}

	return tx.Commit()

}

//loadConfig 读取配置
func (wm *WalletManager) LoadConfig() error {

	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(wm.Config.configFilePath, wm.Config.configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd Config -s <symbol>' ")
	}

	wm.Config.ServerAPI = c.String("serverAPI")
	wm.Config.Threshold, _ = decimal.NewFromString(c.String("threshold"))
	wm.Config.SumAddress = c.String("sumAddress")
	wm.Config.RpcUser = c.String("rpcUser")
	wm.Config.RpcPassword = c.String("rpcPassword")
	wm.Config.NodeInstallPath = c.String("nodeInstallPath")
	wm.Config.IsTestNet, _ = c.Bool("isTestNet")
	wm.Config.WalletPassword = c.String("walletPassword")
	if wm.Config.IsTestNet {
		wm.Config.WalletDataPath = c.String("testNetDataPath")
	} else {
		wm.Config.WalletDataPath = c.String("mainNetDataPath")
	}

	cyclesec := c.String("cycleSeconds")
	if cyclesec == "" {
		return errors.New(fmt.Sprintf(" cycleSeconds is not set, sample: 1m , 30s, 3m20s etc... Please set it in './conf/%s.ini' \n", Symbol))
	}

	wm.Config.CycleSeconds, _ = time.ParseDuration(cyclesec)

	token := BasicAuth(wm.Config.RpcUser, wm.Config.RpcPassword)

	wm.WalletClient = NewClient(wm.Config.ServerAPI, token, false)

	return nil
}

//打印钱包列表
func (wm *WalletManager) printWalletList(list []*openwallet.Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		a := w.SingleAssetsAccount(wm.Config.Symbol)
		a.Balance = wm.GetWalletBalance(a.AccountID)
		tableInfo = append(tableInfo, []interface{}{
			i, a.WalletID, a.Alias, a.Balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}

//startNode 开启节点
func (wm *WalletManager) startNode() error {

	//读取配置
	absFile := filepath.Join(wm.Config.configFilePath, wm.Config.configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup! ")
	}

	startNodeCMD := c.String("startNodeCMD")
	return wm.cmdCall(startNodeCMD, false)
}

//stopNode 关闭节点
func (wm *WalletManager) stopNode() error {
	//读取配置
	absFile := filepath.Join(wm.Config.configFilePath, wm.Config.configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup! ")
	}

	stopNodeCMD := c.String("stopNodeCMD")
	return wm.cmdCall(stopNodeCMD, true)
}

//cmdCall 执行命令
func (wm *WalletManager) cmdCall(cmd string, wait bool) error {

	var (
		cmdName string
		args    []string
	)

	cmds := strings.Split(cmd, " ")
	if len(cmds) > 0 {
		cmdName = cmds[0]
		args = cmds[1:]
	} else {
		return errors.New("command not found ")
	}
	session := sh.Command(cmdName, args)
	if wait {
		return session.Run()
	} else {
		return session.Start()
	}
}
