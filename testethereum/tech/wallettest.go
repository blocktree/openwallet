package tech

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	//"log"
	"math/big"
	"strconv"

	"github.com/blocktree/OpenWallet/assets/ethereum"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/go-OWCrypt"
	"github.com/bytom/common"
	"github.com/ethereum/go-ethereum/accounts"
	ethKStore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/rlp"

	//crypto
	"github.com/ethereum/go-ethereum/crypto"
)

func TestNewWallet() {
	manager, _ := GetEthWalletManager()
	err := manager.CreateWalletFlow()
	if err != nil {
		log.Error("create wallet failed, err=", err)
	}
}

func TestBatchCreateAddr() {
	manager, _ := GetEthWalletManager()
	err := manager.CreateAddressFlow()
	if err != nil {
		log.Error("create wallet failed, err=", err)
	}
}

func TestBitInt() {
	i := new(big.Int)
	_, success := i.SetString("ff", 16)
	if success {
		fmt.Println("i:", i.String())
	}
}

func TestTransferFlow() {
	manager, _ := GetEthWalletManager()
	err := manager.TransferFlow()
	if err != nil {
		log.Debug("transfer flow failed, err = ", err)
	}
}

func TestSummaryFlow() {
	manager, _ := GetEthWalletManager()

	err := manager.SummaryFollow()
	if err != nil {
		log.Debug("summary flow failed, err = ", err)
	}
}

func TestBackupWallet() {
	manager, _ := GetEthWalletManager() //manager := &ethereum.WalletManager{}

	err := manager.BackupWalletFlow()
	if err != nil {
		log.Debug("backup wallet flow failed, err = ", err)
	}
}

func TestRestoreWallet() {
	manager, _ := GetEthWalletManager()

	err := manager.RestoreWalletFlow()
	if err != nil {
		log.Debug("restore wallet flow failed, err = ", err)
	}
}

func TestConfigErcToken() {
	manager, _ := GetEthWalletManager()

	err := manager.ConfigERC20Token()
	if err != nil {
		log.Debug("config erc20 token failed, err = ", err)
	}
}

func TestERC20TokenTransfer() {
	manager, _ := GetEthWalletManager()

	err := manager.ERC20TokenTransferFlow()
	if err != nil {
		log.Debug("transfer erc20 token failed, err = ", err)
	}
}

func TestERC20TokenSummary() {
	manager, _ := GetEthWalletManager()

	err := manager.ERC20TokenSummaryFollow()
	if err != nil {
		log.Debug("summary erc20 token failed, err = ", err)
	}
}

func TestErc20GetEventLog() {
	manager, _ := GetEthWalletManager()

	//0x2090475a6ab5b302dffa07239543cd6b4262a0df85b40ce1a75dd1dca83f0431
	//0x05ca0b5078926148ccb7bea73cffb56dc310a3fdb66fdd71274f9309beec9921
	event, err := manager.GetErc20TokenEvent("0x2090475a6ab5b302dffa07239543cd6b4262a0df85b40ce1a75dd1dca83f0431")
	if err != nil {
		log.Error("Get erc20 event failed, err = ", err)
		return
	}

	eventStr, _ := json.MarshalIndent(event, "", " ")
	log.Debugf("event:%v", string(eventStr))
}

func PrepareTestForBlockScan() error {
	/*pending, queued, err := ethereum.EthGetTxpoolStatus()
	if err != nil {
		log.Debugf("get txpool status failed, err=", err)
		return
	}
	fmt.Println("pending number is ", pending, " queued number is ", queued)*/
	manager, _ := GetEthWalletManager()
	type txsample struct {
		From            string
		To              string
		ContractAddress string
	}

	//DumpWalletDB("data/eth/db", "peter2-W4WUFawFp6FbzunjztXDDaN4nsdW8U4PrN.db")
	simpleTxs := []txsample{
		{
			From: "0xe7134824df22750a42726483e64047ef652d6194",
			To:   "0xf1320dcbe77711745554a18a572418537328d5b4",
		},
	}

	tokenTxs := []txsample{
		{
			From:            "0xe7134824df22750a42726483e64047ef652d6194",
			To:              "0xf1320dcbe77711745554a18a572418537328d5b4",
			ContractAddress: "0x8847E5F841458ace82dbb0692C97115799fe28d3",
		},
	}

	var txs []string
	for i, _ := range simpleTxs {
		tx, err := manager.SimpleSendEthTransaction(simpleTxs[i].From, simpleTxs[i].To, "1",
			"data/eth/db/peter2-W4WUFawFp6FbzunjztXDDaN4nsdW8U4PrN.db", "12345678", 12)
		if err != nil {
			log.Errorf("send tx failed, err=%v", err)
			continue
		}
		txs = append(txs, tx)
		log.Infof("sent tx[%v] successfully.", tx)
	}

	for i, _ := range tokenTxs {
		tx, err := manager.SimpleSendEthTokenTransaction(simpleTxs[i].From, simpleTxs[i].To, "1",
			"data/eth/db/peter2-W4WUFawFp6FbzunjztXDDaN4nsdW8U4PrN.db", "12345678", 12, "0x8847E5F841458ace82dbb0692C97115799fe28d3")
		if err != nil {
			log.Errorf("send token tx failed, err=%v", err)
			continue
		}
		txs = append(txs, tx)
		log.Infof("sent token tx[%v] successfully.", tx)
	}
	manager.WaitTxpoolPendingEmpty()
	log.Info("transaction sent out. txs:", txs)

	return nil
}

func TestDbInf() error {
	manager := &ethereum.WalletManager{}
	wallets, err := manager.GetLocalWalletList(manager.GetConfig().KeyDir, manager.GetConfig().DbPath, true)
	if err != nil {
		fmt.Println("get Wallet list failed, err=", err)
		return err
	}

	if len(wallets) == 0 {
		fmt.Println("no wallet found.")
		return err
	}
	wallets[len(wallets)-1].DumpWalletDB(manager.GetConfig().DbPath)
	DumpBlockScanDb()
	return nil
}

func TestBlockScanWhenFork() error {
	//ethereum.OpenDB(ethereum.)

	db, err := ethereum.OpenDB("/Users/peter/workspace/bitcoin/wallet/src/github.com/blocktree/OpenWallet/test/data/eth/db", "blockchain.db")
	if err != nil {
		fmt.Println("open eth block scan db failed, err=", err)
		return err
	}

	//手动修改block的hash,
	blocknums := []string{
		"0x2a19f",
		"0x2a19e",
		"0x2a19d",
	}

	for i, _ := range blocknums {
		var theBlocks []ethereum.BlockHeader
		err = db.Find("BlockNumber", blocknums[i], &theBlocks)
		if err != nil {
			fmt.Println("find block bumber failed, err=", err)
			return err
		}

		for j, _ := range theBlocks {
			theBlocks[j].BlockHash = "123456"
			err = db.Update(&theBlocks[j])
			if err != nil {
				fmt.Println("update block bumber failed, err=", err)
				return err
			}
		}
	}

	db.Close()

	manager := &ethereum.WalletManager{}
	scanner := ethereum.NewETHBlockScannerOld(manager)
	wallets, err := manager.GetLocalWalletList(manager.GetConfig().KeyDir, manager.GetConfig().DbPath, true)
	if err != nil {
		fmt.Println("get Wallet list failed, err=", err)
		return err
	}

	if len(wallets) == 0 {
		fmt.Println("no wallet found.")
		return err
	}

	w := wallets[len(wallets)-1]
	err = scanner.AddWallet(w.WalletID, w)
	if err != nil {
		fmt.Println("scanner add wallet failed, err=", err)
		return err
	}

	scanner.ScanBlock()
	fmt.Println("after scan block, show db following:")
	w.DumpWalletDB(manager.GetConfig().DbPath)
	DumpBlockScanDb()
	return nil
}

func TestBlockScan() error {
	manager := &ethereum.WalletManager{}
	scanner := ethereum.NewETHBlockScannerOld(manager)
	fromAddrs := make([]string, 0, 2)
	passwords := make([]string, 0, 2)
	fromAddrs = append(fromAddrs, "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776")
	passwords = append(passwords, "123456")
	fromAddrs = append(fromAddrs, "0x2a63b2203955b84fefe52baca3881b3614991b34")
	passwords = append(passwords, "123456")
	beginBlockNum, err := scanner.PrepareForBlockScanTest(fromAddrs, passwords)
	if err != nil {
		fmt.Println("PrepareForBlockScanTest failed, err=", err)
		return err
	}

	//manager := &ethereum.WalletManager{}
	//scanner := ethereum.NewETHBlockScannerOld(manager)
	wallets, err := manager.GetLocalWalletList("/Users/peter/workspace/bitcoin/wallet/src/github.com/blocktree/OpenWallet/test/data/eth/db", "blockchain.db", true)
	if err != nil {
		fmt.Println("get Wallet list failed, err=", err)
		return err
	}

	if len(wallets) == 0 {
		fmt.Println("no wallet found.")
		return err
	}

	w := wallets[len(wallets)-1]
	err = scanner.AddWallet(w.WalletID, w)
	if err != nil {
		fmt.Println("scanner add wallet failed, err=", err)
		return err
	}

	w.ClearAllTransactions("/Users/peter/workspace/bitcoin/wallet/src/github.com/blocktree/OpenWallet/test/data/eth/db")

	manager.ClearBlockScanDb()
	scanner.SetLocalBlock(beginBlockNum)
	scanner.ScanBlock()
	fmt.Println("after scan block, show db following:")
	w.DumpWalletDB("/Users/peter/workspace/bitcoin/wallet/src/github.com/blocktree/OpenWallet/test/data/eth/db")
	DumpBlockScanDb()
	return nil
}

func TestAddr() {
	addr := ethcommon.HexToAddress("0x45990739752539ae4C5DA39442177466292096eB")
	fmt.Println("addr str:", addr.String())
}

func TestOWCrypt_sign() {
	manager := &ethereum.WalletManager{}
	ethKeyStore := ethKStore.NewKeyStore(manager.GetConfig().EthereumKeyPath, ethKStore.StandardScryptN, ethKStore.StandardScryptP)
	from := ethcommon.HexToAddress("0x50068fd632c1a6e6c5bd407b4ccf8861a589e776")
	a := accounts.Account{Address: from}
	a, key, err := ethKeyStore.GetDecryptedKeyForOpenWallet(a, "123456")
	if err != nil {
		fmt.Println("get decrypted key failed, err= ", err)
		return
	}

	amount, err := ethereum.ConvertToBigInt("0x56bc75e2d63100000", 16)
	if err != nil {
		fmt.Println("amount format error.")
		return
	}

	gasPrice, err := ethereum.ConvertToBigInt("0x430e23400", 16)
	if err != nil {
		fmt.Println("gas price format error.")
		return
	}

	tx := types.NewTransaction(5, ethcommon.HexToAddress("0x2a63b2203955b84fefe52baca3881b3614991b34"),
		amount, 121000, gasPrice, nil)
	signer := types.NewEIP155Signer(big.NewInt(12))
	message := signer.Hash(tx)
	seckey := math.PaddedBigBytes(key.PrivateKey.D, key.PrivateKey.Params().BitSize/8)

	sig, ret := owcrypt.ETHsignature(seckey, message[:])
	if ret != owcrypt.SUCCESS {
		fmt.Println("signature error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		return
	}

	toPublicKey := func(pk *ecdsa.PublicKey) []byte {
		testByteX := pk.X.Bytes() //[]byte(*pk.X)
		testByteY := pk.Y.Bytes() //[]byte(*pk.X)
		return append(testByteX, testByteY...)
	}

	ret = owcrypt.Verify(toPublicKey(&key.PrivateKey.PublicKey), nil, 0, message[:], 32, sig, owcrypt.ECC_CURVE_SECP256K1|owcrypt.HASH_OUTSIDE_FLAG)
	if ret != owcrypt.SUCCESS {
		fmt.Println("verify error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		return
	}

	tx, err = tx.WithSignature(signer, sig)
	if err != nil {
		fmt.Println("with signature failed, err=", err)
	}

	tx.PrintTransaction()

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		fmt.Println("EncodeToBytes failed, err = ", err)
		return
	}

	fmt.Println("signature:", common.ToHex(data))
}

/*
 web3.eth.signTransaction({
    from: "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776",
    to: '0x2a63b2203955b84fefe52baca3881b3614991b34',
    value: "0x56bc75e2d63100000",
    data: "",
    gas: "0x1d8a8",
    gasPrice: "0x430e23400",
    nonce:"0x5"
})
*/

func TestGetNonce() {
	addr := "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776"
	manager := &ethereum.WalletManager{}
	nonce, err := manager.GetNonceForAddress(addr)
	if err != nil {
		fmt.Printf("get nonce for address[%v] failed, err=%v\n", addr, err)
		return
	}
	fmt.Println("the nonce is ", nonce)
}

func TestOfficialVerify() {
	message := []byte{0xA4, 0x4C, 0x69, 0x32, 0x00, 0xC3, 0x7B, 0x00, 0x32, 0x68, 0x76, 0x27, 0x17, 0x6E, 0x41, 0xDF, 0xAC, 0xC9, 0x53, 0xCC, 0x77, 0xEB, 0x97, 0x63, 0x81, 0xCD, 0xB7, 0xA6, 0x6B, 0x17, 0x21, 0x58}
	//prv := []byte{0xA8, 0xDE, 0xCB, 0xDF, 0x2A, 0x5C, 0x92, 0xF8, 0xD8, 0xFC, 0x4D, 0x53, 0x36, 0x7F, 0x3A, 0x21, 0x55, 0x84, 0xB0, 0xDD, 0xA9, 0x2E, 0xFC, 0x30, 0xBE, 0x89, 0x51, 0x44, 0xD3, 0xD5, 0x6F, 0x97}
	prikey := []byte{0xA8, 0xDE, 0xCB, 0xDF, 0x2A, 0x5C, 0x92, 0xF8, 0xD8, 0xFC, 0x4D, 0x53, 0x36, 0x7F, 0x3A, 0x21, 0x55, 0x84, 0xB0, 0xDD, 0xA9, 0x2E, 0xFC, 0x30, 0xBE, 0x89, 0x51, 0x44, 0xD3, 0xD5, 0x6F, 0x97}
	pubkey := []byte{0x0B, 0xF0, 0xAE, 0xD1, 0x07, 0x11, 0xCC, 0xE9, 0xC0, 0x7D, 0x6F, 0xFB, 0xB4, 0xCD, 0x9D, 0x93, 0xA0, 0x0B, 0xF5, 0x3A, 0x97, 0x22, 0x08, 0x1E, 0x5A, 0x1A, 0x6C, 0xB5, 0x94, 0xB0, 0xF0, 0x4E, 0xAF, 0x97, 0x8B, 0x8F, 0x7B, 0x7F, 0xCA, 0xFE, 0xEF, 0x85, 0xA3, 0x6F, 0xBA, 0xF6, 0x6C, 0x6F, 0xA0, 0xEA, 0xC0, 0x5D, 0x46, 0x8E, 0x83, 0x41, 0x80, 0xDE, 0x34, 0xCB, 0x74, 0xDD, 0x45, 0xCA}

	sig, ret := owcrypt.ETHsignature(prikey, message)
	if ret != owcrypt.SUCCESS {
		fmt.Println("signature error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		return
	}

	ret = owcrypt.Verify(pubkey, nil, 0, message, 32, sig, owcrypt.ECC_CURVE_SECP256K1|owcrypt.HASH_OUTSIDE_FLAG)
	if ret != owcrypt.SUCCESS {
		fmt.Println("verify error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		return
	}

	if !crypto.VerifySignature(pubkey, message, sig[0:64]) {
		fmt.Println("verify error official")
		return
	}
}

func TestEthereumSigningFunc() {
	//h := []byte{0xA4, 0x4C, 0x69, 0x32, 0x00, 0xC3, 0x7B, 0x00, 0x32, 0x68, 0x76, 0x27, 0x17, 0x6E, 0x41, 0xDF, 0xAC, 0xC9, 0x53, 0xCC, 0x77, 0xEB, 0x97, 0x63, 0x81, 0xCD, 0xB7, 0xA6, 0x6B, 0x17, 0x21, 0x58}
	//prv := []byte{0xA8, 0xDE, 0xCB, 0xDF, 0x2A, 0x5C, 0x92, 0xF8, 0xD8, 0xFC, 0x4D, 0x53, 0x36, 0x7F, 0x3A, 0x21, 0x55, 0x84, 0xB0, 0xDD, 0xA9, 0x2E, 0xFC, 0x30, 0xBE, 0x89, 0x51, 0x44, 0xD3, 0xD5, 0x6F, 0x97}
	//{0xBC, 0xB9, 0x71, 0xDD, 0x9A, 0x73, 0x1B, 0x66, 0xA4, 0x25, 0x51, 0x7F, 0x1F, 0x02, 0xC8, 0xC3, 0xAF, 0x46, 0xAF, 0x74, 0xFF, 0x2F, 0x62, 0xF4, 0xEF, 0x21, 0x14, 0x70, 0x41, 0xC6, 0xBB, 0xA5}
	msg := "0xce99901c86cb500bb1d135b6ab6bf1d918de2a7bc554c9adc25b5a09d25cf303"
	priv := "0xc1a119f163def072f21a1b3b96052c659d71773f2075ec005d4ad84924717e6a"
	sig, err := secp256k1.Sign(common.FromHex(msg), common.FromHex(priv))
	if err != nil {
		fmt.Println("sign error:", err)
		return
	}
	fmt.Printf("sig:")
	for i, b := range sig {
		fmt.Printf("0x%x", b)
		if i != len(sig)-1 {
			fmt.Printf(",")
		}
	}
	fmt.Printf("\n")
}

//key: "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776" password:"123456"
func ExportPrivateKeyFromGeth(address string, password string) string {
	addr := ethcommon.HexToAddress(address)
	manager := &ethereum.WalletManager{}
	ethKeyStore := ethKStore.NewKeyStore(manager.GetConfig().EthereumKeyPath, ethKStore.StandardScryptN, ethKStore.StandardScryptP)
	a := accounts.Account{Address: addr}
	_, key, err := ethKeyStore.GetDecryptedKeyForOpenWallet(a, password)
	if err != nil {
		fmt.Println("get decrypted key failed, err= ", err)
		return ""
	}
	seckey := math.PaddedBigBytes(key.PrivateKey.D, key.PrivateKey.Params().BitSize/8) //key.PrivateKey
	prikey := common.ToHex(seckey)
	log.Debugf("address[%v] private key is:%v", address, prikey)
	return prikey
}

func TestEIP155Signing() {
	//key, _ := crypto.GenerateKey()
	//addr := crypto.PubkeyToAddress(key.PublicKey)
	addr := ethcommon.HexToAddress("0x24de55281e65b8ac88ec0f06e50c325f1b7fb25c")

	signer := types.NewEIP155Signer(big.NewInt(12))
	fmt.Println("addr:", addr.String())

	manager := &ethereum.WalletManager{}
	ethKeyStore := ethKStore.NewKeyStore(manager.GetConfig().EthereumKeyPath, ethKStore.StandardScryptN, ethKStore.StandardScryptP)
	a := accounts.Account{Address: addr}
	a, key, err := ethKeyStore.GetDecryptedKeyForOpenWallet(a, "123456")
	if err != nil {
		fmt.Println("get decrypted key failed, err= ", err)
		return
	}

	amount, err := ethereum.ConvertToBigInt("0xde0b6b3a7640000", 16)
	if err != nil {
		fmt.Println("amount format error.")
		return
	}

	gasPrice, err := ethereum.ConvertToBigInt("0x430e23400", 16)
	if err != nil {
		fmt.Println("gas price format error.")
		return
	}

	tx, err := types.SignTx(types.NewTransaction(0, ethcommon.HexToAddress("0xfcca397e8c94fa147d5c423be55444333c8b9010"),
		amount, 121000, gasPrice, nil), signer, key.PrivateKey)
	if err != nil {
		//t.Fatal(err)
		fmt.Println("sign tx failed, err = ", err)
		return
	}

	toPublicKey := func(pk *ecdsa.PublicKey) []byte {
		testByteX := pk.X.Bytes() //[]byte(*pk.X)
		testByteY := pk.Y.Bytes() //[]byte(*pk.X)
		return append(testByteX, testByteY...)
	}

	fmt.Println("public key:", common.ToHex(toPublicKey(&key.PrivateKey.PublicKey)))

	//fmt.Println("tx:", tx.data)
	tx.PrintTransaction()

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		fmt.Println("EncodeToBytes failed, err = ", err)
		return
	}

	fmt.Println("signature:", common.ToHex(data))
}

func signEIP155FromGethAccount(from string, password string, to string, nonce uint64) (string, error) {
	addr := ethcommon.HexToAddress(from)

	signer := types.NewEIP155Signer(big.NewInt(12))
	fmt.Println("addr:", addr.String())
	manager, _ := GetEthWalletManager()
	ethKeyStore := ethKStore.NewKeyStore(manager.GetConfig().EthereumKeyPath, ethKStore.StandardScryptN, ethKStore.StandardScryptP)
	a := accounts.Account{Address: addr}
	a, key, err := ethKeyStore.GetDecryptedKeyForOpenWallet(a, password)
	if err != nil {
		fmt.Println("get decrypted key failed, err= ", err)
		return "", err
	}

	//100个以太币
	amount, err := ethereum.ConvertToBigInt("0x56bc75e2d63100000", 16)
	if err != nil {
		fmt.Println("amount format error.")
		return "", err
	}

	gasPrice, err := ethereum.ConvertToBigInt("0x430e23400", 16)
	if err != nil {
		fmt.Println("gas price format error.")
		return "", err
	}

	tx, err := types.SignTx(types.NewTransaction(nonce, ethcommon.HexToAddress(to),
		amount, 121000, gasPrice, nil), signer, key.PrivateKey)
	if err != nil {
		//t.Fatal(err)
		fmt.Println("sign tx failed, err = ", err)
		return "", err
	}

	toPublicKey := func(pk *ecdsa.PublicKey) []byte {
		testByteX := pk.X.Bytes() //[]byte(*pk.X)
		testByteY := pk.Y.Bytes() //[]byte(*pk.X)
		return append(testByteX, testByteY...)
	}

	fmt.Println("public key:", common.ToHex(toPublicKey(&key.PrivateKey.PublicKey)))

	//fmt.Println("tx:", tx.data)
	tx.PrintTransaction()

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		fmt.Println("EncodeToBytes failed, err = ", err)
		return "", err
	}

	raw := common.ToHex(data)
	//fmt.Println("signature:",)
	return raw, nil
}

func TestSendRawTransactionFromGethAccount() {
	raw, err := signEIP155FromGethAccount("0x2a63b2203955b84fefe52baca3881b3614991b34", "123456",
		"0xf68f4713AD430b030668aD9DceEEF3591bD9660e", 113)
	if err != nil {
		log.Error("sign failed, err=", err)
		return
	}
	manager, _ := GetEthWalletManager()
	txid, err := manager.EthSendRawTransaction(raw)
	if err != nil {
		log.Error("send raw transaction failed, err=", err)
		return
	}

	log.Debugf("TXID:%v", txid)
}
