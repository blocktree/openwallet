/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package keystore

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/btcsuite/btcutil/hdkeychain"
	"os"
	"path/filepath"
	"io/ioutil"
	"fmt"
	"encoding/hex"
	"time"
	"github.com/ethereum/go-ethereum/crypto/randentropy"
	"golang.org/x/crypto/scrypt"
	"github.com/ethereum/go-ethereum/common/math"
	"crypto/aes"
	"github.com/ethereum/go-ethereum/crypto"
	"encoding/json"
	"bytes"
	"golang.org/x/crypto/pbkdf2"
	"crypto/sha256"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha512"
	"math/big"
	"github.com/btcsuite/btcd/btcec"
	"strings"
	t "github.com/blocktree/OpenWallet/common"
	"errors"
	"crypto/ecdsa"
	"github.com/blocktree/OpenWallet/openwallet"
)

const (

	// HDKey的规范版本号
	version = 1

	//HDKeystore的账户路径字段，hardened账户类型开始位置
	HDKeystoreHardenedKeyStart = 0x80

	// maxCoinType is the maximum allowed coin type used when structuring
	// the BIP0044 multi-account hierarchy.  This value is based on the
	// limitation of the underlying hierarchical deterministic key
	// derivation.
	maxCoinType = hdkeychain.HardenedKeyStart - 1

	// The hierarchy described by BIP0043 is:
	//  m/<purpose>'/*
	// This is further extended by BIP0044 to:
	//  m/44'/<coin type>'/<account>'
	// BIP0044，m/44'/
	//openwallet coin type is 88': m/44'/88'
	openwCoinTypePath = "m/44'/88'"

)

var (
	// masterKey is the master key used along with a random seed used to generate
	// the master node in the hierarchical tree.
	masterKey = []byte("openwallet seed")

	// BIP32 hierarchical deterministic extended key magics
	hdPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4} // starts with xprv

	//Derived路径错误
	ErrInvalidDerivedPath = errors.New("Invalid DerivedPath")

	//错误的HDPath
	ErrInvalidHDPath = errors.New("Invalid HDPath")
)

// HDKey 分层确定性密钥，基于BIP32模型创建的账户模型
type HDKey struct {
	// 所持密钥的用户key
	UserKey string
	// 用于校验私钥的地址
	Address openwallet.Address
	//账户路径
	HDPath string
	//账户数量
	AccountNum uint
	// 根私钥
	MasterKey *hdkeychain.ExtendedKey

}

type encryptedHDKeyJSON struct {
	UserKey string 		`json:userkey`
	Address string     `json:"address"`
	Crypto  cryptoJSON `json:"crypto"`
	HDPath 	string 		`json:"hdpath"`
	Version int     `json:"version"`
}


type cryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

type miningJSON struct {
	Cycle       uint              `json:"cycle"`
	Algorithm   string            `json:"algorithm"`
	Delay 		uint       		  `json:"delay"`
}

// getDerivedKeyWithPath
// "" (root key)
// "m" (root key)
// "/" (root key)
// "m/0'" (hardened child #0 of the root key)
// "/0'" (hardened child #0 of the root key)
// "0'" (hardened child #0 of the root key)
// "m/44'/1'/2'" (BIP44 testnet account #2)
// "/44'/1'/2'" (BIP44 testnet account #2)
// "44'/1'/2'" (BIP44 testnet account #2)
//
// The following paths are invalid:
//
// "m / 0 / 1" (contains spaces)
// "m/b/c" (alphabetical characters instead of numerical indexes)
// "m/1.2^3" (contains illegal characters)
func getDerivedKeyWithPath(key *hdkeychain.ExtendedKey, path string) (*hdkeychain.ExtendedKey, error) {

	var (
		err error
	)

	if len(path) == 0 {
		return nil, ErrInvalidDerivedPath
	}

	if path == "m" || path == "/" || path == "" {
		// 直接返回当前根
		return key, nil
	}

	// strip "m/" from the beginning.
	if strings.Index(path, "m/") == 0 {
		path = path[2:]
	}

	derivedKey := key

	// m/<purpose>'/<coin type>' 分解路径
	elements := strings.Split(path, "/")

	for i, elem := range elements {
		if len(elem) == 0 {
			continue
		}
		var value t.String
		hardened := false
		if strings.Index(elem, "'") == len(elem) - 1 {
			hardened = true
			elem = elem[0:len(elem) - 1]
		}

		value = t.NewString(elem)
		if i >= 0 && value.String() == elem {
			if hardened {
				derivedKey.Child(hdkeychain.HardenedKeyStart + value.UInt32())
			} else {
				derivedKey.Child(value.UInt32())
			}
		} else {
			return nil, ErrInvalidDerivedPath
		}
	}

	return derivedKey, err
}

// DerivedKeyWithPath 根据BIP32的规则获取子密钥，例如：m/<purpose>'/*
func (k *HDKey)DerivedKeyWithPath(path string) (*hdkeychain.ExtendedKey, error) {
	return getDerivedKeyWithPath(k.MasterKey, path)
}

// RootKey 获取用于管理账户的根私钥
func (k *HDKey)RootKey() *hdkeychain.ExtendedKey {

	var (
		rootkey = k.MasterKey
	)

	pathBytes := common.Hex2Bytes(k.HDPath)
	if len(pathBytes) == 0 {
		return rootkey
	}

	return getRootKeyWithHDPath(k.MasterKey, pathBytes)
}

// getRootKeyWithHDPath 通过HDPath获取账户管理的根私钥
func getRootKeyWithHDPath(key *hdkeychain.ExtendedKey, hdPath []byte) *hdkeychain.ExtendedKey {

	var (
		rootkey = key
	)

	index := int(hdPath[0])

	if len(hdPath) - 1 < index {
		return rootkey
	}

	for i := 0; i < index; i++ {
		value := uint32(hdPath[i+1])	//首字节记录开始位置，所以i+1
		if value >= HDKeystoreHardenedKeyStart {
			rootkey, _ = rootkey.Child(hdkeychain.HardenedKeyStart + value)
		} else {
			rootkey, _ = rootkey.Child(value)
		}
	}

	return rootkey
}

// EncryptKey encrypts a key using the specified scrypt parameters into a json
// blob that can be decrypted later on.
func EncryptKey(hdkey *HDKey, auth string, scryptN, scryptP int) ([]byte, error) {
	authArray := []byte(auth)
	salt := randentropy.GetEntropyCSPRNG(32)
	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptR, scryptP, scryptDKLen)
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:16]
	privatekey, err := hdkey.MasterKey.ECPrivKey()
	if err != nil {
		return nil, err
	}
	keyBytes := math.PaddedBigBytes(privatekey.D, 32)

	iv := randentropy.GetEntropyCSPRNG(aes.BlockSize) // 16
	cipherText, err := aesCTRXOR(encryptKey, keyBytes, iv)
	if err != nil {
		return nil, err
	}
	mac := crypto.Keccak256(derivedKey[16:32], cipherText)

	scryptParamsJSON := make(map[string]interface{}, 5)
	scryptParamsJSON["n"] = scryptN
	scryptParamsJSON["r"] = scryptR
	scryptParamsJSON["p"] = scryptP
	scryptParamsJSON["dklen"] = scryptDKLen
	scryptParamsJSON["salt"] = hex.EncodeToString(salt)

	cipherParamsJSON := cipherparamsJSON{
		IV: hex.EncodeToString(iv),
	}

	cryptoStruct := cryptoJSON{
		Cipher:       "aes-128-ctr",
		CipherText:   hex.EncodeToString(cipherText),
		CipherParams: cipherParamsJSON,
		KDF:          keyHeaderKDF,
		KDFParams:    scryptParamsJSON,
		MAC:          hex.EncodeToString(mac),
	}

	//生成Rootkey的底子
	rootPub, err := hdkey.RootKey().Neuter()
	if err != nil {
		return nil, err
	}

	encryptedHDKeyJSON := encryptedHDKeyJSON{
		UserKey:hdkey.UserKey,
		Address: rootPub.String(),
		Crypto:cryptoStruct,
		HDPath: hdkey.HDPath,
		Version: version,

	}
	return json.Marshal(encryptedHDKeyJSON)
}

// DecryptKey decrypts a key from a json blob, returning the private key itself.
func DecryptHDKey(keyjson []byte, auth string) (*HDKey, error) {
	// Parse the json into a simple map to fetch the key version
	m := make(map[string]interface{})
	if err := json.Unmarshal(keyjson, &m); err != nil {
		return nil, err
	}
	// Depending on the version try to parse one way or another
	var (
		hdpath			[]byte
		keyBytes 		[]byte
		userKey 		string
		err             error
	)
	k := new(encryptedHDKeyJSON)
	if err := json.Unmarshal(keyjson, k); err != nil {
		return nil, err
	}
	keyBytes, userKey, hdpath, err = decryptHDKey(k, auth)
	// Handle any decryption errors and return the key
	if err != nil {
		return nil, err
	}

	master, err := newKeyFromBIP32(keyBytes)
	if err != nil {
		return nil, err
	}

	rootkey := getRootKeyWithHDPath(master, hdpath)

	return &HDKey{
		UserKey: userKey,
		Address: openwallet.ExtendedKeyToAddress(rootkey),
		HDPath: common.Bytes2Hex(hdpath),
		AccountNum: 0,
		MasterKey: master,
	}, nil
}

// decryptHDKey 解密HDKey的文件内容
func decryptHDKey(keyProtected *encryptedHDKeyJSON, auth string) (keyBytes []byte, userKey string, hdPath []byte, err error) {

	if keyProtected.Crypto.Cipher != "aes-128-ctr" {
		return nil, "", nil, fmt.Errorf("Cipher not supported: %v", keyProtected.Crypto.Cipher)
	}

	userKey = keyProtected.UserKey
	mac, err := hex.DecodeString(keyProtected.Crypto.MAC)
	if err != nil {
		return nil, "", nil, err
	}

	iv, err := hex.DecodeString(keyProtected.Crypto.CipherParams.IV)
	if err != nil {
		return nil, "", nil, err
	}

	cipherText, err := hex.DecodeString(keyProtected.Crypto.CipherText)
	if err != nil {
		return nil, "", nil, err
	}

	derivedKey, err := getKDFKey(keyProtected.Crypto, auth)
	if err != nil {
		return nil, "", nil, err
	}

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, "", nil, ErrDecrypt
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, "", nil, err
	}
	return plainText, userKey, nil, err
}

// getKDFKey
func getKDFKey(cryptoJSON cryptoJSON, auth string) ([]byte, error) {
	authArray := []byte(auth)
	salt, err := hex.DecodeString(cryptoJSON.KDFParams["salt"].(string))
	if err != nil {
		return nil, err
	}
	dkLen := ensureInt(cryptoJSON.KDFParams["dklen"])

	if cryptoJSON.KDF == keyHeaderKDF {
		n := ensureInt(cryptoJSON.KDFParams["n"])
		r := ensureInt(cryptoJSON.KDFParams["r"])
		p := ensureInt(cryptoJSON.KDFParams["p"])
		return scrypt.Key(authArray, salt, n, r, p, dkLen)

	} else if cryptoJSON.KDF == "pbkdf2" {
		c := ensureInt(cryptoJSON.KDFParams["c"])
		prf := cryptoJSON.KDFParams["prf"].(string)
		if prf != "hmac-sha256" {
			return nil, fmt.Errorf("Unsupported PBKDF2 PRF: %s", prf)
		}
		key := pbkdf2.Key(authArray, salt, c, dkLen, sha256.New)
		return key, nil
	}

	return nil, fmt.Errorf("Unsupported KDF: %s", cryptoJSON.KDF)
}

// TODO: can we do without this when unmarshalling dynamic JSON?
// why do integers in KDF params end up as float64 and not int after
// unmarshal?
func ensureInt(x interface{}) int {
	res, ok := x.(int)
	if !ok {
		res = int(x.(float64))
	}
	return res
}

// newKeyFromBIP32 创建根私钥
func newKeyFromBIP32(seed []byte) (*hdkeychain.ExtendedKey, error) {
	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
	if len(seed) < hdkeychain.MinSeedBytes || len(seed) > hdkeychain.MaxSeedBytes {
		return nil, hdkeychain.ErrInvalidSeedLen
	}

	// First take the HMAC-SHA512 of the master key and the seed data:
	//   I = HMAC-SHA512(Key = "Bitcoin seed", Data = S)
	hmac512 := hmac.New(sha512.New, masterKey)
	hmac512.Write(seed)
	lr := hmac512.Sum(nil)

	// Split "I" into two 32-byte sequences Il and Ir where:
	//   Il = master secret key
	//   Ir = master chain code
	secretKey := lr[:len(lr)/2]
	chainCode := lr[len(lr)/2:]

	// Ensure the key in usable.
	secretKeyNum := new(big.Int).SetBytes(secretKey)
	if secretKeyNum.Cmp(btcec.S256().N) >= 0 || secretKeyNum.Sign() == 0 {
		return nil, hdkeychain.ErrUnusableSeed
	}

	parentFP := []byte{0x00, 0x00, 0x00, 0x00}
	return hdkeychain.NewExtendedKey(hdPrivateKeyID[:], secretKey, chainCode,
		parentFP, 0, 0, true), nil
}

// NewHDKey 通过userkey，私钥种子，账户路径，创建HDKey
func NewHDKey(userKey string, seed []byte, startPath string) (*HDKey, error) {

	var (
		err error
	)

	//创建根私钥
	master, err := newKeyFromBIP32(seed)
	if err != nil {
		return nil, err
	}

	//把startPath编码，m/44'/<coin type>'
	hdPath, err := encodeStartPath(startPath)
	if err != nil {
		return nil, err
	}

	//获取账户的根私钥
	key, err := getDerivedKeyWithPath(master, startPath)
	if err != nil {
		return nil, err
	}

	pubKey, err := key.ECPubKey()
	if err != nil {
		return nil, err
	}

	//地址不正确
	hdkey := &HDKey{
		UserKey: userKey,
		Address: openwallet.PubkeyToAddress(ecdsa.PublicKey(*pubKey)),
		HDPath: common.Bytes2Hex(hdPath),
		AccountNum: 0,
		MasterKey: master,
	}

	return hdkey, nil
}


// encodeStartPath 编码账户开始位置路径
func encodeStartPath(path string) ([]byte, error) {

	var (
		err error
		hdPath []byte = make([]byte,0)
	)

	// m/44'/<coin type>' 分解路径
	elements := strings.Split(path, "/")

	if len(elements) == 0 {
		hdPath = append(hdPath, uint8(1))
		hdPath = append(hdPath, 0)
		return hdPath, err
	}

	hdPath = append(hdPath, uint8(len(elements)))

	for _, elem := range elements {
		if len(elem) == 0 {
			continue
		}
		value := uint8(0)
		hardened := false
		if strings.Index(elem, "'") == len(elem) - 1 {
			hardened = true
			elem = elem[0:len(elem) - 1]
		}

		if elem == "m" {
			value = 0
		}

		value = t.NewString(elem).UInt8()
		if hardened {
			value = HDKeystoreHardenedKeyStart + value
		}

		hdPath = append(hdPath, value)
	}

	return hdPath, err
}

//func storeNewKey(ks keyStore, rand io.Reader, auth string) (*Key, accounts.Account, error) {
//	key, err := newKey(rand)
//	if err != nil {
//		return nil, accounts.Account{}, err
//	}
//	a := accounts.Account{Address: key.Address, URL: accounts.URL{Scheme: KeyStoreScheme, Path: ks.JoinPath(keyFileName(key.Address))}}
//	if err := ks.StoreKey(a.URL.Path, key, auth); err != nil {
//		zeroKey(key.PrivateKey)
//		return nil, a, err
//	}
//	return key, a, err
//}

func writeKeyFile(file string, content []byte) error {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(), file)
}

// keyFileName implements the naming convention for keyfiles:
// UTC--<created_at UTC ISO8601>-<address hex>
func keyFileName(keyAddr common.Address) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr[:]))
}

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}


func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

func aesCBCDecrypt(key, cipherText, iv []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCBCDecrypter(aesBlock, iv)
	paddedPlaintext := make([]byte, len(cipherText))
	decrypter.CryptBlocks(paddedPlaintext, cipherText)
	plaintext := pkcs7Unpad(paddedPlaintext)
	if plaintext == nil {
		return nil, ErrDecrypt
	}
	return plaintext, err
}

// From https://leanpub.com/gocrypto/read#leanpub-auto-block-cipher-modes
func pkcs7Unpad(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}

	padding := in[len(in)-1]
	if int(padding) > len(in) || padding > aes.BlockSize {
		return nil
	} else if padding == 0 {
		return nil
	}

	for i := len(in) - 1; i > len(in)-int(padding)-1; i-- {
		if in[i] != padding {
			return nil
		}
	}
	return in[:len(in)-int(padding)]
}