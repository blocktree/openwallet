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
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/openwallet/accounts/keystore"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	storage *keystore.HDKeystore
)

func init() {
	storage = keystore.NewHDKeystore(keyDir, MasterKey, keystore.StandardScryptN, keystore.StandardScryptP)
}

func GetAddressesByAccount(walletID string) ([]string, error) {

	var (
		addresses = make([]string, 0)
	)

	request := []interface{}{
		walletID,
	}

	result, err := client.Call("getaddressesbyaccount", request)
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
func GetAddressesFromLocalDB(walletID string) ([]*Address, error) {

	db, err := storm.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var addresses []*Address
	err = db.Find("Account", walletID, &addresses)
	if err != nil {
		return nil, err
	}

	return addresses, nil

}

//ImportPrivKey 导入私钥
func ImportPrivKey(wif, walletID string) error {

	request := []interface{}{
		wif,
		walletID,
		false,
	}

	_, err := client.Call("importprivkey", request)

	if err != nil {
		return err
	}

	return err

}

//ImportMulti 批量导入地址和私钥
func ImportMulti(addresses []*Address, keys []string, walletID string, watchOnly bool) ([]int, error) {

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

	if len(addresses) != len(keys) {
		return nil, errors.New("Import addresses is not equal keys count!")
	}

	for i, a := range addresses {
		k := keys[i]
		obj := map[string]interface{}{
			"scriptPubKey": map[string]interface{}{
				"address": a.Address,
			},
			"label":     walletID,
			"timestamp": "now",
			"watchonly": watchOnly,
		}

		if !watchOnly {
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

	result, err := client.Call("importmulti", request)
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
func GetCoreWalletinfo() error {

	_, err := client.Call("getwalletinfo", nil)

	if err != nil {
		return err
	}

	return err

}

//UnlockWallet 解锁钱包
func UnlockWallet(passphrase string, seconds int) error {

	request := []interface{}{
		passphrase,
		seconds,
	}

	_, err := client.Call("walletpassphrase", request)
	if err != nil {
		return err
	}

	return err

}

//KeyPoolRefill 重新填充私钥池
func KeyPoolRefill(keyPoolSize uint64) error {

	request := []interface{}{
		keyPoolSize,
	}

	_, err := client.Call("keypoolrefill", request)
	if err != nil {
		return err
	}

	return nil
}

//CreateReceiverAddress 给指定账户创建地址
func CreateReceiverAddress(account string) (string, error) {

	request := []interface{}{
		account,
	}

	result, err := client.Call("getnewaddress", request)
	if err != nil {
		return "", err
	}

	return result.String(), err

}

//CreateBatchAddress 批量创建地址
func CreateBatchAddress(name, password string, count uint64) (string, error) {

	var (
		synCount   uint64 = 20
		quit              = make(chan struct{})
		done              = 0 //完成标记
		shouldDone        = 0 //需要完成的总数
	)

	//读取钱包
	w, err := GetWalletInfo(name)
	if err != nil {
		return "", err
	}

	//加载钱包
	key, err := w.HDKey(password)
	if err != nil {
		return "", err
	}

	timestamp := time.Now()
	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405", timestamp) + ".txt"
	filePath := filepath.Join(addressDir, filename)

	//生产通道
	producer := make(chan []*Address)
	defer close(producer)

	//消费通道
	worker := make(chan []*Address)
	defer close(worker)

	//保存地址过程
	saveAddressWork := func(addresses chan []*Address, filename string) {

		var (
			saveErr error
		)

		for {
			//回收创建的地址
			getAddrs := <-addresses

			//批量写入数据库
			saveErr = saveAddressToDB(getAddrs)
			//数据保存成功才导出文件
			if saveErr == nil {
				//导出一批地址
				exportAddressToFile(getAddrs, filename)
			}

			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//解锁钱包
	err = UnlockWallet(password, 120)
	if err != nil {
		return "", err
	}

	/*	开启导出的线程，监听新地址，批量导出	*/

	go saveAddressWork(worker, filePath)

	/*	计算synCount个线程，内部运行的次数	*/

	//每个线程内循环的数量，以synCount个线程并行处理
	runCount := count / synCount
	otherCount := count % synCount

	if runCount > 0 {

		for i := uint64(0); i < synCount; i++ {

			//开始创建地址
			log.Printf("Start create address thread[%d]\n", i)
			s := i * runCount
			e := (i + 1) * runCount
			go createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		log.Printf("Start create address thread[REST]\n")
		s := count - otherCount
		e := count
		go createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)

		shouldDone++
	}

	values := make([][]*Address, 0)

	//以下使用生产消费模式

	for {

		var activeWorker chan<- []*Address
		var activeValue []*Address

		//当数据队列有数据时，释放顶部，激活消费
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
			//log.Printf("completed %d", len(pa))
			//当激活消费者后，传输数据给消费者，并把顶部数据出队
		case activeWorker <- activeValue:
			//log.Printf("Get %d", len(activeValue))
			values = values[1:]

		case <-quit:
			//退出
			log.Printf("All addresses have been created!")
			return filePath, nil
		}
	}

	return filePath, nil
}

//CreateNewWallet 创建钱包
func CreateNewWallet(name, password string) error {

	var (
		err     error
		wallets []*Wallet
	)

	//检查钱包名是否存在
	wallets, err = GetWalletKeys(keyDir)
	for _, w := range wallets {
		if w.Alias == name {
			return errors.New("The wallet's alias is duplicated!")
		}
	}

	fmt.Printf("Verify password in bitcoin-core wallet...\n")

	err = EncryptWallet(password)
	if err != nil {
		//钱包已经加密，解锁钱包1秒，检查密码
		err = UnlockWallet(password, 1)
		if err != nil {
			return errors.New("The wallet's password is not equal bitcoin-core wallet!\n")
		}
	}

	fmt.Printf("Create new wallet keystore...\n")

	keyFile, err := keystore.StoreHDKey(keyDir, MasterKey, name, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return err
	}

	fmt.Printf("Wallet create successfully, key path: %s\n", keyFile)

	return nil
}

//EncryptWallet 通过密码加密钱包，只在第一次加密码时才有效
func EncryptWallet(password string) error {

	request := []interface{}{
		password,
	}

	_, err := client.Call("encryptwallet", request)
	if err != nil {
		return err
	}
	return nil
}

//GetWalletKeys 通过给定的文件路径加载keystore文件得到钱包列表
func GetWalletKeys(dir string) ([]*Wallet, error) {

	var (
		buf = new(bufio.Reader)
		key struct {
			Alias  string `json:"alias"`
			RootId string `json:"rootid"`
		}

		wallets = make([]*Wallet, 0)
	)

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
	files, err := ioutil.ReadDir(dir)
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
		wallets = append(wallets, w)

	}

	return wallets, nil

}

//GetWalleList 获取钱包列表
func GetWalleList() ([]*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		balance, _ := GetWalletBalance(w.WalletID)
		w.Balance = balance
	}

	return wallets, nil
}

//GetWalletInfo 获取钱包列表
func GetWalletInfo(walletID string) (*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			balance, _ := GetWalletBalance(w.WalletID)
			w.Balance = balance

			return w, nil
		}

	}

	return nil, errors.New("The wallet that your given name is not exist!")
}

//GetWalletBalance 获取钱包余额
func GetWalletBalance(name string) (string, error) {

	request := []interface{}{
		name,
		1,
		true,
	}

	balance, err := client.Call("getbalance", request)
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

//CreateNewPrivateKey 创建私钥，返回私钥wif格式字符串
func CreateNewPrivateKey(key *keystore.HDKey, start, index uint64) (string, *Address, error) {

	derivedPath := fmt.Sprintf("%s/%d/%d", key.RootPath, start, index)
	//fmt.Printf("derivedPath = %s\n", derivedPath)
	childKey, err := key.DerivedKeyWithPath(derivedPath)

	privateKey, err := childKey.ECPrivKey()
	if err != nil {
		return "", nil, err
	}

	cfg := chaincfg.MainNetParams
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
	}

	addr := Address{
		Address:   address.String(),
		Account:   key.RootId,
		HDPath:    derivedPath,
		CreatedAt: time.Now(),
	}

	return wif.String(), &addr, err
}

//CreateBatchPrivateKey
func CreateBatchPrivateKey(key *keystore.HDKey, count uint64) ([]string, error) {

	var (
		wifs = make([]string, 0)
	)

	start := time.Now().Unix()
	for i := uint64(0); i < count; i++ {
		wif, _, err := CreateNewPrivateKey(key, uint64(start), i)
		if err != nil {
			continue
		}
		wifs = append(wifs, wif)
	}

	return wifs, nil

}

//BackupWallet 备份钱包
func BackupWallet(dest string) error {

	request := []interface{}{
		dest,
	}

	_, err := client.Call("backupwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//DumpWallet 导出钱包所有私钥文件
func DumpWallet(filename string) error {

	request := []interface{}{
		filename,
	}

	_, err := client.Call("dumpwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//ImportWallet 导入钱包私钥文件
func ImportWallet(filename string) error {

	request := []interface{}{
		filename,
	}

	_, err := client.Call("importwallet", request)
	if err != nil {
		return err
	}

	return nil

}

//GetBlockChainInfo 获取钱包区块链信息
func GetBlockChainInfo() (*BlockchainInfo, error) {

	result, err := client.Call("getblockchaininfo", nil)
	if err != nil {
		return nil, err
	}

	blockchain := NewBlockchainInfo(result)

	return blockchain, nil

}

//ListUnspent 获取未花记录
func ListUnspent(min uint64) ([]*Unspent, error) {

	var (
		utxos = make([]*Unspent, 0)
	)

	request := []interface{}{
		min,
	}

	result, err := client.Call("listunspent", request)
	if err != nil {
		return nil, err
	}

	array := result.Array()
	for _, a := range array {
		utxos = append(utxos, NewUnspent(&a))
	}

	return utxos, nil

}

//BatchInsertWalletUnspent 批量插入未花记录到本地
func BatchInsertWalletUnspent() error {

	//查找核心钱包确认数大于1的
	utxos, err := ListUnspent(1)
	if err != nil {
		return err
	}

	db, err := storm.Open(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	//开始事务
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//批量插入到本地数据库
	//设置utxo的钱包账户
	for _, utxo := range utxos {
		var addr Address
		err = db.One("Address", utxo.Address, &addr)
		utxo.Account = addr.Account
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
func ListUnspentFromLocalDB(walletID string) ([]*Unspent, error) {

	db, err := storm.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var utxos []*Unspent
	err = db.Find("Account", walletID, &utxos)
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

//BuildTransaction 构建交易单
func BuildTransaction(utxos []*Unspent, to, change string, amount, fees decimal.Decimal) (string, decimal.Decimal, error) {

	var (
		inputs      = make([]interface{}, 0)
		outputs     = make(map[string]interface{})
		totalAmount = decimal.New(0, 0)
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

	if totalAmount.LessThan(amount) {
		return "", decimal.New(0, 0), errors.New("The balance is not enough!")
	}

	changeAmount := totalAmount.Sub(amount).Sub(fees)
	if changeAmount.GreaterThan(decimal.New(0, 0)) {
		//ca, _ := changeAmount.Float64()
		outputs[change] = changeAmount.StringFixed(8)

		fmt.Printf("Create change address for receiving %s coin.", outputs[change])
	}

	//ta, _ := amount.Float64()
	outputs[to] = amount.StringFixed(8)

	request := []interface{}{
		inputs,
		outputs,
	}

	rawTx, err := client.Call("createrawtransaction", request)
	if err != nil {
		return "", decimal.New(0, 0), err
	}

	return rawTx.String(), changeAmount, nil
}

//SignRawTransaction 钱包交易单
func SignRawTransaction(txHex, walletID, password string, utxos []*Unspent) (string, error) {

	var (
		keys = make([]string, 0)
	)

	//读取钱包
	w, err := GetWalletInfo(walletID)
	if err != nil {
		return "", err
	}

	//加载钱包
	hdkey, err := w.HDKey(password)
	if err != nil {
		return "", err
	}

	//查找未花签名需要的私钥
	for _, u := range utxos {

		childKey, err := hdkey.DerivedKeyWithPath(u.HDAddress.HDPath)

		privateKey, err := childKey.ECPrivKey()
		if err != nil {
			return "", err
		}

		cfg := chaincfg.MainNetParams
		if isTestNet {
			cfg = chaincfg.TestNet3Params
		}

		wif, err := btcutil.NewWIF(privateKey, &cfg, true)
		if err != nil {
			return "", err
		}

		keys = append(keys, wif.String())

	}

	request := []interface{}{
		txHex,
		utxos,
		keys,
	}

	result, err := client.Call("signrawtransaction", request)
	if err != nil {
		return "", err
	}

	return result.Get("hex").String(), nil

}

//SendRawTransaction 广播交易
func SendRawTransaction(txHex string) (string, error) {

	request := []interface{}{
		txHex,
	}

	result, err := client.Call("sendrawtransaction", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil

}

//SendTransaction 发送交易
func SendTransaction(walletID, to string, amount decimal.Decimal, password string, feesInSender bool) (string, error) {

	var (
		usedUTXO  []*Unspent
		balance   = decimal.New(0, 0)
		totalSend = amount
		actualFees = decimal.New(0, 0)
	)

	utxos, err := ListUnspentFromLocalDB(walletID)
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

	fmt.Printf("Calculating wallet unspent record to build transaction...")

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

		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
		fees, err := EstimateFee(int64(len(usedUTXO)), 2)
		if err != nil {
			return "", err
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
			totalSend = amount.Sub(fees)
		}

		actualFees = fees

		break

	}

	//解锁钱包
	err = UnlockWallet(password, 120)
	if err != nil {
		return "", err
	}

	//创建找零地址
	changeAddr, err := CreateChangeAddress(walletID, password)
	if err != nil {
		return "", err
	}

	//创建交易
	txRaw, changeAmount, err := BuildTransaction(usedUTXO, to, changeAddr.Address, totalSend, actualFees)
	if err != nil {
		return "", err
	}

	fmt.Printf("Build Transaction Successfully\n")

	fmt.Printf("-----------------------------------------------\n")
	fmt.Printf("From WalletID: %s\n", walletID)
	fmt.Printf("To Address: %s\n", to)
	fmt.Printf("Use: %v\n", balance.StringFixed(8))
	fmt.Printf("Fees: %v\n", actualFees.StringFixed(8))
	fmt.Printf("Receive: %v\n", totalSend.StringFixed(8))
	fmt.Printf("Change: %v\n", changeAmount.StringFixed(8))
	fmt.Printf("-----------------------------------------------\n")

	//签名交易
	signedHex, err := SignRawTransaction(txRaw, walletID, password, utxos)
	if err != nil {
		return "", err
	}

	fmt.Printf("Sign Transaction Successfully\n")

	txid, err := SendRawTransaction(signedHex)
	if err != nil {
		return "", err
	}

	fmt.Printf("Submit Transaction Successfully\n")

	return txid, nil
}

//CreateChangeAddress 创建找零地址
func CreateChangeAddress(walletID, password string) (*Address, error) {

	//读取钱包
	w, err := GetWalletInfo(walletID)
	if err != nil {
		return nil, err
	}

	//加载钱包
	key, err := w.HDKey(password)
	if err != nil {
		return nil, err
	}
	//生产通道
	producer := make(chan []*Address)
	defer close(producer)

	go createAddressWork(key, producer, walletID, uint64(time.Now().Unix()), 0, 1)

	//回收创建的地址
	getAddrs := <-producer

	if len(getAddrs) == 0 {
		return nil, errors.New("Change address creation failed!")
	}

	//批量写入数据库
	err = saveAddressToDB(getAddrs)
	if err != nil {
		return nil, err
	}

	return getAddrs[0], nil
}

//预估手续费
func EstimateFee(inputs, outputs int64) (decimal.Decimal, error) {

	feeRate, _ := decimal.NewFromString("0.0001")

	//估算交易大小 手续费
	request := []interface{}{
		2,
	}

	result, err := client.Call("estimatefee", request)
	if err != nil {
		return decimal.New(0, 0), err
	}

	feeRate, _ = decimal.NewFromString(result.String())

	trx_bytes := decimal.New(inputs*148+outputs*34+10, 0)
	trx_fee := trx_bytes.Div(decimal.New(1000, 0)).Mul(feeRate)

	return trx_fee, nil
}

//createAddressWork 创建地址过程
func createAddressWork(k *keystore.HDKey, producer chan<- []*Address, walletID string, index, start, end uint64) {

	runAddress := make([]*Address, 0)
	runWIFs := make([]string, 0)

	for i := start; i < end; i++ {
		// 生成地址
		wif, address, errRun := CreateNewPrivateKey(k, index, i)
		if errRun != nil {
			log.Printf("Create new privKey failed unexpected error: %v\n", errRun)
			continue
		}

		////导入私钥
		//errRun = ImportPrivKey(wif, alias)
		//if errRun != nil {
		//	log.Printf("Import privKey failed unexpected error: %v\n", errRun)
		//	continue
		//}

		runAddress = append(runAddress, address)
		runWIFs = append(runWIFs, wif)
	}

	//批量导入私钥
	failed, errRun := ImportMulti(runAddress, runWIFs, walletID, CoreWalletWatchOnly)
	if errRun != nil {
		producer <- make([]*Address, 0)
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
func generateSeed() []byte {
	seed, err := hdkeychain.GenerateSeed(32)
	if err != nil {
		return nil
	}

	return seed
}

//exportAddressToFile 导出地址到文件中
func exportAddressToFile(addrs []*Address, filePath string) {

	var (
		content string
	)

	for _, a := range addrs {

		log.Printf("Export: %s \n", a)

		content = content + a.Address + "\n"
	}

	file.MkdirAll(addressDir)
	file.WriteFile(filePath, []byte(content), true)
}

//saveAddressToDB 保存地址到数据库
func saveAddressToDB(addrs []*Address) error {
	db, err := storm.Open(dbPath)
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

// skipKeyFile ignores editor backups, hidden files and folders/symlinks.
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
