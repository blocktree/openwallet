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

package nebulasio

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/asdine/storm"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"github.com/blocktree/go-OWCrypt"
	"github.com/bndr/gotabulate"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/gogo/protobuf/proto"
	"github.com/nebulasio/go-nebulas/core/pb"
	"github.com/nebulasio/go-nebulas/rpc/pb"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/shopspring/decimal"
	"math/big"
	"path/filepath"
	"strconv"
	"time"
)

//from nebulasio
const (
	TxPayloadBinaryType = "binary"
	TxPayloadDeployType = "deploy"
	TxPayloadCallType   = "call"
)


const (
	maxAddresNum = 10000000
)
var (
	//coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000)
	coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000000000000000)
)

var(
	//Account Address前缀
	addr_prefix = []byte{0x19,0x57}
	//Smart Contract Address前缀
	Contract_prefix = []byte{0x19,0x58}
)

//保存每个账户的地址、公钥、私钥、nonce
type Key struct {
	Address    string `storm:"id"`
	PublicKey  []byte
	PrivateKey []byte
	Nonce  	   string
}

type WalletManager struct {
	Storage      *hdkeystore.HDKeystore         //秘钥存取
	WalletClient *Client                        // 节点客户端
	Config       *WalletConfig                  //钱包管理配置
	WalletsInSum map[string]*openwallet.Wallet  //参与汇总的钱包
	Blockscanner *NASBlockScanner             //区块扫描器
	Decoder        openwallet.AddressDecoder     //地址编码器
	TxDecoder      openwallet.TransactionDecoder //交易单编码器
	ContractDecoder openwallet.SmartContractDecoder //智能合约解析器
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey)
	storage := hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.Storage = storage
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = NewNASBlockScanner(&wm)
	//地址解析器
	wm.Decoder =  NewAddressDecoder(&wm)
	//交易单解析器
	wm.TxDecoder = NewTransactionDecoder(&wm)

	wm.WalletClient = NewClient(wm.Config.ServerAPI, false)

	return &wm
}


//SubmitTransaction 存放最终广播出去的交易单信息
type SubmitTransaction struct{
	Hash []byte		//交易hash
	From []byte
	To	 []byte
	Value[]byte
	Nonce uint64
	Timestamp int64
	Data *corepb.Data
	ChainId uint32
	GasPrice []byte
	GasLimit []byte
	Alg	uint32
	Sign []byte		//交易hash签名结果
}


//将入参address string地址转成[]byte
func NasAddrTobyte(addr string) ([]byte, error){
		//解压出20字节地址部分hash
		only_address,err := addressEncoder.AddressDecode(addr,addressEncoder.NAS_AccountAddress)
		if err != nil{
			return nil,err
		}
		//拼接前缀+20字节地址hash
		addr_tmp := [][]byte{
			addr_prefix,
			only_address,
		}
		prefix_address  := bytes.Join(addr_tmp, []byte(""))
		//对前缀+20字节地址算哈希后取哈希值后4字节为后缀
		hash_prefix_address := owcrypt.Hash(prefix_address,32,owcrypt.HASH_ALG_SHA3_256)
		//前缀+20字节地址+后缀
		addr_tmp2 := [][]byte{
			prefix_address,
			hash_prefix_address[0:4],
		}
		//前缀+20字节地址+后缀 ：[]byte
		address := bytes.Join(addr_tmp2, []byte(""))

		return address,err
}

//对签名结果按照官方进行编码
func (tx *SubmitTransaction) ToProto() (proto.Message) {

	return &corepb.Transaction{
		Hash:      tx.Hash,
		From:      tx.From,
		To:        tx.To,
		Value:     tx.Value,
		Nonce:     tx.Nonce,
		Timestamp: tx.Timestamp,
		Data:      tx.Data,
		ChainId:   tx.ChainId,
		GasPrice:  tx.GasPrice,
		GasLimit:  tx.GasLimit,
		Alg:       tx.Alg,
		Sign:      tx.Sign,
	}
}

//CreateRawTransaction创建交易单hash
func (wm *WalletManager) CreateRawTransaction(from , to, gasLimit, gasPrice,value string, nonce uint64) (*SubmitTransaction,error){

	chainID,_ := wm.GetChainID()
	timestamp := time.Now().Unix()
	address_from ,err := NasAddrTobyte(from)
	if err != nil{
		return nil,err
	}
	address_to ,err := NasAddrTobyte(to)
	if err != nil{
		return nil,err
	}
	payloadType := TxPayloadBinaryType
	payload := []byte{}

	data_p := &corepb.Data{Type: payloadType, Payload: payload}
	value_p, err := util.NewUint128FromString(value)
	if err != nil {
		return nil, errors.New("invalid value")
	}
	gasPrice_p, err := util.NewUint128FromString(gasPrice)
	if err != nil {
		return nil, errors.New("invalid gasPrice")
	}
	gasLimit_p, err := util.NewUint128FromString(gasLimit)
	if err != nil {
		return nil, errors.New("invalid gasLimit")
	}

	data_byte, err := proto.Marshal(data_p)
	if err != nil {
		return nil, err
	}
	value_byte, err := value_p.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	gasPrice_byte, err := gasPrice_p.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	gasLimit_byte, err := gasLimit_p.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}

	//拼接前缀+20字节地址hash
	transaction_tmp := [][]byte{
		address_from,
		address_to,
		value_byte,
		byteutils.FromUint64(nonce),
		byteutils.FromInt64(timestamp),
		data_byte,
		byteutils.FromUint32(chainID),
		gasPrice_byte,
		gasLimit_byte,
	}
	transaction  := bytes.Join(transaction_tmp, []byte(""))
	//计算交易hash
	transaction_hash := owcrypt.Hash( transaction, uint16(len(transaction)), owcrypt.HASH_ALG_SHA3_256)

	Submit_tx := &SubmitTransaction{
		transaction_hash,
		address_from,
		address_to,
		value_byte,
		nonce,
		timestamp ,//int64(1), //timestamp
		data_p,
		chainID,
		gasPrice_byte,
		gasLimit_byte,
		uint32(1),
		nil,
	}

	return Submit_tx,err
}

//SignRawTransaction对交易hash进行签名
func SignRawTransaction(PrivateKey []byte,txhash []byte) ([]byte,error){

	//调用ow库签名交易结果
	signed,ret := owcrypt.NAS_signature(PrivateKey,txhash)   //对交易体进行签名结果 //65字节
	if ret != owcrypt.SUCCESS {
		errdesc := fmt.Sprintln("signature error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		log.Error(errdesc)
		return nil,errors.New(errdesc)
	}
	//log.Std.Info("sigx=%x\n",sig)
	//调用ow库签名交易结果	End

	return signed,nil
}

//VerifyRawTransaction对签名进行验签
//return :owcrypt.SUCCESS：成功，其他：失败
func VerifyRawTransaction(PubKey []byte, txhash []byte, signed []byte) uint16{

	//调用ow库进行验证签名
	//公钥为33字节，需要解压缩出65字节公钥
	PublicKey :=owcrypt.PointDecompress(PubKey , CurveType)
	//去掉65字节公钥的第一个字节后进行验签
	verify :=owcrypt.Verify(PublicKey[1:65],nil,0,txhash,32,signed[0:64],owcrypt.ECC_CURVE_SECP256K1 | (1<<9))
	//log.Std.Info("Verify success! verify=%x\n", verify)
	//调用ow库进行验证签名 End
	return verify
}

//SubmitRawTransaction对签名结果进行编码
func  EncodeTransaction( submit_tx *SubmitTransaction)(string ,error){

	//参照官方对广播交易体进行编码
	message := submit_tx.ToProto()
	data_tmp, err := proto.Marshal(message)
	if err != nil {
		return "", err
	}
	data := rpcpb.SignTransactionPassphraseResponse{Data: data_tmp}
	broadcastsend_data := base64.StdEncoding.EncodeToString(data.Data)

	return broadcastsend_data,nil
}

//SubmitRawTransaction对签名编码后的数据进行广播
func (wm *WalletManager)SubmitRawTransaction( submit_data string)(string ,error){

	txhash, err := wm.WalletClient.CallSendRawTransaction(submit_data)
	if (err != nil) || (len(txhash)==0){
		return "" ,err
	}

	log.Std.Info("txhash=%v\n", txhash)
	return txhash,nil
}



//发送交易
func (wm *WalletManager) Transfer(key *Key, from ,to, gasLimit, value string) (string, error) {

	//CreateRawTransaction 创建交易单
	gasPrice := wm.EstimateFeeRate()
	nonce := wm.WalletClient.CheckNonce(key)
	transaction,err := wm.CreateRawTransaction(from, to, gasLimit, gasPrice,value,nonce)
	if err != nil{
		return "",err
	}
	//SignRawTransaction 签名交易单
	signed, err := SignRawTransaction(key.PrivateKey,transaction.Hash)
	if err != nil{
		return "",err
	}
	//VerifyRawTransaction 验证交易单，
	verify_result := VerifyRawTransaction(key.PublicKey,transaction.Hash,signed)
	if verify_result != owcrypt.SUCCESS {
		return "", errors.New("Verify Failed !")
	}
	//验证通过后对交易单进行相应的编码
	transaction.Sign = signed
	SubmitData,err := EncodeTransaction(transaction)
	if err != nil {
		return "", err
	}

	//SendRawTransaction 广播交易单
	txid ,err := wm.SubmitRawTransaction(SubmitData)
	if err != nil{
		return "",err
	}

	return txid,nil
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

	//fmt.Printf("Create new wallet keystore...\n")

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

	//fmt.Printf("keyFile=%v\n",keyFile)
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
	//fmt.Printf("count=%v,runCount=%v,otherCount=%v\n",count,runCount,otherCount)

	if runCount > 0 {
		for i := 0; i < synCount; i++ {
			//开始
			//log.Std.Info("Start get balance thread[%d]", i)
			start := i * runCount
			end := (i + 1) * runCount
			as := addrs[start:end]

			go func(producer chan []*openwallet.Address, addrs []*openwallet.Address, wm *WalletManager) {
				var bs []*openwallet.Address
				for _, a := range(addrs) {
					b,_ := wm.WalletClient.CallGetaccountstate(a.Address,"balance")
					//fmt.Printf("runCount_a.Address=%s,balance=%s\n",a.Address,b)

					a.Balance = b
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
				b,_ := wm.WalletClient.CallGetaccountstate(a.Address,"balance")
				//fmt.Printf("otherCount_a.Address=%s,balance=%s\n",a.Address,b)

				a.Balance = b
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
			//log.Std.Info("wallet %s get all addresses's balance finished", wallet.Alias)
			return balance.Div(coinDecimal), outputAddress, nil
		}
	}

	return balance.Div(coinDecimal), outputAddress, nil
}


//冒泡算法，对一个钱包下面的所有地址进行从大到小排序后供转账使用，提高效率
func SortAddrFromBalance(MixAddr []*openwallet.Address) []*openwallet.Address {

	var BigTosmall []*openwallet.Address
	num := len(MixAddr)
	//冒泡排序
	for i := 0; i < num; i++ {
		for j := i + 1; j < num; j++ {

			if decimal.RequireFromString(MixAddr[i].Balance).LessThanOrEqual(decimal.RequireFromString(MixAddr[j].Balance)) {
				MixAddr[i], MixAddr[j] = MixAddr[j], MixAddr[i]
			}
		}
		BigTosmall = append(BigTosmall,MixAddr[i])
		//log.Std.Info("BigTosmall[%d],addr=%v,balance=%v",i,BigTosmall[i].Address,BigTosmall[i].Balance)
	}
	return BigTosmall
}

//打印钱包列表
func (wm *WalletManager) printWalletList(list []*openwallet.Wallet, getBalance bool) ([][]*openwallet.Address) {
	tableInfo := make([][]interface{}, 0)
	var addrs [][]*openwallet.Address
	for i, w := range list {
		if getBalance {
			balance, addr, _ := wm.getWalletBalance(w)
			addr_sort := SortAddrFromBalance(addr)
			tableInfo = append(tableInfo, []interface{}{
				i, w.WalletID, w.Alias, w.DBFile, balance,
			})

			addrs = append(addrs, addr_sort)
			//休眠2秒是因为http请求会导致下一个钱包获取余额API请求失败
			if i != (len(list) - 1) {
				time.Sleep(time.Second * 2)
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


/*//CreateNewPrivateKey 创建私钥，返回私钥wif格式字符串
func (wm *WalletManager) CreateNewPrivateKey(key *hdkeystore.HDKey, start, index uint64) (string, *openwallet.Address, error) {

	derivedPath := fmt.Sprintf("%s/%d/%d", key.RootPath, start, index)
	//fmt.Printf("derivedPath = %s\n", derivedPath)
	wm.Config.CurveType = owcrypt.ECC_CURVE_SECP256K1

	childKey, err := key.DerivedKeyWithPath(derivedPath, wm.Config.CurveType)
	if err != nil {
		return "", nil, err
	}

	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return "", nil, err
	}

	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)

	cfg := chaincfg.MainNetParams
	if wm.Config.isTestNet {
		cfg = chaincfg.TestNet3Params
	}

	wif, err := btcutil.NewWIF(privateKey, &cfg, true)
	if err != nil {
		return "", nil, err
	}

	//address, err := childKey.Address(&cfg)
	//if err != nil {
	//	return "", nil, err
	//}

	pubkey, _ := owcrypt.GenPubkey(privateKey.Serialize(), owcrypt.ECC_CURVE_SECP256K1)

	//fmt.Println(hex.EncodeToString(pubkey))

	//pubdata := append([]byte{0x04}, pubkey[:]...)
	pubdata := owcrypt.PointCompress(pubkey, owcrypt.ECC_CURVE_SECP256K1)

	//fmt.Println(hex.EncodeToString(pubdata))

	pubkeyHash := owcrypt.Hash(pubdata, 0, owcrypt.HASH_ALG_HASH160)

	//fmt.Println(hex.EncodeToString(pubkeyHash))

	address := addressEncoder.AddressEncode(pubkeyHash, addressEncoder.QTUM_mainnetAddressP2PKH)

	addr := openwallet.Address{
		Address:   address,
		AccountID: key.KeyID,
		HDPath:    derivedPath,
		CreatedAt: time.Now(),
		Symbol:    wm.Config.symbol,
		Index:     index,
		WatchOnly: false,
	}

	//addr := Address{
	//	Address:   address.String(),
	//	Account:   key.RootId,
	//	HDPath:    derivedPath,
	//	CreatedAt: time.Now(),
	//}

	return wif.String(), &addr, err
}
*/

//CreateAddrFromPublic 创建地址
func (wm *WalletManager) CreateAddrFromPublic (key *hdkeystore.HDKey, start, index uint64) (*openwallet.Address, error) {
	derivedPath := fmt.Sprintf("%s/%d/%d", key.RootPath, start, index)
	//根据derivedPath和曲线得到childkey
	childKey, err := key.DerivedKeyWithPath(derivedPath, wm.Config.CurveType)
	if err != nil {
		return nil, err
	}

	//根据childKey取公钥,压缩33字节,02730923f3f99eb587cbbcfa4876a9be518d8893a2106ebb93e40def9af95c308c
	publicKey := childKey.GetPublicKeyBytes()
	//fmt.Printf("publicKey_encode=%x\n", publicKey)
	// 对压缩的公钥进行解压,65字节
	// 04730923f3f99eb587cbbcfa4876a9be518d8893a2106ebb93e40def9af95c308ce09de098076869d067c3e82564e673de3f965585d2466383349b20bd8bface0a
	PublicKey_decode :=owcrypt.PointDecompress(publicKey , owcrypt.ECC_CURVE_SECP256K1)
	//log.Std.Info("PublicKey_decode=%x\n", PublicKey_decode)

	//对于有些币种只对整个解压的公钥[1:65]或者压缩的33字节公钥算hash之后编码得到地址
	//NAS需要对整个解压的公钥算hash之后编码得到地址 n1VC5UsE2mVRYMuctV66BtqFkdi1KCUZodY,
	//因为官方节点验签规则：根据广播交易单对签名结果恢复得到65字节公钥，再去算签名的地址,与from地址进行对比，一致则验签通过
	cfg := addressEncoder.NAS_AccountAddress
	pkHash := owcrypt.Hash(PublicKey_decode, 20, owcrypt.HASH_ALG_SHA3_256_RIPEMD160)
	address := addressEncoder.AddressEncode(pkHash, cfg)
	//log.Std.Info("Create_address=%v\n",address)

	addr := openwallet.Address{
		Address:   address,
		AccountID: key.KeyID,
		HDPath:    derivedPath,
		CreatedTime: time.Now().Unix(),
		Symbol:    wm.Config.Symbol,
		Index:     index,
		WatchOnly: false,
		ExtParam:  "0",			//创建地址时初始nonce为0
	}

	return &addr, err
}

//createAddressWork 创建地址过程
func (wm *WalletManager) createAddressWork(k *hdkeystore.HDKey, producer chan<- []*openwallet.Address, walletID string, index, start, end uint64) {

	runAddress := make([]*openwallet.Address, 0)
	for i := start; i < end; i++ {
		// 生成地址
		address, errRun := wm.CreateAddrFromPublic(k, index, i)
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
			//log.Std.Info("Start create address thread[%d]", i)
			s := i * runCount
			e := (i + 1) * runCount
			go wm.createAddressWork(key, producer, walletId, uint64(timestamp.Unix()), s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {
		//开始创建地址
		//log.Std.Info("Start create address thread[REST]")
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
	//defer db.Close()

	var addrs []*openwallet.Address
	db.All(&addrs)


	//加载钱包
	key, err := wallet.HDKey(password)
	if err != nil {
		return err
	}

	for _, a := range addrs {
		k, _ := wm.getKeys(key, a)

		//get balance,单位Wei
		balance ,_ := wm.WalletClient.CallGetaccountstate(a.Address,"balance")
		balance_decimal := decimal.RequireFromString(balance)

		//该地址预留一点币，否则交易会失败，暂定0.00001 NAS
		balance_leave := decimal.RequireFromString("10000000000000")
		cmp_result := balance_decimal.Cmp(balance_leave)
		if  cmp_result==0 ||cmp_result==-1{
			continue
		}
		balance_safe := balance_decimal.Sub(balance_leave)

		//读取config的值 单位Wei
		//log.Std.Info("Threshold:%v", wm.Config.Threshold.String())
		//log.Std.Info("balance_safe=%v",balance_safe)
		if balance_safe.GreaterThan(wm.Config.Threshold) {
			txid, err := wm.Transfer(k,k.Address, wm.Config.SumAddress, wm.Config.GasLimit.String(),
				balance_safe.String())
			//log.Std.Info("summary form address:%s, to address:%s, amount:%s, txid:%s", k.Address, wm.Config.SumAddress, balance_safe.String(), txid)
			if err != nil{
				log.Std.Info("Transfer Fail!\n",)
			}else{
				log.Std.Info("Transfer Success! txid=%s\n",txid)

				err := NotenonceInDB(k.Address,db)
				if err != nil {
					log.Std.Info("NotenonceInDB error!\n")
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
		//log.Std.Info("Export: %s ", a.Address)
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

	if  c.String("threshold") == ""{
		return errors.New(fmt.Sprintf(" threshold is not set, uint is NAS... Please set it in './conf/%s.ini' \n", Symbol))
	}
	wm.Config.Threshold = (decimal.RequireFromString(c.String("threshold"))).Mul(coinDecimal)
	wm.Config.SumAddress = c.String("sumAddress")
	wm.Config.GasLimit = (decimal.RequireFromString(c.String("gasLimit"))).Mul(coinDecimal)

	cyclesec := c.String("cycleSeconds")
	if cyclesec == "" {
		return errors.New(fmt.Sprintf(" cycleSeconds is not set, sample: 1m , 30s, 3m20s etc... Please set it in './conf/%s.ini' \n", Symbol))
	}
	wm.Config.CycleSeconds, _ = time.ParseDuration(cyclesec)
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

func (wm *WalletManager) EstimateFeeRate() string {

	gasprice := wm.WalletClient.CallGetGasPrice()
	return gasprice
}

//通过hdpath获取地址、公钥、私钥、数据库nonce值
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

	//创建地址时ExtParam记录每个地址的nonce
	nonce := a.ExtParam

	k := &Key{a.Address,pubkey,prikey,nonce}

	//转换成带前缀公钥，交易结构中需要填充此类型公钥 for xtz
	//pk := base58checkEncode(pubkey, prefix["edpk"])
	//k := &Key{a.Address,pk,prikey}

	return k, nil
}

//GetBlockHeight 获取区块链高度
func (wm *WalletManager) GetChainID() (uint32, error) {

	result, err := wm.WalletClient.CallGetnebstate("chain_id")
	if err != nil {
		return 0, err
	}

	return uint32(result.Uint()), nil
}

//将签名成功广播出去后的nonce值记录在对应address的DB中
func  NotenonceInDB(addr string, db *storm.DB) error{

	//有db说明db已经在前文打开，不需要重复打开
/*	if db == nil{
		db, err := wallet.OpenDB()
		if err != nil {
			log.Error("open db failed, err=", err)
			return err

		}
		defer db.Close()
	}*/

	var address openwallet.Address
	err := db.One("Address", addr, &address)
	if err != nil {
		log.Debugf("get address failed, err=%v", err)
		return err
	}

	//modifyAddress.ExtParam for note nonce
	//Nonce为最后一次成功上链的nonce值
	address.ExtParam = strconv.Itoa(Nonce_Chain)

	//saveAddress
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.Save(&address)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}


type estimateGasParameter struct{
	from  string
	to    string
	value string
	nonce uint64
	gasPrice string
	gasLimit string
}

//构建gas估算入参
func (wm *WalletManager) CreatestimateGasParameters(from string, to string, value *big.Int) (*estimateGasParameter,error){

	Parameter := &estimateGasParameter{
		from : from,
		to   : to,
		value: value.String(),
		gasLimit: Gaslimit,
	}

	nonce,err := wm.WalletClient.CallGetaccountstate(from , "nonce")
	if err != nil{
		return nil,err
	}
	Nonce,_ :=strconv.ParseUint(nonce, 10, 64)
	Parameter.nonce = Nonce

	gasPrice := wm.EstimateFeeRate()
	Parameter.gasPrice = gasPrice

	return Parameter ,err
}

//计算花费gas所用的Wei = gasuse * gasprice
func (feeinfo *txFeeInfo) CalcFee() error {
	fee := new(big.Int)
	fee.Mul(feeinfo.GasUse, feeinfo.GasPrice)
	feeinfo.Fee = fee
	return nil
}

//估算gas花费的Wei
func (wm *WalletManager) Getestimatefee(from string, to string, value *big.Int) (*txFeeInfo, error) {

	Parameter,err := wm.CreatestimateGasParameters(from,to,value)
	if err != nil{
		return nil,err
	}

	result ,err := wm.WalletClient.CallGetestimateGas(Parameter)
	if err != nil{
		return nil,err
	}

	EstimateGas, err := ConvertToBigInt(result.Get("gas").String())
	if err != nil{
		return nil,err
	}

	GasPrice, err := ConvertToBigInt(Parameter.gasPrice)
	if err != nil{
		return nil,err
	}

	estimatefee := &txFeeInfo{
		GasUse: EstimateGas,
		GasPrice: GasPrice,
	}
	estimatefee.CalcFee()

	return estimatefee, nil
}