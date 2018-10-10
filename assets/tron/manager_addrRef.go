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

package tron

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCrypt"
	"github.com/bndr/gotabulate"
	"github.com/shengdoushi/base58"
)

func createAddressByPkRef(pubKey []byte) (addrBytes []byte, err error) {
	// First: calculate sha3-256 of PublicKey, get Hash as pkHash
	pkHash := owcrypt.Hash(pubKey, 0, owcrypt.HASH_ALG_KECCAK256)[12:32]
	// Second: expend 0x41 as prefix of pkHash to mark Tron
	address := append([]byte{0x41}, pkHash...)
	// Third: double sha256 to generate Checksum
	sha256_0_1 := owcrypt.Hash(address, 0, owcrypt.HASh_ALG_DOUBLE_SHA256)
	// Fourth: Append checksum to pkHash from sha256_0_1 with the last 4
	addrBytes = append(address, sha256_0_1[0:4]...)

	return addrBytes, nil
}

// Done
// Function: Create address from a specified private key string
func (wm *WalletManager) CreateAddressRef(privateKey string) (addrBase58 string, err error) {

	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		log.Println(err)
		return "", err
	}

	pubKey, res := owcrypt.GenPubkey(privateKeyBytes, owcrypt.ECC_CURVE_SECP256K1)
	if res != owcrypt.SUCCESS {
		err := errors.New("Error from owcrypt.GenPubkey: failed!")
		log.Println(err)
		return "", err
	}

	if address, err := createAddressByPkRef(pubKey); err != nil {
		return "", err
	} else {
		// Last: encoding with Base58(alphabet use BitcoinAlphabet)
		addrBase58 = base58.Encode(address, base58.BitcoinAlphabet)
	}

	return addrBase58, nil
}

// Done
func (wm *WalletManager) ValidateAddressRef(addrBase58 string) (err error) {

	addressBytes, err := base58.Decode(addrBase58, base58.BitcoinAlphabet)
	if err != nil {
		return err
	}

	l := len(addressBytes)
	addressBytes, checksum := addressBytes[:l-4], addressBytes[l-4:]
	sha256_0_1 := owcrypt.Hash(addressBytes, 0, owcrypt.HASh_ALG_DOUBLE_SHA256)

	if hex.EncodeToString(sha256_0_1[0:4]) != hex.EncodeToString(checksum) {
		return errors.New("Address invalid!")
	}

	return nil
}

// -------------------------------------------------------------------------------------------------------------------------------
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
		log.Println(err)
		return "", nil, err
	}

	//加载钱包
	key, err := w.HDKey(password)
	if err != nil {
		log.Println(err)
		return "", nil, err
	}

	timestamp := time.Now()
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
			fmt.Printf("Start create address thread[%d]\n", i+1)
			s := i * runCount
			e := (i + 1) * runCount
			go wm.createAddressWork(key, producer, name, uint64(timestamp.Unix()), s, e)

			shouldDone++
		}
	}

	if otherCount > 0 {

		//开始创建地址
		// log.Std.Info("Start create address thread[REST]")
		fmt.Println("Start create address thread[REST]")
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
			fmt.Printf("\tcompleted %d \n", len(pa))
			//当激活消费者后，传输数据给消费者，并把顶部数据出队
		case activeWorker <- activeValue:
			//log.Std.Info("Get %d", len(activeValue))
			fmt.Printf("\tExport to file: %d\n", len(activeValue))
			values = values[1:]

		case <-quit:
			//退出
			log.Println("\tAll addresses have been created!")
			return filePath, outputAddress, nil
		}
	}

	// wm.LockWallet()

	return filePath, outputAddress, nil
	// return []*openwallet.Address{}, nil
}

//createAddressWork 创建地址过程
func (wm *WalletManager) createAddressWork(k *hdkeystore.HDKey, producer chan<- []*openwallet.Address, walletID string, index, start, end uint64) {

	fmt.Printf("createAddressWork: index=%d, start=%d, end=%d \n", index, start, end)

	runAddress := make([]*openwallet.Address, 0)

	derivedPath := fmt.Sprintf("%s/%d", k.RootPath, index)
	childKey, err := k.DerivedKeyWithPath(derivedPath, wm.Config.CurveType)
	if err != nil {
		producer <- make([]*openwallet.Address, 0)
		return
	}

	// Generate address
	for i := start; i < end; i++ {
		childKey, err := childKey.GenPrivateChild(uint32(i))
		if err != nil {
			log.Println(err)
			return
		}

		keyBytes, err := childKey.GetPrivateKeyBytes()
		if err != nil {
			log.Println(err)
			return
		}

		addrBase58, err := wm.CreateAddressRef(hex.EncodeToString(keyBytes))
		if err != nil {
			log.Println(err)
			return
		}

		address := &openwallet.Address{
			Address:   addrBase58,
			AccountID: k.KeyID,
			HDPath:    fmt.Sprintf("%s/%d", derivedPath, i),
			CreatedAt: time.Now().Unix(),
			Symbol:    wm.Config.Symbol,
			Index:     index,
			WatchOnly: false,
		}

		runAddress = append(runAddress, address)
	}

	//生成完成
	producer <- runAddress

	fmt.Println("Producer done!")
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
		// err = db.Find("AccountID", walletID, &addresses, storm.Limit(limit), storm.Skip(offset))
		query := db.Select().Limit(limit).Skip(offset).OrderBy("Index", "HDPath")
		err = query.Find(&addresses)
	} else {
		// err = db.Find("AccountID", walletID, &addresses, storm.Skip(offset))
		query := db.Select().Limit(-1).Skip(0).Reverse().OrderBy("Index", "HDPath")
		err = query.Find(&addresses)
	}

	if err != nil {
		return nil, err
	}

	return addresses, nil

}

//exportAddressToFile 导出地址到文件中
func (wm *WalletManager) exportAddressToFile(addrs []*openwallet.Address, filePath string) {

	var (
		content string
	)

	for _, a := range addrs {

		// fmt.Printf("Export: %+v \n", a.Address)
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

/* -------------------------------------------------------------------------------------------------------------- */
// type Address struct {
// 	AccountID string `json:"accountID" storm:"index"` //钱包ID
// 	Address   string `json:"address" storm:"id"`      //地址字符串
// 	PublicKey string `json:"publicKey"`               //地址公钥/赎回脚本
// 	Alias     string `json:"alias"`                   //地址别名，可绑定用户
// 	Tag       string `json:"tag"`                     //标签
// 	Index     uint64 `json:"index"`                   //账户ID，索引位
// 	HDPath    string `json:"hdPath"`                  //地址公钥根路径
// 	WatchOnly bool   `json:"watchOnly"`               //是否观察地址，true的时候，Index，RootPath，Alias都没有。
// 	Symbol    string `json:"symbol"`                  //币种类别
// 	Balance   string `json:"balance"`                 //余额
// 	IsMemo    bool   `json:"isMemo"`                  //是否备注
// 	Memo      string `json:"memo"`                    //备注
// 	CreatedAt int64  `json:"createdAt"`               //创建时间
// 	IsChange  bool   `json:"isChange"`                //是否找零地址
// 	ExtParam  string `json:"extParam"`                //扩展参数，用于调用智能合约，json结构

// 	//核心地址指针
// 	Core interface{}
// }

//打印地址列表
func (wm *WalletManager) printAddressList(list []*openwallet.Address) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		// a.Balance = wm.GetWalletBalance(a.AccountID)  ?500
		tableInfo = append(tableInfo, []interface{}{
			i, w.AccountID, w.Address, w.Alias, w.PublicKey, w.Index, w.Memo, w.HDPath, w.IsChange, w.ExtParam,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "AccountID", "Address", "Alias", "PublicKey", "Index", "Momo", "HDPath", "IsChange", "Extparam"})

	//打印信息
	fmt.Println(t.Render("simple"))

}
