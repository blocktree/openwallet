package tech

import (
	"encoding/json"
	"path/filepath"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/manager"
	"github.com/blocktree/OpenWallet/openwallet"
	owcrypt "github.com/blocktree/go-OWCrypt"
	"github.com/bytom/common"
)

const (
	FRAME_DEFAULT_DIR = "frame_data/eth"
)

var (
	tm      = manager.NewWalletManager(NewEthTestConfig())
	testApp = "openw"
)

func NewEthTestConfig() *manager.Config {

	c := manager.Config{}

	//钥匙备份路径
	c.KeyDir = filepath.Join(FRAME_DEFAULT_DIR, "key")
	//本地数据库文件路径
	c.DBPath = filepath.Join(FRAME_DEFAULT_DIR, "db")
	//备份路径
	c.BackupDir = filepath.Join(FRAME_DEFAULT_DIR, "backup")
	//支持资产
	c.SupportAssets = []string{"ETH"}
	//开启区块扫描
	c.EnableBlockScan = true
	//测试网
	c.IsTestnet = true

	return &c
}

func TestFrameWalletManager_CreateWallet() {
	w := &openwallet.Wallet{Alias: "MAI", IsTrust: true, Password: "12345678"}
	nw, key, err := tm.CreateWallet(testApp, w)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("wallet:", nw)
	log.Info("key:", key)
}

func TestWalletManager_CreateAssetsAccount() {

	walletID := "VySSrRvfpuzwy5TdZeCLkSy6P2DGTh3MaD" //"WFPHAs2uyeHcfBzKF4vN4NkMpArX8wkCxp"
	account := &openwallet.AssetsAccount{Alias: "Alice", WalletID: walletID, Required: 1, Symbol: "ETH"}
	account, err := tm.CreateAssetsAccount(testApp, walletID, account, nil)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("account:", account)

	tm.CloseDB(testApp)
}

func TestFrameWalletManager_CreateAddress() {
	walletID := "W9Azzt5LAttoyaYufQHZKsbqkwmZbNPM95" //"WFPHAs2uyeHcfBzKF4vN4NkMpArX8wkCxp"
	//accountID := "KhJdnr4UJLdbeQcMvZgedyYykRVTdLaMLbsV2mx3GZiMva9Kfb"
	accountID := "JcSuy6pN2QijemvRh5xyuXAMaurXcU6RceyehPrQ8mjhM1RQ4x" //"L77QWWMRhsKMiArgaMTiVaxa6knz2Wo2eNk5F3Bw764XeDyq3T"
	address, err := tm.CreateAddress(testApp, walletID, accountID, 1)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func ListInfoFromDb() {
	db, err := tm.OpenDB(testApp)
	if err != nil {
		log.Debugf("open db failed, err=%v", err)
		return
	}

	var wallets []openwallet.Wallet
	log.Debugf("print wallet info in db:")
	err = db.All(&wallets)
	if err != nil {
		log.Debugf("get wallets failed,err=%v", err)
		return
	}

	walletsStr, _ := json.MarshalIndent(&wallets, "", " ")
	log.Debug(string(walletsStr))

	var accounts []openwallet.AssetsAccount
	log.Debugf("print account info in db:")
	err = db.All(&accounts)
	if err != nil {
		log.Debugf("get accounts failed,err=%v", err)
		return
	}

	accountsStr, _ := json.MarshalIndent(&accounts, "", " ")
	log.Debug(string(accountsStr))

	var addresses []openwallet.Address
	log.Debugf("print addresses info in db:")
	err = db.All(&addresses)
	if err != nil {
		log.Debugf("get addresses failed,err=%v", err)
		return
	}

	addressesStr, _ := json.MarshalIndent(&addresses, "", " ")
	log.Debug(string(addressesStr))
}

//wallet id:VySSrRvfpuzwy5TdZeCLkSy6P2DGTh3MaD
//password:12345678
//address: 13fdCB740f1C79A7018BF66E3Bc710DD36714a16
func GetPrivateKeyInWallet(walletID string, password string, address string) string {
	//walletID := "VySSrRvfpuzwy5TdZeCLkSy6P2DGTh3MaD"
	w, err := tm.GetWalletInfo(testApp, walletID)
	if err != nil {
		log.Debugf("CreateNewPrivateKey failed unexpected error: %v\n", err)
		return ""
	}

	key, err := w.HDKey(password)
	if err != nil {
		log.Debugf("CreateNewPrivateKey failed unexpected error: %v\n", err)
		return ""
	}

	fromAddr := w.GetAddress(address)
	if err != nil {
		log.Debugf("wallet[%v] get address failed, err = %v", w.WalletID, err)
		return ""
	}

	log.Debugf("fromAddr:%v", fromAddr)

	childKey, _ := key.DerivedKeyWithPath(fromAddr.HDPath, owcrypt.ECC_CURVE_SECP256K1)
	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		log.Debugf("get private key bytes, err=%v", err)
		return ""
	}
	prikeyStr := common.ToHex(keyBytes)
	log.Debugf("privateStr selfmade:%v", prikeyStr)
	return prikeyStr
}

func TestGetPrivateKey() {
	prikeyselfmade := GetPrivateKeyInWallet("VySSrRvfpuzwy5TdZeCLkSy6P2DGTh3MaD", "12345678", "13fdCB740f1C79A7018BF66E3Bc710DD36714a16")
	//prikeyeth := ExportPrivateKeyFromGeth("0x50068fd632c1a6e6c5bd407b4ccf8861a589e776", "123456")

	log.Debug("self made private key:%v", prikeyselfmade)
	//log.Debug("eth private key exported:%v", prikeyeth)
}
