package ethereum

import (
	"errors"
	"fmt"

	"github.com/blocktree/OpenWallet/keystore"
	"github.com/btcsuite/btcutil/hdkeychain"
)

//CreateNewWallet 创建钱包
//过时的函数,不推荐使用
func (this *WalletManager) CreateNewWallet(name, password string) (*Wallet, string, error) {

	//检查钱包名是否存在
	wallets, err := GetWalletKeys(this.GetConfig().KeyDir)
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

	key, keyFile, err := keystore.StoreHDKeyWithSeed(this.GetConfig().KeyDir, name, password, extSeed, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, "", err
	}

	w := Wallet{WalletID: key.RootId, Alias: key.Alias}

	return &w, keyFile, nil
}

//HDKey 获取钱包密钥，需要密码
//已经过期, 不推荐使用
func (w *Wallet) HDKey(password string, s *keystore.HDKeystore) (*keystore.HDKey, error) {
	fmt.Println("w.KeyFile:", w.KeyFile)
	key, err := s.GetKey(w.WalletID, w.KeyFile, password)
	if err != nil {
		return nil, err
	}
	return key, err
}
