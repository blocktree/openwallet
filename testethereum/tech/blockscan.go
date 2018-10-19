package tech

import (
	"encoding/json"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
)

type subscriber struct {
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriber) BlockScanNotify(header *openwallet.BlockHeader) error {
	objStr, _ := json.MarshalIndent(header, "", " ")
	log.Debug("header:", string(objStr))
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriber) BlockTxExtractDataNotify(account *openwallet.AssetsAccount, data *openwallet.TxExtractData) error {
	objStr, _ := json.MarshalIndent(account, "", " ")
	log.Debug("account:", string(objStr))
	objStr, _ = json.MarshalIndent(data, "", " ")
	log.Debug("data:", string(objStr))
	return nil
}

func TestSubscribe() {

	var (
		endRunning = make(chan bool, 1)
	)
	//   wm, _ := GetAssetsManager(symbol)
	sub := subscriber{}
	tm.AddObserver(&sub)
	manager, _ := GetEthWalletManager()
	manager.Blockscanner.AddAddress("0xe7134824df22750a42726483e64047ef652d6194", "XXXXXXXXXXXXXXXXX")

	err := PrepareTestForBlockScan()

	if err != nil {
		log.Errorf("prepare block scan failed, err=%v", err)
		return
	}

	//tm.SetRescanBlockHeight("QTUM", 236098)
	//log.Debug("SupportAssets:", tm.cfg)
	<-endRunning
}

func TestScanBlockByHeight() {
	//   wm, _ := GetAssetsManager(symbol)
	sub := subscriber{}
	tm.AddObserver(&sub)
	manager, _ := GetEthWalletManager()
	manager.Blockscanner.AddAddress("0xe7134824df22750a42726483e64047ef652d6194", "openw:KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8")

	err := manager.Blockscanner.ScanBlock(319482)
	if err != nil {
		log.Errorf("scan block[%v] failed, err=%v", 319482, err)
	}
}
