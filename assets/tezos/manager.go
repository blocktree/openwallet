package tezos

import (
	"github.com/shopspring/decimal"
	"time"
	"encoding/hex"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/blake2b"
	"log"
	"strconv"
	"github.com/tidwall/gjson"
	"path/filepath"
	"github.com/astaxie/beego/config"
	"errors"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/common"
)

const (
	maxAddresNum = 10000000
)

var (
	//钱包服务API
	serverAPI = "https://rpc.tezrpc.me"
	//钱包主链私钥文件路径
	walletPath = ""
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(1000000)
	//参与汇总的钱包
	//walletsInSum = make(map[string]*Wallet)
	//汇总阀值
	threshold decimal.Decimal = decimal.NewFromFloat(1).Mul(coinDecimal)
	//最小转账额度
	minSendAmount decimal.Decimal = decimal.NewFromFloat(1).Mul(coinDecimal)
	//最小矿工费
	minFees decimal.Decimal = decimal.NewFromFloat(0.0001).Mul(coinDecimal)
	gasLimit decimal.Decimal = decimal.NewFromFloat(0.0001).Mul(coinDecimal)
	storageLimit decimal.Decimal = decimal.NewFromFloat(0).Mul(coinDecimal)
	sumAddress = ""
	//汇总执行间隔时间
	cycleSeconds = time.Second * 10
)


 var prefix = map[string][]byte{
	"tz1": {6, 161, 159},
	"tz2": {6, 161, 161},
	"edpk": {13, 15, 37, 217},
	"edsk": {43, 246, 78, 7},
	"edsig": {9, 245, 205, 134, 18},
}

var watermark = map[string][]byte{
	"block": {1},
	"endorsement": {2},
	"generic": {3},
}

func createAccount() (string, string , string) {
	pub, pri, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Println(err.Error())
	}

	ctx, err:=blake2b.New(20,nil)
	ctx.Write(pub[:])
	pubhash := ctx.Sum(nil)

	pk := base58checkEncode(pub, prefix["edpk"])
	sk := base58checkEncode(pri, prefix["edsk"])
	pkh := base58checkEncode(pubhash, prefix["tz1"])

	return pk, sk, pkh
}

func encryptSecretKey(sk string, password string) string {
	ret, err := Encrypt(password, sk)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}

	return ret
}

func decryptSecretKey(esk string, password string) string {
	ret, err := Decrypt(password, esk)
	if err != nil {
		log.Println(err.Error())
		return err.Error()
	}

	return ret
}

func signTransaction(hash string, sk string, wm []byte) (string, string, error) {
	bhash,_ := hex.DecodeString(hash)
	merbuf := append(wm, bhash...)
	ctx, err :=blake2b.New(32,nil)
	if err != nil {
		return "", "", err
	}
	ctx.Write(merbuf[:])
	bb := ctx.Sum(nil)

	sks, err:= base58checkDecodeNormal(sk, prefix["edsk"])
	if err != nil {
		return "", "", err
	}

	sig := ed25519.Sign(sks[:], bb[:])
	edsig := base58checkEncode(sig, prefix["edsig"])

	sbyte := hash + hex.EncodeToString(sig[:])

	return edsig, sbyte, nil
}

func transfer(keys map[string]string, dst string, fee, gas_limit, storage_limit, amount string) (string, string){
	header := callGetHeader()
	blk_hash := gjson.GetBytes(header, "hash").Str
	chain_id := gjson.GetBytes(header, "chain_id").Str
	protocol := gjson.GetBytes(header, "protocol").Str

	counter :=callGetCounter(keys["pkh"])
	icounter,_ := strconv.Atoi(string(counter))
	icounter = icounter + 1

	manager_key := callGetManagerKey(keys["pkh"])
	//manager := gjson.GetBytes(ret, "manager")
	key := gjson.GetBytes(manager_key, "key")

	var ops []interface{}
	reverl := map[string]string{
		"kind": "reveal",
		"fee": fee,
		"public_key": keys["pk"],
		"source": keys["pkh"],
		"gas_limit": gas_limit,
		"storage_limit": storage_limit,
		"counter": strconv.Itoa(icounter),
	}
	if key.Str == ""{
		icounter = icounter + 1
		ops = append(ops, reverl)
	}

	transaction := map[string]string{
		"kind" : "transaction",
		"amount" : amount,
		"destination" : dst,
		"fee": fee,
		"gas_limit": gas_limit,
		"storage_limit": storage_limit,
		"counter": strconv.Itoa(icounter),
		"source": keys["pkh"],
	}

	ops = append(ops, transaction)
	opOb := make(map[string]interface{})
	opOb["branch"] = blk_hash
	opOb["contents"] = ops
	hash := callForgeOps(chain_id, blk_hash, opOb)

	//sign
	edsig, sbyte, _ := signTransaction(hash, keys["sk"], watermark["generic"])

	//preapply operations
	var opObs []interface{}
	opOb["signature"] = edsig
	opOb["protocol"] = protocol
	opObs = append(opObs, opOb)
	pre := callPreapplyOps(opObs)

	//jnject aperations
	inj := callInjectOps(sbyte)
	return string(inj), string(pre)
}

func CreateNewWallet(name, pw string) error {

	return nil
}

//inputNumber 输入地址数量
func inputNumber() uint64 {

	var (
		count uint64 = 0 // 输入的创建数量
	)

	for {
		// 等待用户输入参数
		line, err := console.Stdin.PromptInput("Enter the number of addresses you want: ")
		if err != nil {
			openwLogger.Log.Errorf("unexpected error: %v", err)
			return 0
		}
		count = common.NewString(line).UInt64()
		if count < 1 {
			log.Printf("Input number must be greater than 0!\n")
			continue
		}
		break
	}

	return count
}

//loadConfig 读取配置
func loadConfig() error {

	var (
		c   config.Configer
		err error
	)

	//读取配置
	absFile := filepath.Join(configFilePath, configFileName)
	c, err = config.NewConfig("json", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd config -s <symbol>' ")
	}

	serverAPI = c.String("apiURL")
	walletPath = c.String("walletPath")
	threshold, _ = decimal.NewFromString(c.String("threshold"))
	threshold = threshold.Mul(coinDecimal)
	minSendAmount, _ = decimal.NewFromString(c.String("minSendAmount"))
	minSendAmount = minSendAmount.Mul(coinDecimal)
	minFees, _ = decimal.NewFromString(c.String("minFees"))
	minFees = minFees.Mul(coinDecimal)
	gasLimit, _ = decimal.NewFromString(c.String("gasLimit"))
	gasLimit = gasLimit.Mul(coinDecimal)
	storageLimit, _ = decimal.NewFromString(c.String("storageLimit"))
	storageLimit = storageLimit.Mul(coinDecimal)
	sumAddress = c.String("sumAddress")

	return nil
}
