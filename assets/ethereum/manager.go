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

//GetWalletList 获取钱包列表
func GetWalletList() ([]*Wallet, error) {

	wallets, err := GetWalletKeys(keyDir)
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		//fmt.Println("loop to wallet balance")
		balance, err := GetWalletBalance(w)
		if err != nil {

			openwLogger.Log.Errorf(fmt.Sprintf("find wallet balance failed, err=%v\n", err))
			return nil, err
		}
		w.Balance = balance
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

func BackupWalletToDefaultPath(wallet *Wallet) (string, error) {
	newBackupDir := filepath.Join(backupDir, wallet.FileName()+"-"+common.TimeFormat(TIME_POSTFIX))
	return BackupWallet(newBackupDir, wallet)
}

func BackupWallet(newBackupDir string, wallet *Wallet) (string, error) {
	/*w, err := GetWalletInfo(wallet.WalletID)
	if err != nil {
		return "", err
	}*/

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
			w.Balance = balance
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

func CreateBatchAddress(name, password string, count uint64) error {
	//读取钱包
	w, err := GetWalletInfo(name)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get wallet info, err=%v\n", err))
		return err
	}

	//加载钱包
	keyroot, err := w.HDKey(password)
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
		ethKeyStore.NewAccountForWalletBT(keyCombo, DefaultPasswordForEthKey)
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

type txFeeInfo struct {
	GasLimit *big.Int
	GasPrice *big.Int
	Fee      *big.Int
}

func GetTransactionFeeEstimated(from string, to string, amount *big.Int) (*txFeeInfo, error) {
	gasLimit, err := ethGetGasEstimated(from, to, amount)
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

func SendTransaction(wallet *Wallet, to string, amount *big.Int, password string, feesInSender bool) ([]string, error) {
	var txIds []string
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
		//如果该地址的余额足够支付转账
		if addrs[i].balance.Cmp(amount) >= 0 {
			amountToSend = *amount
			fee, err = GetTransactionFeeEstimated(addrs[i].Address, to, &amountToSend)
			if err != nil {
				openwLogger.Log.Errorf("%v", err)
				continue
			}

			balanceLeft := *addrs[i].balance
			balanceLeft.Sub(&balanceLeft, fee.Fee)

			//灰尘账户, 余额不足以发起一次transaction
			fmt.Println("amount to send ignore fee:", amountToSend.String())
			if balanceLeft.Cmp(big.NewInt(0)) < 0 {
				errinfo := fmt.Sprintf("[%v] is a dust address, will skip. /n", addrs[i].Address)
				openwLogger.Log.Errorf(errinfo)
				continue
			}

			//如果改地址的余额除去手续费后, 不足以支付转账, set 转账金额 = 账户余额 - 手续费
			if balanceLeft.Cmp(amount) < 0 {
				amountToSend = balanceLeft
				fmt.Println("amount to send plus fee:", amountToSend.String())
			}

		} else {
			amountToSend = *addrs[i].balance
			fee, err = GetTransactionFeeEstimated(addrs[i].Address, to, &amountToSend)
			if err != nil {
				openwLogger.Log.Errorf("%v", err)
				continue
			}

			//灰尘账户, 余额不足以发起一次transaction
			if amountToSend.Cmp(fee.Fee) <= 0 {
				errinfo := fmt.Sprintf("[%v] is a dust address, will skip. /n", addrs[i].Address)
				openwLogger.Log.Errorf(errinfo)
				continue
			}

			fmt.Println("amount to send without fee, ", amountToSend.String(), " , fee:", fee.Fee.String())
			amountToSend.Sub(&amountToSend, fee.Fee)
			fmt.Println("amount to send applied fee, ", amountToSend.String())
		}

		txid, err := SendTransactionToAddr(addrs[i], to, &amountToSend, password, fee)
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

func SendTransactionToAddr(addr *Address, to string, amount *big.Int, password string, fee *txFeeInfo) (string, error) {
	err := UnlockAddr(addr.Address, password, 300)
	if err != nil {
		openwLogger.Log.Errorf("unlock addr failed, err = %v", err)
		return "", err
	}

	txId, err := ethSendTransaction(addr.Address, to, amount, fee)
	if err != nil {
		openwLogger.Log.Errorf("ethSendTransaction failed, err = %v", err)
		return "", err
	}

	err = LockAddr(addr.Address)
	if err != nil {
		openwLogger.Log.Errorf("lock addr failed, err = %v", err)
		return txId, err
	}

	return txId, nil
}

func convertToBigInt(value string, base int) (*big.Int, error) {
	bigvalue := new(big.Int)
	var success bool
	if base == 16 {
		if strings.Index(value, "0x") != -1 {
			value = common.Substr(value, 2, len(value))
		}
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

	gasLimit, err := convertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert estimated gas[%v] format to bigint failed, err = %v\n", result.String(), err)
		openwLogger.Log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return gasLimit, nil
}

func ethGetGasEstimated(fromAddr string, toAddr string, amount *big.Int) (*big.Int, error) {
	trans := make(map[string]interface{})
	trans["from"] = fromAddr
	trans["to"] = toAddr
	trans["value"] = "0x" + amount.Text(16)

	params := []interface{}{
		trans,
	}

	result, err := client.Call("eth_estimateGas", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get estimated gas limit from [%v] to [%v] faield, err = %v \n", fromAddr, toAddr, err))
		return big.NewInt(0), err
	}

	if result.Type != gjson.String {
		openwLogger.Log.Errorf(fmt.Sprintf("get estimated gas from [%v] to [%v] failed, response is %v\n", fromAddr, toAddr, err))
		return big.NewInt(0), err
	}

	gasLimit, err := convertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert estimated gas[%v] format to bigint failed, err = %v\n", result.String(), err)
		openwLogger.Log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return gasLimit, nil
}

/*
params: [{
  "from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
  "to": "0xd46e8dd67c5d32be8058bb8eb970870f07244567",
  "gas": "0x76c0", // 30400
  "gasPrice": "0x9184e72a000", // 10000000000000
  "value": "0x9184e72a", // 2441406250
  "data": "0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"
}]
*/

func ethSendTransaction(fromAddr string, toAddr string, amount *big.Int, fee *txFeeInfo) (string, error) {
	trans := make(map[string]interface{})
	trans["from"] = fromAddr
	trans["to"] = toAddr
	trans["value"] = "0x" + amount.Text(16)
	trans["gas"] = "0x" + fee.GasLimit.Text(16)
	trans["gasPrice"] = "0x" + fee.GasPrice.Text(16)

	params := []interface{}{
		trans,
	}

	result, err := client.Call("eth_sendTransaction", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("start transaction from [%v] to [%v] faield, err = %v \n", fromAddr, toAddr, err))
		return "", err
	}

	if result.Type != gjson.String {
		openwLogger.Log.Errorf(fmt.Sprintf("send transaction from [%v] to [%v] failed, response is %v\n", fromAddr, toAddr, err))
		return "", err
	}
	return result.String(), nil
}

func GetAddrBalance(address string) (*big.Int, error) {

	params := []interface{}{
		address,
		"latest",
	}
	result, err := client.Call("eth_getBalance", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get addr[%v] balance failed, err=%v\n", address, err))
		return big.NewInt(0), err
	}
	if result.Type != gjson.String {
		openwLogger.Log.Errorf(fmt.Sprintf("get addr[%v] balance format failed, response is %v\n", address, result.Type))
		return big.NewInt(0), err
	}

	/*balance := new(big.Int)
	resultStr := result.String()
	if strings.Index(resultStr, "0x") != -1 {
		//fmt.Println("resultStr:", resultStr)
		resultStr = common.Substr(resultStr, 2, len(resultStr))
		//fmt.Println("ater trim resultStr:", resultStr)
	}
	_, success := balance.SetString(resultStr, 16)*/
	balance, err := convertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert addr[%v] balance format to bigint failed, response is %v, and err = %v\n", address, result.String(), err)
		openwLogger.Log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return balance, nil
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
		fmt.Printf("addr[%v] : %v\n", addr.Address, addr.balance)
		balanceTotal = balanceTotal.Add(balanceTotal, addr.balance)
	}

	return balanceTotal, nil
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

			txId, err := SendTransaction(wallet, sumAddress, balance, DefaultPasswordForEthKey, true)
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
func RestoreWallet(keyFile string) error {
	fmt.Printf("Validating key file... \n")

	finfo, err := os.Stat(keyFile)
	if err != nil || !finfo.IsDir() {
		errinfo := fmt.Sprintf("stat file[%v] failed, err = %v\n", keyFile, err)
		openwLogger.Log.Errorf(errinfo)
		return err
	}
	parts := filepath.SplitList(keyFile)
	l := len(parts)
	if l == 0 {
		errinfo := fmt.Sprintf("wrong keyFile[%v] passed through...", keyFile)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	}

	dirName := parts[l-1]
	parts = strings.Split(dirName, "-")
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
	wallet, err := GetWalletInfo(walletId)
	if err != nil && err.Error() != WALLET_NOT_EXIST_ERR {
		errinfo := fmt.Sprintf("get wallet [%v] info failed, err = %v ", walletId, err)
		openwLogger.Log.Errorf(errinfo)
		return errors.New(errinfo)
	} else if err == nil {
		newBackupDir := filepath.Join(backupDir+"/restore", wallet.FileName()+"-"+common.TimeFormat(TIME_POSTFIX))
		_, err := BackupWallet(newBackupDir, wallet)
		if err != nil {
			errinfo := fmt.Sprintf("backup wallet[%v] before restore failed,err = %v ", wallet.WalletID, err)
			openwLogger.Log.Errorf(errinfo)
			return errors.New(errinfo)
		}
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
