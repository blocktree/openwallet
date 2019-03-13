package openw

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
)

func testONTInitWalletManager() *WalletManager {

	tc := NewConfig()

	tc.IsTestnet = true
	tc.EnableBlockScan = true
	tc.SupportAssets = []string{
		"ONT",
	}
	return NewWalletManager(tc)
	//tm.Init()
}

func TestONTWalletManager_CreateWallet(t *testing.T) {
	tm := testONTInitWalletManager()
	w := &openwallet.Wallet{Alias: "HELLO KITTY", IsTrust: true, Password: "12345678"}
	nw, key, err := tm.CreateWallet(testApp, w)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("wallet:", nw)
	fmt.Println("key:", key)
}

func TestONTWalletManager_ConcurrentCreateWallet(t *testing.T) {

	tm := testONTInitWalletManager()

	var wg sync.WaitGroup
	timestamp := time.Now().Unix()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				wid := fmt.Sprintf("w_%d_%d_%d", timestamp, id, j)
				w := &openwallet.Wallet{WalletID: wid, Alias: "bitbank", IsTrust: false, Password: "12345678"}
				_, _, err := tm.CreateWallet(testApp, w)
				if err != nil {
					log.Error("wallet[", id, "-", j, "] unexpected error:", err)
					continue
				}
			}

		}(i)

	}

	wg.Wait()

	tm.CloseDB(testApp)
}

func TestONTWalletManager_GetWalletInfo(t *testing.T) {

	tm := testONTInitWalletManager()

	wallet, err := tm.GetWalletInfo(testApp, "WNQH8mJYSFZ3RqXoiyxfRxznwaiQRBpRYj")
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	log.Info("wallet:", wallet)
}

func TestONTWalletManager_GetWalletList(t *testing.T) {

	tm := testONTInitWalletManager()

	list, err := tm.GetWalletList(testApp, 0, 10000000)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("wallet[", i, "] :", w)
	}
	log.Info("wallet count:", len(list))

	tm.CloseDB(testApp)
}

func TestONTWalletManager_CreateAssetsAccount(t *testing.T) {

	tm := testONTInitWalletManager()

	walletID := "WNQH8mJYSFZ3RqXoiyxfRxznwaiQRBpRYj"
	account := &openwallet.AssetsAccount{Alias: "HELLO KITTY", WalletID: walletID, Required: 1, Symbol: "ONT", IsTrust: true}
	account, address, err := tm.CreateAssetsAccount(testApp, walletID, "12345678", account, nil)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("account:", account)
	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func TestONTWalletManager_GetAssetsAccountList(t *testing.T) {

	tm := testONTInitWalletManager()

	walletID := "WNQH8mJYSFZ3RqXoiyxfRxznwaiQRBpRYj"
	list, err := tm.GetAssetsAccountList(testApp, walletID, 0, 10000000)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("account[", i, "] :", w)
	}
	log.Info("account count:", len(list))

	tm.CloseDB(testApp)

}

func TestONTWalletManager_CreateAddress(t *testing.T) {

	tm := testONTInitWalletManager()

	walletID := "WNQH8mJYSFZ3RqXoiyxfRxznwaiQRBpRYj"
	accountID := "7NHXSXEaBViL5koJgexd3qHeGLggHjen99nyjEigdGnn"
	address, err := tm.CreateAddress(testApp, walletID, accountID, 3)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func TestONTWalletManager_GetAddressList(t *testing.T) {

	tm := testONTInitWalletManager()

	walletID := "WNQH8mJYSFZ3RqXoiyxfRxznwaiQRBpRYj"
	accountID := "7NHXSXEaBViL5koJgexd3qHeGLggHjen99nyjEigdGnn"
	list, err := tm.GetAddressList(testApp, walletID, accountID, 0, -1, false)
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	for i, w := range list {
		log.Info("address[", i, "] :", w)
	}
	log.Info("address count:", len(list))

	tm.CloseDB(testApp)
}
