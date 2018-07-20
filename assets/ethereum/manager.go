package ethereum

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/btcsuite/btcutil/hdkeychain"
	ethKStore "github.com/ethereum/go-ethereum/accounts/keystore"
)

const (
	maxAddresNum = 10000
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
		fmt.Println("loop to wallet balance")
		balance, err := GetWalletBalance(w)
		if err != nil {

			log.Fatal(fmt.Sprintf("find wallet balance failed, err=%v\n", err))
			return nil, err
		}
		w.Balance = balance.String()
	}

	return wallets, nil
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
			w.Balance = balance.String()
			return w, nil
		}

	}

	return nil, errors.New("The wallet that your given name is not exist!")
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
		Address:   keyCombo.Address.String(), //address.String(),
		Account:   parentKey.RootId,
		HDPath:    derivedPath,
		Balance:   "0",
		CreatedAt: time.Now(),
	}

	return keyCombo, &addr, err
}

func CreateBatchAddress(name, password string, count uint64) error {
	//读取钱包
	w, err := GetWalletInfo(name)
	if err != nil {
		log.Fatal(fmt.Printf("get wallet info, err=%v\n", err))
		return err
	}

	//加载钱包
	keyroot, err := w.HDKey(password)
	if err != nil {
		log.Fatal(fmt.Printf("get HDkey, err=%v\n", err))
		return err
	}

	timestamp := uint64(time.Now().Unix())

	db, err := w.OpenDB()
	if err != nil {
		log.Fatal(fmt.Printf("open db, err=%v\n", err))
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

func GetAddrBalance(address string) (*big.Int, error) {

	params := []interface{}{
		address,
		"latest",
	}
	result, err := client.Call("eth_getBalance", 1, params)
	if err != nil {
		log.Fatal(fmt.Sprintf("get addr[%v] balance failed, err=%v\n", address, err))
		return big.NewInt(0), err
	}
	if result.Type != gjson.String {
		log.Fatal(fmt.Sprintf("get addr[%v] balance format failed, response is %v\n", address, result.Type))
		return big.NewInt(0), err
	}

	balance := new(big.Int)
	resultStr := result.String()
	if strings.Index(resultStr, "0x") != -1 {
		fmt.Println("resultStr:", resultStr)
		resultStr = common.Substr(resultStr, 2, len(resultStr))
		fmt.Println("ater trim resultStr:", resultStr)
	}
	_, success := balance.SetString(resultStr, 16)
	if !success {
		log.Fatal(fmt.Sprintf("get addr[%v] balance format failed, response is %v\n", address, result.String()))
		return big.NewInt(0), err
	}
	return balance, nil
}

//金额的单位是wei
func GetWalletBalance(wallet *Wallet) (*big.Int, error) {

	db, err := wallet.OpenDB()
	if err != nil {
		return big.NewInt(0), err
	}
	defer db.Close()

	var addresses []*Address
	err = db.Find("Account", wallet.WalletID, &addresses)
	if err != nil && strings.Index(err.Error(), "not found") == -1 {
		return nil, err
	}

	balanceTotal := new(big.Int)
	for _, addr := range addresses {
		balance, err := GetAddrBalance(addr.Address)
		if err != nil {
			errinfo := fmt.Sprintf("get balance of addr[%v] failed, err=%v", addr.Address, err)
			return balanceTotal, errors.New(errinfo)
		}

		balanceTotal = balanceTotal.Add(balanceTotal, balance)
	}

	return balanceTotal, nil
}
