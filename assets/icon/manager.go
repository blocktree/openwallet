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

package icon

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"math/big"
	"path/filepath"
	"strconv"
	"time"

	"github.com/astaxie/beego/config"
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/common/file"
	"github.com/blocktree/openwallet/v2/crypto/sha3"
	"github.com/blocktree/openwallet/v2/hdkeystore"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/bndr/gotabulate"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/shopspring/decimal"

	//"github.com/go-ethereum/crypto/secp256k1"
	"sort"

	"github.com/blocktree/go-owcrypt"
)

const (
	maxAddresNum = 10000000
)

var (
	coinDecimal         = decimal.New(10, 17)
	coinDecimal_float64 = 1E18
)

type WalletManager struct {
	Storage      *hdkeystore.HDKeystore        //秘钥存取
	WalletClient *Client                       // 节点客户端
	Config       *WalletConfig                 //钱包管理配置
	WalletsInSum map[string]*openwallet.Wallet //参与汇总的钱包
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
func (wm *WalletManager) signTransaction(hash []byte, sk []byte) (string, error) {
	//s, _ := owcrypt.ICX_signature(sk[:], hash[:])
	s, v, _ := owcrypt.Signature(sk[:], nil, hash[:], owcrypt.ECC_CURVE_SECP256K1)
	s = append(s, v)
	//s, _ := secp256k1.Sign(hash[:], sk[:])
	es := base64.StdEncoding.EncodeToString(s)

	return es, nil
}

//计算交易体hash
/*
{
	 	"version": "0x3",
	 	"from": "hx2006f91de4cd0b9ce74cb00a06e66eaeb44c70b1",
	 	"stepLimit": "0xf4240",
	 	"timestamp": "0x5796c0695145c",
	 	"nid": "0x1",
	 	"to": "hxb12addba58c934ff924aa87ee65d06ee20f89eb8",
	 	"value": "0x8f0d180",
	 	"nonce": "0x64",
	 	"signature": "9G7dKZx1KcdTAwbwtSlnlNklB3EQRpGBUK5kb5nqW1MyM9RxmiVvODbD9HaCuERKbf/KN5tlL1T5V95kN6x+RAE=",
	 }

*/
func (wm *WalletManager) CalculateTxHash(from, to, value string, stepLimit, nonce int64) (map[string]interface{}, [32]byte) {
	tx_temp := make(map[string]interface{})

	tx_temp["version"] = "0x3" //API版本
	tx_temp["from"] = from
	tx_temp["nid"] = "0x1" //主网ID

	//hnonce := fmt.Sprintf("%x", nonce)
	//tx_temp["nonce"] = "0x" + hnonce

	hstepLimit := fmt.Sprintf("%x", stepLimit)
	tx_temp["stepLimit"] = "0x" + hstepLimit

	tx_temp["timestamp"] = "0x" + strconv.FormatInt(time.Now().UnixNano()/1000, 16)
	tx_temp["to"] = to

	vd, _ := decimal.NewFromString(value)
	//v := vd.Mul(coinDecimal).IntPart()
	vstr := vd.Mul(coinDecimal).String()
	bigVal := new(big.Int)
	bigVal.SetString(vstr, 10)
	hvalue := fmt.Sprintf("%x", bigVal)
	tx_temp["value"] = "0x" + hvalue

	//bigVal2 := new(big.Int)
	//bigVal2.SetString(hvalue, 16)
	//dec2 := decimal.NewFromBigInt(bigVal2, 0)
	//dec2str := dec2.String()
	//fmt.Println(dec2str)

	//把key 按首字母升序排序，否则交易hash就不对，导致address match错误
	var keys []string
	for k := range tx_temp {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	//icx_sendTransaction.from.nid.nonce.stepLimit.timestamp.to.value.version.
	str_tmp := "icx_sendTransaction"
	for _, key := range keys {
		str_tmp += "."
		str_tmp += string(key)
		str_tmp += "."
		str_tmp += tx_temp[key].(string)
	}
	//log.Info(str_tmp)

	hash := sha3.Sum256([]byte(str_tmp)[:])
	//log.Info(hexutil.Encode(hash[:]))

	return tx_temp, hash
}

//转账
func (wm *WalletManager) Transfer(sk []byte, from, to, value string, stepLimit, nonce int64) (string, error) {
	tx, hash := wm.CalculateTxHash(from, to, value, stepLimit, nonce)

	sig, err := wm.signTransaction(hash[:], sk)
	if err != nil {
		return "", err
	}
	//log.Info(sig)

	tx["signature"] = sig
	//log.Info(tx)

	ret, err := wm.WalletClient.Call_icx_sendTransaction(tx)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return ret, nil
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
		synCount   = 10
		quit       = make(chan struct{})
		done       = 0 //完成标记
		shouldDone = 0 //需要完成的总数
	)

	db, err := wallet.OpenDB()
	if err != nil {
		return decimal.NewFromFloat(0), nil, err
	}
	defer db.Close()

	var addrs []*openwallet.Address
	db.All(&addrs)

	balance, _ := decimal.NewFromString("0")
	count := len(addrs)
	if count <= 0 {
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
		for balances := range addrs {
			//
			//balances := <-addrs

			for _, b := range balances {
				addrB, _ := decimal.NewFromString(b.Balance)
				balance = balance.Add(addrB)
			}

			//累计完成的线程数
			done++
			if done == shouldDone {
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}(worker)

	/*	计算synCount个线程，内部运行的次数	*/
	//每个线程内循环的数量，以synCount个线程并行处理
	runCount := count / synCount
	otherCount := count % synCount

	if runCount > 0 {
		for i := 0; i < synCount; i++ {
			//开始
			//log.Std.Info("Start get balance thread[%d]", i)
			start := i * runCount
			end := (i+1)*runCount - 1
			as := addrs[start:end]

			go func(producer chan []*openwallet.Address, addrs []*openwallet.Address, wm *WalletManager) {
				var bs []*openwallet.Address
				for _, a := range addrs {
					b, err := wm.WalletClient.Call_icx_getBalance(a.Address)
					if err != nil {
						log.Error(err)
						continue
					}
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
		start := runCount * synCount
		as := addrs[start:]

		go func(producer chan []*openwallet.Address, addrs []*openwallet.Address, wm *WalletManager) {
			var bs []*openwallet.Address
			for _, a := range addrs {
				b, err := wm.WalletClient.Call_icx_getBalance(a.Address)
				if err != nil {
					log.Error(err)
					continue
				}
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
			//log.Debug("produced")
		case activeWorker <- activeValue:
			values = values[1:]
			//log.Debug("consumed")
		case <-quit:
			//退出
			log.Std.Info("wallet %s get all addresses's balance finished", wallet.Alias)
			return balance, outputAddress, nil
		}
	}

	return balance, outputAddress, nil
}

//打印钱包列表
func (wm *WalletManager) printWalletList(list []*openwallet.Wallet, getBalance bool) [][]*openwallet.Address {
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
		t.SetHeaders([]string{"No.", "ID", "Name", "DBFile", "Balance"})
	} else {
		t.SetHeaders([]string{"No.", "ID", "Name", "DBFile"})
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

	pk := childKey.GetUncompressedPublicKeyBytes()

	cfg := addressEncoder.ICX_walletAddress
	address := addressEncoder.AddressEncode(pk, cfg)

	addr := openwallet.Address{
		Address:     address,
		AccountID:   key.KeyID,
		HDPath:      derivedPath,
		CreatedTime: time.Now().Unix(),
		Symbol:      wm.Config.Symbol,
		Index:       index,
		WatchOnly:   false,
	}

	return &addr, err
}

func (wm *WalletManager) GetLastBlock() (*gjson.Result, error) {
	return wm.WalletClient.Call("icx_getLastBlock", map[string]interface{}{})
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

func (wm *WalletManager) summaryWallet(wallet *openwallet.Wallet, password string) error {

	//加载钱包
	key, err := wallet.HDKey(password)
	if err != nil {
		return err
	}

	totalBalance, addrs, err := wm.getWalletBalance(wallet)
	if err != nil {
		return err
	}

	if totalBalance.GreaterThan(wm.Config.Threshold) {
		for _, a := range addrs {
			k, _ := wm.getKeys(key, a)

			decimal_balance, _ := decimal.NewFromString(a.Balance)

			if decimal_balance.GreaterThanOrEqual(wm.Config.minTransfer) {
				// 将该地址的余额减去矿工费后，全部转到汇总地址
				amount := decimal_balance.Sub(wm.Config.fees)
				if amount.GreaterThan(decimal.Zero) {
					txid, _ := wm.Transfer(k.PrivateKey, a.Address, wm.Config.SumAddress, amount.String(), wm.Config.StepLimit, 100)
					log.Std.Info("summary from address:%s, to address:%s, amount:%s, txid:%s", k.Address, wm.Config.SumAddress, amount.String(), txid)
				}
			}
		}
	}

	return nil
}

//汇总钱包
func (wm *WalletManager) SummaryWallets() {
	log.Std.Info("[Summary Wallet Start]------%s", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for _, wallet := range wm.WalletsInSum {
		wm.summaryWallet(wallet, wallet.Password)
	}

	log.Std.Info("[Summary Wallet end]------%s", common.TimeFormat("2006-01-02 15:04:05"))
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
		return errors.New("Config is not setup. Please run 'wmd wallet config -s <symbol>' ")
	}

	wm.Config.ServerAPI = c.String("apiUrl")
	wm.Config.Threshold, _ = decimal.NewFromString(c.String("threshold"))
	wm.Config.SumAddress = c.String("sumAddress")
	wm.Config.StepLimit, _ = c.Int64("stepLimit")
	wm.Config.minTransfer, _ = decimal.NewFromString(c.String("minTransfer"))
	wm.Config.fees, _ = decimal.NewFromString(c.String("fees"))

	cyclesec := c.String("cycleSeconds")
	if cyclesec == "" {
		return errors.New(fmt.Sprintf(" cycleSeconds is not set, sample: 1m , 30s, 3m20s etc... Please set it in './conf/%s.ini' \n", Symbol))
	}

	wm.Config.CycleSeconds, _ = time.ParseDuration(cyclesec)
	wm.WalletClient = NewClient(wm.Config.ServerAPI, false)

	return nil
}

//RestoreWallet 恢复钱包
func (wm *WalletManager) RestoreWallet(keyFile, dbFile, password string) error {

	//根据流程，提供种子文件路径，wallet.db文件的路径，钱包数据库文件的路径。
	//输入钱包密码。
	//复制种子文件到data/btc/key/。
	//复制钱包数据库文件到data/btc/db/。

	var (
		err error
		key *hdkeystore.HDKey
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
func (wm *WalletManager) getKeys(key *hdkeystore.HDKey, a *openwallet.Address) (*Key, error) {
	childKey, err := key.DerivedKeyWithPath(a.HDPath, wm.Config.CurveType)
	if err != nil {
		return nil, err
	}

	prikey, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return nil, err
	}

	pubkey := childKey.GetPublicKeyBytes()

	k := &Key{a.Address, pubkey, prikey}

	return k, nil
}
