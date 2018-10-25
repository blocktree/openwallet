package tech

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"

	"github.com/blocktree/OpenWallet/assets/ethereum"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/manager"
	"github.com/blocktree/OpenWallet/openwallet"
	owcrypt "github.com/blocktree/go-OWCrypt"
	"github.com/bytom/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	FRAME_DEFAULT_DIR = "frame_data/eth"
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
	w := &openwallet.Wallet{Alias: "framepeter", IsTrust: true, Password: "12345678"}
	nw, key, err := tm.CreateWallet(testApp, w)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("wallet:", nw)
	log.Info("key:", key)
}

func TestWalletManager_CreateAssetsAccount() {

	walletID := "W9cRnfgyZ7T4imjbQuiafz6Ca5aUf8qJRJ" //"WFPHAs2uyeHcfBzKF4vN4NkMpArX8wkCxp"

	//account := &openwallet.AssetsAccount{Alias: "Tim", WalletID: walletID, Required: 1, Symbol: "BTC", IsTrust: true}
	//account, err := tm.CreateAssetsAccount(testApp, walletID, "12345678", account, nil)
	account := &openwallet.AssetsAccount{Alias: "Alice", WalletID: walletID, Required: 1, Symbol: "ETH", IsTrust: true}
	account, _, err := tm.CreateAssetsAccount(testApp, walletID, "12345678", account, nil)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("account:", account)

	tm.CloseDB(testApp)
}

func TestFrameWalletManager_CreateAddress() {
	walletID := "W9cRnfgyZ7T4imjbQuiafz6Ca5aUf8qJRJ" //"WFPHAs2uyeHcfBzKF4vN4NkMpArX8wkCxp"
	//accountID := "KhJdnr4UJLdbeQcMvZgedyYykRVTdLaMLbsV2mx3GZiMva9Kfb"
	accountID := "4mNzv15wSPeUgqfw2Y4UieRJzUoJJMS9DM1L136gxFMZ" //"L77QWWMRhsKMiArgaMTiVaxa6knz2Wo2eNk5F3Bw764XeDyq3T"
	address, err := tm.CreateAddress(testApp, walletID, accountID, 1)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("address:", address)

	tm.CloseDB(testApp)
}

func GetAddress(address string) *openwallet.Address {
	db, err := tm.OpenDB(testApp)
	if err != nil {
		log.Debugf("open db failed, err=%v", err)
		return nil
	}
	var obj openwallet.Address
	err = db.One("Address", address, &obj)
	if err != nil {
		log.Debugf("get address failed, err=%v", err)
	}
	return &obj
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

	fromAddr := GetAddress(address)
	if fromAddr == nil {
		log.Debugf("wallet[%v] get address failed", w.WalletID)
		return ""
	}

	log.Debugf("fromAddr:%v", fromAddr.HDPath)
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
	//5813387dE3fAF2012a8D63580A23090Eca337f61
	prikeyselfmade := GetPrivateKeyInWallet("WJJ59GafizEBC8GZ8Pv6PopohBUK4GUrd6", "12345678", "2d3a164eD8019d3111b0726399a6a9B10F05a8e6")
	prikeyeth := ExportPrivateKeyFromGeth("0x50068fd632c1a6e6c5bd407b4ccf8861a589e776", "123456")

	log.Debugf("self made private key:%v", prikeyselfmade)
	log.Debugf("eth private key exported:%v", prikeyeth)
}

//0x50068fd632c1a6e6c5bd407b4ccf8861a589e776的私钥:0x878dff18f5709bbba276d74cad3bc918f74e745b5599a87f087d0073acf09250
func TestSendRawTransaction() {
	//walletId:WJJ59GafizEBC8GZ8Pv6PopohBUK4GUrd6
	//pssword:12345678
	//from:2d3a164eD8019d3111b0726399a6a9B10F05a8e6
	//to:5813387dE3fAF2012a8D63580A23090Eca337f61
	manager, _ := GetEthWalletManager()
	raw, err := signOWEIP155("W9wpMYxZNB1tRc64dHFuGswBh6NavpJMg8", "12345678", "31ab40e917f646581be5f0d3112d2b06366ce8ce", "0x584a9ed7f95cd04337df791fac32bed88e13b77a", 0)
	if err != nil {
		log.Error("signOWEIP155 failed, err=", err)
		return
	}
	txid, err := manager.EthSendRawTransaction(raw)
	if err != nil {
		log.Debugf("EthSendRawTransaction failed, err = %v", err)
		return
	}
	log.Info("TXID:", txid)
}

func signOWEIP155(walletID string, password string, from string, to string, nonce uint64) (string, error) {
	prikeyselfmade := GetPrivateKeyInWallet(walletID, password, from)
	if prikeyselfmade == "" {
		log.Debugf("GetPrivateKeyInWallet failed")
		return "", errors.New("GetPrivateKeyInWallet failed")
	}

	//100- 0x56bc75e2d63100000
	amount, err := ethereum.ConvertToBigInt("0x1043561a8829300000", 16)
	if err != nil {
		fmt.Println("amount format error.")
		return "", err
	}

	gasPrice, err := ethereum.ConvertToBigInt("0x430e23400", 16)
	if err != nil {
		fmt.Println("gas price format error.")
		return "", err
	}

	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(to),
		amount, 21000, gasPrice, nil)

	signer := types.NewEIP155Signer(big.NewInt(12))
	message := signer.Hash(tx)
	//seckey := math.PaddedBigBytes(key.PrivateKey.D, key.PrivateKey.Params().BitSize/8)

	/*sig, ret := ethereum.ETHsignature(common.FromHex(prikeyselfmade), message[:])
	if ret != owcrypt.SUCCESS {
		fmt.Println("signature error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		return "", err
	}*/
	txstr, _ := json.MarshalIndent(tx, "", " ")
	log.Debug("--txStr:", string(txstr))
	log.Debug("--message:", common.ToHex(message[:]))
	log.Debug("--prikeyselfmade:", prikeyselfmade)
	sig, err := secp256k1.Sign(message[:], common.FromHex(prikeyselfmade))
	if err != nil {
		fmt.Println("signature error, err=", err)
		return "", err
	}

	log.Debug("--sig:", common.ToHex(sig))
	tx, err = tx.WithSignature(signer, sig)
	if err != nil {
		fmt.Println("with signature failed, err=", err)
		return "", err
	}
	txstr, _ = json.MarshalIndent(tx, "", " ")
	log.Debug("--after signed txStr:", string(txstr))
	tx.PrintTransaction()

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		fmt.Println("EncodeToBytes failed, err = ", err)
		return "", err
	}

	fmt.Println("--signature:", common.ToHex(data))
	return common.ToHex(data), nil
}

/*func signEIP155(walletID string, password string, from string, to string, nonce uint64) (string, error) {
	addr := ethcommon.HexToAddress(from)

	signer := types.NewEIP155Signer(big.NewInt(12))
	fmt.Println("addr:", addr.String())

	ethKeyStore := ethKStore.NewKeyStore(ethereum.EthereumKeyPath, ethKStore.StandardScryptN, ethKStore.StandardScryptP)
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
}*/

func TestTokenBalance() {
	manager, _ := GetEthWalletManager()

	addrs := []ethereum.AddrBalanceInf{
		&ethereum.AddrBalance{Address: "0x50068fD632c1A6e6c5bD407b4cCf8861A589E776", Index: 0},
		&ethereum.AddrBalance{Address: "0x2A63B2203955b84FefE52BAca3881b3614991b34", Index: 1},
		&ethereum.AddrBalance{Address: "0x584a9Ed7f95Cd04337df791Fac32bED88E13b77a", Index: 2},
		&ethereum.AddrBalance{Address: "0xdb9a569f7b80030956dc9686b89D5fF15922E175", Index: 3},
	}

	err := manager.GetTokenBalanceByAddress("0x8847E5F841458ace82dbb0692C97115799fe28d3", addrs...)
	if err != nil {
		log.Errorf("get token balance by address failed, err=%v", err)
		return
	}
	objStr, _ := json.MarshalIndent(addrs, "", " ")
	log.Debugf("balance list:%v", string(objStr))
}

func TestGetBalanceByAddress() {
	manager, _ := GetEthWalletManager()
	addrs := []string{
		"0x50068fD632c1A6e6c5bD407b4cCf8861A589E776",
		"0x2A63B2203955b84FefE52BAca3881b3614991b34",
		"0x584a9Ed7f95Cd04337df791Fac32bED88E13b77a",
		"0xdb9a569f7b80030956dc9686b89D5fF15922E175",
	}

	balanceList, err := manager.GetBalanceByAddress(addrs...)
	if err != nil {
		log.Errorf("get token balance by address failed, err=%v", err)
		return
	}

	objStr, _ := json.MarshalIndent(balanceList, "", " ")
	log.Debugf("balance list:%v", string(objStr))
}

func TestTokenDecode() {
	manager, _ := GetEthWalletManager()

	contract := openwallet.SmartContract{
		Address: "0x8847E5F841458ace82dbb0692C97115799fe28d3",
	}
	addrs := []string{
		"0x50068fD632c1A6e6c5bD407b4cCf8861A589E776",
		"0x2A63B2203955b84FefE52BAca3881b3614991b34",
		"0x584a9Ed7f95Cd04337df791Fac32bED88E13b77a",
		"0xdb9a569f7b80030956dc9686b89D5fF15922E175",
	}
	balanceList, err := manager.ContractDecoder.GetTokenBalanceByAddress(contract, addrs...)
	if err != nil {
		log.Errorf("get token balance by address failed, err=%v", err)
		return
	}

	objStr, _ := json.MarshalIndent(balanceList, "", " ")
	log.Debugf("balance list:%v", string(objStr))
}
