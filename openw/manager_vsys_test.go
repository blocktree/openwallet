package openw

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
)

func testVSYSInitWalletManager() *WalletManager {

	tc := NewConfig()

	tc.EnableBlockScan = true
	tc.SupportAssets = []string{
		"VSYS",
	}
	return NewWalletManager(tc)
	//tm.Init()
}

func TestVSYSWalletManager_CreateWallet(t *testing.T) {
	tm := testVSYSInitWalletManager()
	w := &openwallet.Wallet{Alias: "HELLO KITTY", IsTrust: true, Password: "12345678"}
	nw, key, err := tm.CreateWallet(testApp, w)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("wallet:", nw)
	fmt.Println("key:", key)
}

func TestVSYSWalletManager_ConcurrentCreateWallet(t *testing.T) {

	tm := testVSYSInitWalletManager()

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

func TestVSYSWalletManager_GetWalletInfo(t *testing.T) {

	tm := testVSYSInitWalletManager()

	wallet, err := tm.GetWalletInfo(testApp, "W8325QfzEfWq4uevrVh67wMR5xLDMEjiD7")
	if err != nil {
		log.Error("unexpected error:", err)
		return
	}
	log.Info("wallet:", wallet)
}

func TestVSYSWalletManager_GetWalletList(t *testing.T) {

	tm := testVSYSInitWalletManager()

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

func TestVSYSWalletManager_CreateAssetsAccount(t *testing.T) {

	tm := testVSYSInitWalletManager()

	walletID := "W8325QfzEfWq4uevrVh67wMR5xLDMEjiD7"
	account := &openwallet.AssetsAccount{Alias: "HELLO KITTY", WalletID: walletID, Required: 1, Symbol: "VSYS", IsTrust: true}
	account, address, err := tm.CreateAssetsAccount(testApp, walletID, "12345678", account, nil)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("account:", account)
	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func TestVSYSWalletManager_GetAssetsAccountList(t *testing.T) {

	tm := testVSYSInitWalletManager()

	walletID := "WJmcRGkdbT4AoSD7YUHBEVXS7zUmBFgo2g"
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

func TestVSYSWalletManager_CreateAddress(t *testing.T) {

	tm := testVSYSInitWalletManager()

	walletID := "WJmcRGkdbT4AoSD7YUHBEVXS7zUmBFgo2g"
	accountID := "A7X5LL16KDafDdEp9ioyzW8K8R2Edu25YbKxMkKCrimB"
	address, err := tm.CreateAddress(testApp, walletID, accountID, 3)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func TestVSYSWalletManager_GetAddressList(t *testing.T) {

	tm := testVSYSInitWalletManager()

	walletID := "W8325QfzEfWq4uevrVh67wMR5xLDMEjiD7"
	accountID := "GqTrUZF2dFcmo4rAksY2U3SPZABbF4ZDGMgKU6iVAXuU"
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
