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

package tezos

import (
	"github.com/shopspring/decimal"
	"time"
	"encoding/hex"
	"golang.org/x/crypto/blake2b"
	"strconv"
	"github.com/tidwall/gjson"
	"path/filepath"
	"github.com/astaxie/beego/config"
	"errors"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/bndr/gotabulate"
	"fmt"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"github.com/blocktree/go-OWCrypt"
)

const (
	maxAddresNum = 10000000
)
var (
	coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000)
)

//地址，公钥，公钥哈希，私钥，签名前缀
var prefix = map[string][]byte{
	"tz1": {6, 161, 159},
	"tz2": {6, 161, 161},
	"edpk": {13, 15, 37, 217},
	"edsk": {43, 246, 78, 7},
	"edsig": {9, 245, 205, 134, 18},
	"nil": {},
}

//消息前缀
var watermark = map[string][]byte{
	"block": {1},
	"endorsement": {2},
	"generic": {3},
}

type WalletManager struct {
	Storage      *hdkeystore.HDKeystore         //秘钥存取
	WalletClient *Client                        // 节点客户端
	Config       *WalletConfig                  //钱包管理配置
	WalletsInSum map[string]*openwallet.Wallet  //参与汇总的钱包
	//Blockscanner *XTZBlockScanner             //区块扫描器
	//Decoder      *openwallet.AddressDecoder     //地址编码器
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey)
	storage := hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.Storage = storage
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	//wm.Blockscanner = NewXTZBlockScanner(&wm)
	//wm.Decoder = AddressDecoder
	return &wm
}

//签名交易
func (wm *WalletManager) signTransaction(hash string, sk []byte, watermark []byte) (string, string, error) {
	bhash,_ := hex.DecodeString(hash)
	merbuf := append(watermark, bhash...)
	ctx, err :=blake2b.New(32,nil)
	if err != nil {
		return "", "", err
	}
	ctx.Write(merbuf[:])
	bb := ctx.Sum(nil)

	sig, _ := owcrypt.Signature(sk[:], nil, 0, bb[:], uint16(len(bb)), owcrypt.ECC_CURVE_ED25519_EXTEND)
	edsig := base58checkEncode(sig, prefix["edsig"])

	sbyte := hash + hex.EncodeToString(sig[:])

	return edsig, sbyte, nil
}

//判断该key是否需要reverl
func (wm *WalletManager) isReverlKey(pubkey string) bool {
	manager_key := wm.WalletClient.CallGetManagerKey(pubkey)
	//manager := gjson.GetBytes(ret, "manager")
	key := gjson.GetBytes(manager_key, "key")

	if key.Str == "" {
		return true
	}

	return false
}

//转账
func (wm *WalletManager) Transfer(keys Key, dst string, fee, gas_limit, storage_limit, amount string) (string, string){
	header := wm.WalletClient.CallGetHeader()
	blk_hash := gjson.GetBytes(header, "hash").Str
	chain_id := gjson.GetBytes(header, "chain_id").Str
	protocol := gjson.GetBytes(header, "protocol").Str

	counter := wm.WalletClient.CallGetCounter(keys.Address)
	icounter,_ := strconv.Atoi(string(counter))
	icounter = icounter + 1

	isReverlKey := wm.isReverlKey(keys.Address)

	var ops []interface{}
	reverl := map[string]string{
		"kind": "reveal",
		"fee": fee,
		"public_key": keys.PublicKey,
		"source": keys.Address,
		"gas_limit": gas_limit,
		"storage_limit": storage_limit,
		"counter": strconv.Itoa(icounter),
	}
	if isReverlKey {
		icounter = icounter + 1
		ops = append(ops, reverl)
	}

	transaction := map[string]string{
		"kind" : "transaction",
		"amount" : amount,
		"destination" : dst,
		"fee": fee,
		"gas_limit": gas_limit,
		"storage_limit": storage_limit,
		"counter": strconv.Itoa(icounter),
		"source": keys.Address,
	}

	ops = append(ops, transaction)
	opOb := make(map[string]interface{})
	opOb["branch"] = blk_hash
	opOb["contents"] = ops
	hash := wm.WalletClient.CallForgeOps(chain_id, blk_hash, opOb)

	//sign
	edsig, sbyte, _ := wm.signTransaction(hash, keys.PrivateKey, watermark["generic"])

	//preapply operations
	var opObs []interface{}
	opOb["signature"] = edsig
	opOb["protocol"] = protocol
	opObs = append(opObs, opOb)
	pre := wm.WalletClient.CallPreapplyOps(opObs)

	//jnject aperations
	inj := wm.WalletClient.CallInjectOps(sbyte)
	return string(inj), string(pre)
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

	return w, keyFile, nil
}


//GetWalletKeys 通过给定的文件路径加载keystore文件得到钱包列表
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

func (wm *WalletManager) AddWalletInSummary(wid string, wallet *openwallet.Wallet) {
	wm.WalletsInSum[wid] = wallet
}

//获取钱包余额
func (wm *WalletManager) getWalletBalance(wallet *openwallet.Wallet) (decimal.Decimal, []*openwallet.Address, error) {
	var (
		synCount      int = 10
		quit              = make(chan struct{})
		done              = 0 //完成标记
		shouldDone        = 0 //需要完成的总数
	)

	db, err := wallet.OpenDB()
	if err != nil {
		return decimal.NewFromFloat(0), nil, err
	}
	defer db.Close()

	var addrs []*openwallet.Address
	db.All(&addrs)

	var balance decimal.Decimal = decimal.NewFromFloat(0)
	count := len(addrs)
	if  count <= 0 {
		log.Std.Info("This wallet have 0 address!!!")
		return decimal.NewFromFloat(0), nil, nil
	} else {
		log.Std.Info("wallet %s have %d addresses， please wait minutes to get wallet balance", wallet.Alias, count)
	}

	//生产通道
	producer := make(chan []*openwallet.Address)
	defer close(producer)

	//消费通道
	worker := make(chan []*openwallet.Address)
	defer close(worker)

	//统计余额
	go func(addrs chan []*openwallet.Address) {
		for {
			//
			balances := <-addrs

			for _, b := range(balances) {
				balance = balance.Add(decimal.RequireFromString(b.Balance))
			}

			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	} (worker)

	/*	计算synCount个线程，内部运行的次数	*/
	//每个线程内循环的数量，以synCount个线程并行处理
	runCount := count / synCount
	otherCount := count % synCount

	if runCount > 0 {
		for i := 0; i < synCount; i++ {
			//开始
			//log.Std.Info("Start get balance thread[%d]", i)
			start := i * runCount
			end := (i + 1) * runCount - 1
			as := addrs[start:end]

			go func(producer chan []*openwallet.Address, addrs []*openwallet.Address, wm *WalletManager) {
				var bs []*openwallet.Address
				for _, a := range(addrs) {
					b := wm.WalletClient.CallGetbalance(a.Address)
					a.Balance = string(b)
					bs = append(bs, a)
				}

				producer <- bs
			}(producer, as, wm)

			shouldDone++
		}
	}

	if otherCount > 0 {
		//
		//log.Std.Info("Start get balance thread[REST]")
		start := runCount*synCount
		as := addrs[start:]

		go func(producer chan []*openwallet.Address, addrs []*openwallet.Address, wm *WalletManager) {
			var bs []*openwallet.Address
			for _, a := range(addrs) {
				b := wm.WalletClient.CallGetbalance(a.Address)
				a.Balance = string(b)
				bs = append(bs, a)
			}

			producer <- bs
		}(producer, as, wm)

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
			//当激活消费者后，传输数据给消费者，并把顶部数据出队
			outputAddress = append(outputAddress, pa...)
		case activeWorker <- activeValue:
			values = values[1:]
		case <-quit:
			//退出
			log.Std.Info("wallet %s get all addresses's balance finished", wallet.Alias)
			return balance.Div(coinDecimal), outputAddress, nil
		}
	}

	return balance.Div(coinDecimal), outputAddress, nil
}

//打印钱包列表
func (wm *WalletManager) printWalletList(list []*openwallet.Wallet, getBalance bool) ([][]*openwallet.Address) {
	tableInfo := make([][]interface{}, 0)
	var addrs [][]*openwallet.Address

	for i, w := range list {
		if getBalance {
			balance, addr, _ := wm.getWalletBalance(w)
			tableInfo = append(tableInfo, []interface{}{
				i, w.WalletID, w.Alias, w.DBFile, balance,
			})

			addrs = append(addrs, addr)
			//休眠20秒是因为http请求会导致下一个钱包获取余额API请求失败
			if i != (len(list) - 1) {
				time.Sleep(time.Second * 20)
			}
		} else {
			tableInfo = append(tableInfo, []interface{}{
				i, w.WalletID, w.Alias, w.DBFile,
			})
		}

	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	if getBalance {
		t.SetHeaders([]string{"No.", "ID", "Name", "DBFile", "Balance",})
	} else {
		t.SetHeaders([]string{"No.", "ID", "Name", "DBFile",})
	}

	//打印信息
	fmt.Println(t.Render("simple"))

	return addrs
}

//CreateNewPrivateKey 创建私钥，返回私钥wif格式字符串
func (wm *WalletManager) CreateNewPrivateKey(key *hdkeystore.HDKey, start, index uint64) (*openwallet.Address, error) {
	derivedPath := fmt.Sprintf("%s/%d/%d", key.RootPath, start, index)
	childKey, err := key.DerivedKeyWithPath(derivedPath, wm.Config.CurveType)
	if err != nil {
		return nil, err
	}

	publicKey := childKey.GetPublicKeyBytes()

	cfg := addressEncoder.XTZ_mainnetAddress_tz1
	pkHash := owcrypt.Hash(publicKey, 20, owcrypt.HASH_ALG_BLAKE2B)

	address := addressEncoder.AddressEncode(pkHash, cfg)

	addr := openwallet.Address{
		Address:   address,
		AccountID: key.KeyID,
		HDPath:    derivedPath,
		CreatedAt: time.Now(),
		Symbol:    wm.Config.Symbol,
		Index:     index,
		WatchOnly: false,
	}

	return &addr, err
}

//createAddressWork 创建地址过程
func (wm *WalletManager) createAddressWork(k *hdkeystore.HDKey, producer chan<- []*openwallet.Address, walletID string, index, start, end uint64) {

	runAddress := make([]*openwallet.Address, 0)
	for i := start; i < end; i++ {
		// 生成地址
		address, errRun := wm.CreateNewPrivateKey(k, index, i)
		if errRun != nil {
			log.Std.Info("Create new privKey failed unexpected error: %v", errRun)
			continue
		}

		runAddress = append(runAddress, address)
	}

	//生成完成
	producer <- runAddress
}

func (wm *WalletManager) CreateBatchAddress(walletId, password string, count uint64) (string, []*openwallet.Address, error) {

	var (
		synCount   uint64 = 20
		quit              = make(chan struct{})
		done              = 0 //完成标记
		shouldDone        = 0 //需要完成的总数
	)

	//读取钱包
	w, err := wm.GetWalletByID(walletId)
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
			go wm.createAddressWork(key, producer, walletId, uint64(timestamp.Unix()), s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {
		//开始创建地址
		log.Std.Info("Start create address thread[REST]")
		s := count - otherCount
		e := count
		go wm.createAddressWork(key, producer, walletId, uint64(timestamp.Unix()), s, e)

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
			log.Std.Info("All addresses have been created!")
			return filePath, outputAddress, nil
		}
	}

	return filePath, outputAddress, nil
}

func (wm *WalletManager) summaryWallet(wallet *openwallet.Wallet, password string) error{
	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var addrs []*openwallet.Address
	db.All(&addrs)


	//加载钱包
	key, err := wallet.HDKey(password)
	if err != nil {
		return err
	}

	for _, a := range addrs {
		k, _ := wm.getKeys(key, a)

		//get balance
		decimal_balance, _ := decimal.NewFromString(string(wm.WalletClient.CallGetbalance(a.Address)))

		//判断是否是reveal交易
		fee := wm.Config.MinFee
		isReverl := wm.isReverlKey(a.Address)
		if isReverl {
			//多了reveal操作后，fee * 2
			fee = wm.Config.MinFee.Mul(decimal.RequireFromString("2"))
		}
		// 将该地址多余额减去矿工费后，全部转到汇总地址
		amount := decimal_balance.Sub(fee)
		//该地址预留一点币，否则交易会失败，暂定0.00001 tez
		amount = amount.Sub(decimal.RequireFromString("10"))
		//log.Printf("address:%s banlance:%d amount:%d fee:%d\n", k.Address, decimal_balance.IntPart(), amount.IntPart(), fee.IntPart())

		if decimal_balance.GreaterThan(wm.Config.Threshold) {
			txid, _ := wm.Transfer(*k, wm.Config.SumAddress, strconv.FormatInt(wm.Config.MinFee.IntPart(), 10), strconv.FormatInt(wm.Config.GasLimit.IntPart(), 10),
				strconv.FormatInt(wm.Config.StorageLimit.IntPart(), 10),strconv.FormatInt(amount.IntPart(), 10))

			log.Std.Info("summary form address:%s, to address:%s, amount:%d, txid:%s\n", k.Address, wm.Config.SumAddress, amount.IntPart(), txid)
		}
	}

	return nil
}

//汇总钱包
func (wm *WalletManager) SummaryWallets() {
	log.Std.Info("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for _, wallet := range wm.WalletsInSum {
		wm.summaryWallet(wallet, wallet.Password)
	}

	log.Std.Info("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
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

//GetWalletByID 获取钱包
func (wm *WalletManager) GetWalletByID(walletID string) (*openwallet.Wallet, error) {
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

	wm.Config.ServerAPI = c.String("apiUrl")
	wm.Config.Threshold = (decimal.RequireFromString(c.String("threshold"))).Mul(coinDecimal)
	wm.Config.SumAddress = c.String("sumAddress")
	wm.Config.MinFee = (decimal.RequireFromString(c.String("minFee"))).Mul(coinDecimal)
	wm.Config.GasLimit = (decimal.RequireFromString(c.String("gasLimit"))).Mul(coinDecimal)
	wm.Config.StorageLimit = (decimal.RequireFromString(c.String("storageLimit"))).Mul(coinDecimal)

	wm.WalletClient = NewClient(wm.Config.ServerAPI,false)

	return nil
}


//RestoreWallet 恢复钱包
func (wm *WalletManager) RestoreWallet(keyFile, dbFile, password string) error {

	//根据流程，提供种子文件路径，wallet.db文件的路径，钱包数据库文件的路径。
	//输入钱包密码。
	//复制种子文件到data/btc/key/。
	//复制钱包数据库文件到data/btc/db/。

	var (
		err            error
		key            *hdkeystore.HDKey
		//sleepTime      = 30 * time.Second
	)

	fmt.Printf("Validating key file... \n")

	//检查密码是否可以解析种子文件，是否可以解锁钱包。
	key, err = wm.Storage.GetKey("", keyFile, password)
	if err != nil {
		return fmt.Errorf("Passowrd is incorrect! ")
	}

	fmt.Printf("Restore wallet key and datebase file... \n")

	//复制种子文件到data/btc/key/
	file.MkdirAll(wm.Config.keyDir)
	file.Copy(keyFile, filepath.Join(wm.Config.keyDir, key.FileName()+".key"))

	//复制钱包数据库文件到data/btc/db/
	file.MkdirAll(wm.Config.dbPath)
	file.Copy(dbFile, filepath.Join(wm.Config.dbPath, key.FileName()+".db"))

	fmt.Printf("Backup wallet has been restored. \n")

	fmt.Printf("Finally, you should restart the hcwallet to ensure. \n")

	return nil
}

//通过hdpath获取地址，公钥，私钥
func (wm *WalletManager) getKeys(key *hdkeystore.HDKey, a *openwallet.Address) (*Key, error){
	childKey, err := key.DerivedKeyWithPath(a.HDPath, wm.Config.CurveType)
	if err != nil {
		return nil, err
	}

	prikey, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return  nil, err
	}

	pubkey := childKey.GetPublicKeyBytes()
	//转换成带前缀公钥，交易结构中需要填充此类型公钥
	pk := base58checkEncode(pubkey, prefix["edpk"])
	k := &Key{a.Address,pk,prikey}

	return k, nil
}