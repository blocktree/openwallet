package tezos

import (
	"github.com/blocktree/OpenWallet/openwallet"
	"io/ioutil"
	"path/filepath"
	"os"
	"strings"
	"github.com/asdine/storm"
)

type Key struct {
	Address    string `storm:"id"`
	PublicKey  string
	PrivateKey string
}

func NewKey(addr, pub, priv string) *Key {
	return &Key{Address: addr, PublicKey: pub, PrivateKey: priv}
}

//DecryptPrivateKey 解密私钥
func (k *Key) DecryptPrivateKey() string {

	return ""
}

//SaveKeyToWallet 保存私钥给钱包数据库
func SaveKeyToWallet(wallet *openwallet.Wallet, key *Key) error {

	db, err := wallet.OpenDB()
	if err != nil {
		return err
	}
	db.Close()
	return db.Save(key)
}



/***** 钱包相关 *****/


//GetWalletKeys 通过给定的文件路径加载keystore文件得到钱包列表
func GetWallets() ([]*openwallet.Wallet, error) {

	var (
		wallets = make([]*openwallet.Wallet, 0)
	)

	//扫描key目录的所有钱包
	files, err := ioutil.ReadDir(dataDir)
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
		w, err := GetWalletByID(fi.Name())
		if err != nil {
			continue
		}
		wallets = append(wallets, w)

	}

	return wallets, nil

}

//GetWalletByID 获取钱包
func GetWalletByID(walletID string) (*openwallet.Wallet, error) {

	dbFile := filepath.Join(dbPath, walletID+".db")
	db, err := storm.Open(dbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var wallet openwallet.Wallet
	err = db.One("WalletID", walletID, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
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