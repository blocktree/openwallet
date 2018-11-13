package tech

import (
	"encoding/json"
	"strconv"

	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/assets/ethereum"
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
	log.Debugf("amount :%v", data.Transaction.Amount)
	return nil
}

func TestSubscribe() {

	var (
		endRunning = make(chan bool, 1)
	)
	//   wm, _ := GetAssetsManager(symbol)
	sub := subscriber{}
	tm.AddObserver(&sub)
	//manager, _ := GetEthWalletManager()
	//manager.Blockscanner.AddAddress("0xe7134824df22750a42726483e64047ef652d6194", "openw:KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8")
	//manager.Blockscanner.AddAddress("0xdfe55e4f7c1f24a7d9b05a0ac39c0390eb918564", "openw:KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8")

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
	//manager.Blockscanner.AddAddress("0xe7134824df22750a42726483e64047ef652d6194", "openw:KaszkQZb2xsaNuW5UoAukhM5MhzAqtPBWYTwkk4m2QhtDYN9E8")

	err := manager.Blockscanner.ScanBlock(319482)
	if err != nil {
		log.Errorf("scan block[%v] failed, err=%v", 319482, err)
	}
}

func DumpBlockScanDb() {
	db, err := storm.Open("data/eth/db/blockchain.db")
	if err != nil {
		log.Errorf("open block scan db failed, err=%v", err)
		return
	}
	defer db.Close()

	var unscanTransactions []ethereum.UnscanTransaction

	var blockHeight uint64
	var blockHash string
	err = db.All(&unscanTransactions)
	if err != nil {
		log.Errorf("get transactions failed, err = %v", err)
		return
	}

	for i, _ := range unscanTransactions {
		objStr, _ := json.MarshalIndent(unscanTransactions[i], "", " ")
		log.Infof("unscanned tx[%v]:%v", unscanTransactions[i].TxID, string(objStr))
	}
	var blocks []ethereum.BlockHeader
	err = db.All(&blocks)
	if err != nil {
		log.Errorf("get blocks failed failed, err = %v", err)
		return
	}

	for i, _ := range blocks {
		objStr, _ := json.MarshalIndent(blocks[i], "", " ")
		log.Infof("block[%v]:%v", blocks[i].BlockNumber, string(objStr))
	}

	err = db.Get(ethereum.BLOCK_CHAIN_BUCKET, ethereum.BLOCK_HEIGHT_KEY, &blockHeight)
	if err != nil {
		log.Errorf("get block height from db failed, err=%v", err)
		return
	}

	err = db.Get(ethereum.BLOCK_CHAIN_BUCKET, ethereum.BLOCK_HASH_KEY, &blockHash)
	if err != nil {
		log.Errorf("get block height from db failed, err=%v", err)
		return
	}

	log.Infof("current block height:%v, current block hash:%v", "0x"+strconv.FormatUint(blockHeight, 16), blockHash)
}
