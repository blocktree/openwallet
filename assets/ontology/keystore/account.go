package keystore

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/blocktree/OpenWallet/log"
)

type DecryptError struct {
	detail string
}

func (e *DecryptError) Error() string {
	return "encrypt private key error: " + e.detail
}

func NewDecryptError(msg string) *DecryptError {
	return &DecryptError{detail: msg}
}

func DecryptWithCustomScrypt(prot *ProtectedKey, pwd []byte, param *ScryptParam) (*PrivateKey, error) {
	if prot == nil || len(pwd) == 0 {
		return nil, NewDecryptError("invalid argument")
	}

	var plaintext []byte

	// Check parameters
	switch prot.EncAlg {
	case "aes-256-gcm":
		// generate random salt
		salt := prot.Salt
		dkey, err := kdf(pwd, salt, param)
		if err != nil {
			return nil, NewDecryptError(err.Error())
		}
		ekey := dkey[len(dkey)-32:]
		nonce := dkey[:12]
		gcm, err := gcmCipher(ekey)
		plaintext, err = gcm.Open(nil, nonce, prot.Key, []byte(prot.Address))
		if err != nil {
			return nil, NewDecryptError(err.Error())
		}
	default:
		return nil, NewDecryptError("unsupported encryption algorithm")
	}

	switch prot.Alg {
	case "ECDSA", "SM2":
		curve, err := GetNamedCurve(prot.Param["curve"])
		if err != nil {
			return nil, NewDecryptError(err.Error())
		}
		pri := PrivateKey{PrivateKey: ConstructPrivateKey(plaintext, curve)}
		if prot.Alg == "ECDSA" {
			pri.Algorithm = ECDSA
		} else if prot.Alg == "SM2" {
			pri.Algorithm = SM2
		} else {
			return nil, NewDecryptError("unknown ec algorithm")
		}
		return &pri, nil
	default:
		return nil, NewDecryptError("unknown key type")
	}
}

// AccountData - 私钥文件保存的json格式
type AccountData struct {
	ProtectedKey

	Label     string `json:"label"`
	PubKey    string `json:"publicKey"`
	SigSch    string `json:"signatureScheme"`
	IsDefault bool   `json:"isDefault"`
	Lock      bool   `json:"lock"`
}

//SetKeyPair - 设置protected key
func (this *AccountData) SetKeyPair(keyinfo *ProtectedKey) {
	this.Address = keyinfo.Address
	this.EncAlg = keyinfo.EncAlg
	this.Alg = keyinfo.Alg
	this.Hash = keyinfo.Hash
	this.Key = keyinfo.Key
	this.Param = keyinfo.Param
	this.Salt = keyinfo.Salt
}

//GetKeyPair - 获取protected key
func (this *AccountData) GetKeyPair() *ProtectedKey {
	var keyinfo = new(ProtectedKey)
	keyinfo.Address = this.Address
	keyinfo.EncAlg = this.EncAlg
	keyinfo.Alg = this.Alg
	keyinfo.Hash = this.Hash
	keyinfo.Key = this.Key
	keyinfo.Param = this.Param
	keyinfo.Salt = this.Salt
	return keyinfo
}

func (this *AccountData) SetLabel(label string) {
	this.Label = label
}

func (this *WalletData) Load(path string) error {
	msh, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(msh, this)
}

type WalletData struct {
	Name       string         `json:"name"`
	Version    string         `json:"version"`
	Scrypt     *ScryptParam   `json:"scrypt"`
	Identities []Identity     `json:"identities,omitempty"`
	Accounts   []*AccountData `json:"accounts,omitempty"`
	Extra      string         `json:"extra,omitempty"`
}

func (this *WalletData) Save(path string) error {
	data, err := json.Marshal(this)
	if err != nil {
		return err
	}
	if FileExisted(path) {
		filename := path + "~"
		err := ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			return err
		}
		return os.Rename(filename, path)
	} else {
		return ioutil.WriteFile(path, data, 0644)
	}
}

func (this *WalletData) DelAccount(address string) {
	_, index := this.GetAccountByAddress(address)
	if index < 0 {
		return
	}
	this.Accounts = append(this.Accounts[:index], this.Accounts[index+1:]...)
}

func (this *WalletData) GetAccountByAddress(address string) (*AccountData, int) {
	index := -1
	var accData *AccountData
	for i, acc := range this.Accounts {
		if acc.Address == address {
			index = i
			accData = acc
			break
		}
	}
	if index == -1 {
		return nil, -1
	}
	return accData, index
}

func (this *WalletData) GetAccountByIndex(index int) *AccountData {
	if index < 0 || index >= len(this.Accounts) {
		return nil
	}
	return this.Accounts[index]
}

func GetScheme(name string) (SignatureScheme, error) {
	for i, v := range names {
		if strings.ToUpper(v) == strings.ToUpper(name) {
			return SignatureScheme(i), nil
		}
	}

	return 0, errors.New("unknown signature scheme " + name)
}

//ClientImpl keystore实例, 或者叫钱包实例
type ClientImpl struct {
	path       string
	accAddrs   map[string]*AccountData //Map Address(base58) => Account
	accLabels  map[string]*AccountData //Map Label => Account
	defaultAcc *AccountData
	walletData *WalletData
	unlockAccs map[string]*unlockAccountInfo //Map Address(base58) => unlockAccountInfo
	lock       sync.RWMutex
}

func checkSigScheme(keyType, sigScheme string) bool {
	switch strings.ToUpper(keyType) {
	case "ECDSA":
		switch strings.ToUpper(sigScheme) {
		case "SHA224WITHECDSA":
		case "SHA256WITHECDSA":
		case "SHA384WITHECDSA":
		case "SHA512WITHECDSA":
		case "SHA3-224WITHECDSA":
		case "SHA3-256WITHECDSA":
		case "SHA3-384WITHECDSA":
		case "SHA3-512WITHECDSA":
		case "RIPEMD160WITHECDSA":
		default:
			return false
		}
	case "SM2":
		switch strings.ToUpper(sigScheme) {
		case "SM3WITHSM2":
		default:
			return false
		}
	case "ED25519":
		switch strings.ToUpper(sigScheme) {
		case "SHA512WITHEDDSA":
		default:
			return false
		}
	default:
		return false
	}
	return true
}

func (this *ClientImpl) GetDefaultAccount(passwd []byte) (*Account, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if this.defaultAcc == nil {
		return nil, fmt.Errorf("cannot found default account")
	}
	return this.getAccount(this.defaultAcc, passwd)
}

func (this *ClientImpl) getAccount(accData *AccountData, passwd []byte) (*Account, error) {
	privateKey, err := DecryptWithCustomScrypt(&accData.ProtectedKey, passwd, this.walletData.Scrypt)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.Public()
	addr := AddressFromPubKey(publicKey)
	scheme, err := GetScheme(accData.SigSch)
	if err != nil {
		return nil, fmt.Errorf("signature scheme error:%s", err)
	}
	return &Account{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    addr,
		SigScheme:  scheme,
	}, nil
}

func (this *ClientImpl) GetAccountByAddress(address string, passwd []byte) (*Account, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData, ok := this.accAddrs[address]
	if !ok {
		return nil, nil
	}
	return this.getAccount(accData, passwd)
}

func (this *ClientImpl) GetAccountByLabel(label string, passwd []byte) (*Account, error) {
	if len(label) == 0 {
		return nil, nil
	}
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData, ok := this.accLabels[label]
	if !ok {
		return nil, nil
	}
	return this.getAccount(accData, passwd)
}

func (this *ClientImpl) GetAccountByIndex(index int, passwd []byte) (*Account, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData := this.walletData.GetAccountByIndex(index - 1)
	if accData == nil {
		return nil, nil
	}
	return this.getAccount(accData, passwd)
}

func (this *ClientImpl) addAccountData(accData *AccountData) error {
	if !checkSigScheme(accData.Alg, accData.SigSch) {
		return fmt.Errorf("sigScheme:%s does not match KeyType:%s", accData.SigSch, accData.Alg)
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	label := accData.Label
	if label != "" {
		_, ok := this.accLabels[label]
		if ok {
			return fmt.Errorf("duplicate label")
		}
	}
	if len(this.walletData.Accounts) == 0 {
		accData.IsDefault = true
	}
	this.walletData.Accounts = append(this.walletData.Accounts, accData) //this.walletData.AddAccount(accData)
	err := this.walletData.Save(this.path)
	if err != nil {
		this.walletData.DelAccount(accData.Address)
		return fmt.Errorf("save error:%s", err)
	}
	this.accAddrs[accData.Address] = accData
	if accData.IsDefault {
		this.defaultAcc = accData
	}
	if label != "" {
		this.accLabels[label] = accData
	}
	return nil
}

func NewClientImpl(path string) (*ClientImpl, error) {
	cli := &ClientImpl{
		path:       path,
		accAddrs:   make(map[string]*AccountData),
		accLabels:  make(map[string]*AccountData),
		unlockAccs: make(map[string]*unlockAccountInfo),
		walletData: NewWalletData(),
	}
	_, err := os.Stat(path)
	if err != nil { //FileExisted(path) {
		log.Errorf("stat the key file failed, err=%v", err)
		return nil, err
	}

	err = cli.walletData.Load(cli.path)
	if err != nil {
		return nil, fmt.Errorf("load wallet:%s error:%s", cli.path, err)
	}
	for _, accData := range cli.walletData.Accounts {
		cli.accAddrs[accData.Address] = accData
		if accData.Label != "" {
			cli.accLabels[accData.Label] = accData
		}
		if accData.IsDefault {
			cli.defaultAcc = accData
		}
	}

	log.Debugf("default account:%v", cli.defaultAcc.Address)
	return cli, nil
}

func NewWalletData() *WalletData {
	return &WalletData{
		Name:       "MyWallet",
		Version:    "1.1",
		Scrypt:     GetScryptParameters(),
		Identities: nil,
		Extra:      "",
		Accounts:   make([]*AccountData, 0, 0),
	}
}

type SignatureScheme byte

/* crypto object */
type Account struct {
	PrivateKey *PrivateKey
	PublicKey  *PublicKey
	Address    Address
	SigScheme  SignatureScheme
}

type unlockAccountInfo struct {
	acc        *Account
	unlockTime time.Time
	expiredAt  int //s
}

const (
	SHA224withECDSA SignatureScheme = iota
	SHA256withECDSA
	SHA384withECDSA
	SHA512withECDSA
	SHA3_224withECDSA
	SHA3_256withECDSA
	SHA3_384withECDSA
	SHA3_512withECDSA
	RIPEMD160withECDSA

	SM3withSM2

	SHA512withEDDSA
)

var names []string = []string{
	"SHA224withECDSA",
	"SHA256withECDSA",
	"SHA384withECDSA",
	"SHA512withECDSA",
	"SHA3-224withECDSA",
	"SHA3-256withECDSA",
	"SHA3-384withECDSA",
	"SHA3-512withECDSA",
	"RIPEMD160withECDSA",
	"SM3withSM2",
	"SHA512withEdDSA",
}

func (s SignatureScheme) Name() string {
	if int(s) >= len(names) {
		panic(fmt.Sprintf("unknown scheme value %v", s))
	}
	return names[s]
}

func (this *ClientImpl) NewAccount(label string, typeCode KeyType, curveCode byte, sigScheme SignatureScheme, passwd []byte) (*Account, error) {
	if len(passwd) == 0 {
		return nil, fmt.Errorf("password cannot empty")
	}
	prvkey, pubkey, err := GenerateKeyPair(typeCode, curveCode)
	if err != nil {
		return nil, fmt.Errorf("generateKeyPair error:%s", err)
	}
	address := AddressFromPubKey(pubkey)
	addressBase58 := address.ToBase58()
	prvSecret, err := EncryptPrivateKey(prvkey, addressBase58, passwd)
	if err != nil {
		return nil, fmt.Errorf("encryptPrivateKey error:%s", err)
	}
	accData := &AccountData{}
	accData.Label = label
	accData.SetKeyPair(prvSecret)
	accData.SigSch = sigScheme.Name()
	accData.PubKey = hex.EncodeToString(SerializePublicKey(pubkey))

	err = this.addAccountData(accData)
	if err != nil {
		return nil, err
	}
	return &Account{
		PrivateKey: prvkey,
		PublicKey:  pubkey,
		Address:    address,
		SigScheme:  sigScheme,
	}, nil
}

func (this *ClientImpl) NewAccountWithHK(label string, typeCode KeyType, curveCode byte, sigScheme SignatureScheme, passwd []byte) (*Account, error) {
	if len(passwd) == 0 {
		return nil, fmt.Errorf("password cannot empty")
	}
	prvkey, pubkey, err := GenerateKeyPairWithHdKey(1) //GenerateKeyPair(typeCode, curveCode)
	if err != nil {
		return nil, fmt.Errorf("generateKeyPair error:%s", err)
	}
	address := AddressFromPubKey(pubkey)
	addressBase58 := address.ToBase58()
	//log.Debugf("address:%v", string(address[:]))
	//log.Debugf("addressBase58:%v", string(addressBase58))
	prvSecret, err := EncryptPrivateKey(prvkey, addressBase58, passwd)
	if err != nil {
		return nil, fmt.Errorf("encryptPrivateKey error:%s", err)
	}
	accData := &AccountData{}
	accData.Label = label
	accData.SetKeyPair(prvSecret)
	accData.SigSch = sigScheme.Name()
	accData.PubKey = hex.EncodeToString(SerializePublicKey(pubkey))

	//log.Debugf("accData:%v", common.FormatStruct(accData))
	err = this.addAccountData(accData)
	if err != nil {
		return nil, err
	}
	return &Account{
		PrivateKey: prvkey,
		PublicKey:  pubkey,
		Address:    address,
		SigScheme:  sigScheme,
	}, nil
}
