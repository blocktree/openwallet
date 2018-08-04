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
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/blake2b"
	"log"
	"strconv"
	"github.com/tidwall/gjson"
	"path/filepath"
	"github.com/astaxie/beego/config"
	"errors"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/bndr/gotabulate"
	"fmt"
	"github.com/blocktree/OpenWallet/common/file"
)

const (
	maxAddresNum = 10000000
)

var (
	//钱包服务API
	serverAPI = "https://rpc.tezrpc.me"
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000)
	//参与汇总的钱包
	//walletsInSum = make(map[string]*Wallet)
	//汇总阀值
	threshold decimal.Decimal = decimal.NewFromFloat(1).Mul(coinDecimal)
	//最小转账额度
	minSendAmount decimal.Decimal = decimal.NewFromFloat(1).Mul(coinDecimal)
	//最小矿工费
	minFees decimal.Decimal = decimal.NewFromFloat(0.0001).Mul(coinDecimal)

	//gas limit 和 storage limit
	gasLimit decimal.Decimal = decimal.NewFromFloat(0.0001).Mul(coinDecimal)
	storageLimit decimal.Decimal = decimal.NewFromFloat(0.0001).Mul(coinDecimal)
	sumAddress = ""
	//汇总执行间隔时间
	cycleSeconds = time.Second * 300
)


//地址，公钥，公钥哈希，私钥，签名前缀
var prefix = map[string][]byte{
	"tz1": {6, 161, 159},
	"tz2": {6, 161, 161},
	"edpk": {13, 15, 37, 217},
	"edsk": {43, 246, 78, 7},
	"edsig": {9, 245, 205, 134, 18},
}

//消息前缀
var watermark = map[string][]byte{
	"block": {1},
	"endorsement": {2},
	"generic": {3},
}


//创建地址
func createAccount() (string, string , string) {
	pub, pri, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Println(err.Error())
	}

	ctx, err:=blake2b.New(20,nil)
	ctx.Write(pub[:])
	pubhash := ctx.Sum(nil)

	pk := base58checkEncode(pub, prefix["edpk"])
	sk := base58checkEncode(pri, prefix["edsk"])
	pkh := base58checkEncode(pubhash, prefix["tz1"])

	return pk, sk, pkh
}


//加密私钥
func encryptSecretKey(sk string, password string) string {
	ret, err := Encrypt(password, sk)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}

	return ret
}

//解密私钥
func decryptSecretKey(esk string, password string) string {
	ret, err := Decrypt(password, esk)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}

	return ret
}

//签名交易
func signTransaction(hash string, sk string, wm []byte) (string, string, error) {
	bhash,_ := hex.DecodeString(hash)
	merbuf := append(wm, bhash...)
	ctx, err :=blake2b.New(32,nil)
	if err != nil {
		return "", "", err
	}
	ctx.Write(merbuf[:])
	bb := ctx.Sum(nil)

	sks, err:= base58checkDecodeNormal(sk, prefix["edsk"])
	if err != nil {
		return "", "", err
	}

	sig := ed25519.Sign(sks[:], bb[:])
	edsig := base58checkEncode(sig, prefix["edsig"])

	sbyte := hash + hex.EncodeToString(sig[:])

	return edsig, sbyte, nil
}

//判断该key是否需要reverl
func isReverlKey(pubkey string) bool {
	manager_key := callGetManagerKey(pubkey)
	//manager := gjson.GetBytes(ret, "manager")
	key := gjson.GetBytes(manager_key, "key")

	if key.Str == "" {
		return true
	}

	return false
}

//转账
func transfer(keys Key, dst string, fee, gas_limit, storage_limit, amount string) (string, string){
	header := callGetHeader()
	blk_hash := gjson.GetBytes(header, "hash").Str
	chain_id := gjson.GetBytes(header, "chain_id").Str
	protocol := gjson.GetBytes(header, "protocol").Str

	counter :=callGetCounter(keys.Address)
	icounter,_ := strconv.Atoi(string(counter))
	icounter = icounter + 1

	manager_key := callGetManagerKey(keys.Address)
	//manager := gjson.GetBytes(ret, "manager")
	key := gjson.GetBytes(manager_key, "key")

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
	if key.Str == ""{
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
	hash := callForgeOps(chain_id, blk_hash, opOb)

	//sign
	edsig, sbyte, _ := signTransaction(hash, keys.PrivateKey, watermark["generic"])

	//preapply operations
	var opObs []interface{}
	opOb["signature"] = edsig
	opOb["protocol"] = protocol
	opObs = append(opObs, opOb)
	pre := callPreapplyOps(opObs)

	//jnject aperations
	inj := callInjectOps(sbyte)
	return string(inj), string(pre)
}

//exportAddressToFile 导出地址到文件中
func exportAddressToFile(keys []*Key, filePath string) {

	var (
		content string
	)

	for _, a := range keys {

		log.Printf("Export: %s \n", a.Address)

		content = content + a.Address + "\n"
	}

	file.MkdirAll(addressDir)
	file.WriteFile(filePath, []byte(content), true)
}

func createAddressWork(producer chan<- []*Key, password string, start, end uint64) {
	runAddress := make([]*Key, 0)

	for i := start; i < end; i++ {
		// 生成地址
		pk, sk, pkh := createAccount()
		esk, _ := Encrypt(password, sk)
		key := &Key{pkh, pk, esk}
		runAddress = append(runAddress, key)
	}

	//生成完成
	producer <- runAddress
}

//CreateNewWallet 创建钱包
func CreateNewWallet(name string) error {

	walletID := openwallet.NewWalletID()

	wallet := openwallet.NewWatchOnlyWallet(walletID.String(), Symbol)
	wallet.Alias = name

	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Save(wallet)
}

func AddWalletInSummary(wid string, wallet *openwallet.Wallet) {
	walletsInSum[wid] = wallet
}

//打印钱包列表
func printWalletList(list []*openwallet.Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {

		tableInfo = append(tableInfo, []interface{}{
			i, w.WalletID, w.Alias, w.DBFile,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "DBFile"})

	//打印信息
	fmt.Println(t.Render("simple"))

}

func CreateBatchAddress(walletId, password string, count uint64) (string, []*Key, error) {
	var (
		synCount   uint64 = 20
		quit              = make(chan struct{})
		done              = 0 //完成标记
		shouldDone        = 0 //需要完成的总数
	)

	w, _ := GetWalletByID(walletId)

	timestamp := time.Now()
	//建立文件名，时间格式2006-01-02 15:04:05
	filename := "address-" + common.TimeFormat("20060102150405", timestamp) + ".txt"
	filePath := filepath.Join(addressDir, filename)

	//生产通道
	producer := make(chan []*Key)
	defer close(producer)

	//消费通道
	worker := make(chan []*Key)
	defer close(worker)

	//保存地址过程
	saveAddressWork := func(addresses chan []*Key, filename string, w *openwallet.Wallet) {
		var (
			saveErr error
		)

		for {
			//回收创建的地址
			getAddrs := <-addresses

			//批量写入数据库
			saveErr = SaveKeyToWallet(w, getAddrs)
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

	/*	开启导出的线程，监听新地址，批量导出	*/

	go saveAddressWork(worker, filePath, w)

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
			go createAddressWork(producer, password, s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		log.Printf("Start create address thread[REST]\n")
		s := count - otherCount
		e := count
		go createAddressWork(producer, password, s, e)

		shouldDone++
	}

	values := make([][]*Key, 0)
	outputAddress := make([]*Key, 0)

	//以下使用生产消费模式

	for {

		var activeWorker chan<- []*Key
		var activeValue []*Key

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
			//当激活消费者后，传输数据给消费者，并把顶部数据出队
		case activeWorker <- activeValue:
			//log.Printf("Get %d", len(activeValue))
			values = values[1:]

		case <-quit:
			//退出
			log.Printf("All addresses have been created!")
			return filePath, outputAddress, nil
		}
	}

	return filePath, outputAddress, nil
}

//inputNumber 输入地址数量
func inputNumber() uint64 {

	var (
		count uint64 = 0 // 输入的创建数量
	)

	for {
		// 等待用户输入参数
		line, err := console.Stdin.PromptInput("Enter the number of addresses you want: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v", err)
			return 0
		}
		count = common.NewString(line).UInt64()
		if count < 1 {
			log.Printf("Input number must be greater than 0!\n")
			continue
		}
		break
	}

	return count
}

//loadConfig 读取配置
func loadConfig() error {

	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err = config.NewConfig("json", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	serverAPI = c.String("apiURL")
	threshold, _ = decimal.NewFromString(c.String("threshold"))
	threshold = threshold.Mul(coinDecimal)
	minSendAmount, _ = decimal.NewFromString(c.String("minSendAmount"))
	minSendAmount = minSendAmount.Mul(coinDecimal)
	minFees, _ = decimal.NewFromString(c.String("minFees"))
	minFees = minFees.Mul(coinDecimal)
	gasLimit, _ = decimal.NewFromString(c.String("gasLimit"))
	gasLimit = gasLimit.Mul(coinDecimal)
	storageLimit, _ = decimal.NewFromString(c.String("storageLimit"))
	storageLimit = storageLimit.Mul(coinDecimal)
	sumAddress = c.String("sumAddress")

	return nil
}

func summaryWallet(wallet *openwallet.Wallet, password string) error{
	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var keys []*Key
	db.All(&keys)

	for _, k := range keys {
		//get balance
		decimal_balance, _ := decimal.NewFromString(string(callGetbalance(k.Address)))
		//decrypt sk
		sk, _ := Decrypt(password, k.PrivateKey)

		//判断是否是reveal交易
		fee := minFees
		isReverl := isReverlKey(k.Address)
		if isReverl {
			//多了reveal操作后，fee * 2
			fee = minFees.Mul(decimal.RequireFromString("2"))
		}
		// 将该地址多余额减去矿工费后，全部转到汇总地址
		amount := decimal_balance.Sub(fee)
		//该地址预留一点币，否则交易会失败，暂定20／1000000 tez
		amount = amount.Sub(decimal.RequireFromString("20"))

		log.Printf("address:%s banlance:%d amount:%d fee:%d\n", k.Address, decimal_balance.IntPart(), amount.IntPart(), fee.IntPart())

		k.PrivateKey = sk
		if decimal_balance.GreaterThan(threshold) {
			txid, _ := transfer(*k, sumAddress, strconv.FormatInt(minFees.IntPart(), 10), strconv.FormatInt(gasLimit.IntPart(), 10),
				strconv.FormatInt(storageLimit.IntPart(), 10),strconv.FormatInt(amount.IntPart(), 10))

			log.Printf("summary address:%s, to address:%s, amount:%d, txid:%s\n", k.Address, sumAddress, amount.IntPart(), txid)
		}
	}

	return nil
}

//汇总钱包
func SummaryWallets() {
	log.Printf("[Summary Wallet Start]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))

	//读取参与汇总的钱包
	for _, wallet := range walletsInSum {
		summaryWallet(wallet, wallet.Password)
	}

	log.Printf("[Summary Wallet end]------%s\n", common.TimeFormat("2006-01-02 15:04:05"))
}
