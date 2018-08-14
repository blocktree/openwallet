package ethereum

import (
	"fmt"
	"math/big"
	"path/filepath"
	"time"

	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/keystore"
)

type Wallet struct {
	WalletID string   `json:"rootid"`
	Alias    string   `json:"alias"`
	Balance  *big.Int //string `json:"balance"`
	Password string   `json:"password"`
	RootPub  string   `json:"rootpub"`
	KeyFile  string
}

type Address struct {
	Address   string   `json:"address" storm:"id"`
	Account   string   `json:"account" storm:"index"`
	HDPath    string   `json:"hdpath"`
	balance   *big.Int //string `json:"balance"`
	CreatedAt time.Time
}

//HDKey 获取钱包密钥，需要密码
func (w *Wallet) HDKey(password string, s *keystore.HDKeystore) (*keystore.HDKey, error) {
	fmt.Println("w.KeyFile:", w.KeyFile)
	key, err := s.GetKey(w.WalletID, w.KeyFile, password)
	if err != nil {
		return nil, err
	}
	return key, err
}

//openDB 打开钱包数据库
func (w *Wallet) OpenDB() (*storm.DB, error) {
	file.MkdirAll(dbPath)
	file := w.DBFile()
	fmt.Println("dbpath:", dbPath, ", file:", file)
	return storm.Open(file)
}

func (w *Wallet) OpenDbByPath(path string) (*storm.DB, error) {
	return storm.Open(path)
}

//DBFile 数据库文件
func (w *Wallet) DBFile() string {
	return filepath.Join(dbPath, w.FileName()+".db")
}

//FileName 该钱包定义的文件名规则
func (w *Wallet) FileName() string {
	return w.Alias + "-" + w.WalletID
}
