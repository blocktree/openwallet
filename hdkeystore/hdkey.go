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

package hdkeystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/crypto/sha3"
	"github.com/blocktree/go-owcdrivers/owkeychain"
	"github.com/blocktree/go-owcrypt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

const (

	// HDKey的规范版本号
	version = 1

	// maxCoinType is the maximum allowed coin type used when structuring
	// the BIP0044 multi-account hierarchy.  This value is based on the
	// limitation of the underlying hierarchical deterministic key
	// derivation.
	maxCoinType = owkeychain.HardenedKeyStart - 1

	// MinSeedBytes is the minimum number of bytes allowed for a seed to
	// a master node.
	MinSeedBytes = 16 // 128 bits

	// MaxSeedBytes is the maximum number of bytes allowed for a seed to
	// a master node.
	MaxSeedBytes = 64 // 512 bits

	// The hierarchy described by BIP0043 is:
	//  m/<purpose>'/*
	// This is further extended by BIP0044 to:
	//  m/44'/<coin type>'/<account>'
	// BIP0044，m/44'/
	//openwallet coin type is 88': m/44'/88'
	OpenwCoinTypePath = "m/44'/88'"
)

var (

	// masterKey is the master key used along with a random seed used to generate
	// the master node in the hierarchical tree.
	masterKey = []byte("openwallet seed")

	//KeyID首字节的标识
	KeyIDVer = []byte{0x48}

	//Derived路径错误
	ErrInvalidDerivedPath = errors.New("Invalid DerivedPath")

	//错误的HDPath
	ErrInvalidHDPath = errors.New("Invalid HDPath")

	// ErrInvalidSeedLen describes an error in which the provided seed or
	// seed length is not in the allowed range.
	ErrInvalidSeedLen = fmt.Errorf("seed length must be between %d and %d "+
		"bits", MinSeedBytes*8, MaxSeedBytes*8)
)

// HDKey 分层确定性密钥，基于BIP32模型创建的账户模型
type HDKey struct {
	//私钥别名
	Alias string
	//账户路径
	RootPath string
	// 账户的扩展ID
	KeyID string
	//种子，加密保存
	seed []byte
}

// 加密后的HDKey的JSON结构
type encryptedHDKeyJSON struct {
	Alias    string     `json:"alias"`
	KeyID    string     `json:"keyid"`
	Crypto   cryptoJSON `json:"crypto"`
	RootPath string     `json:"rootpath"`
	Version  int        `json:"version"`
}

// 加密内容的JSON结构
type cryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

// 加密初始向量IV
type cipherparamsJSON struct {
	IV string `json:"iv"`
}

type miningJSON struct {
	Cycle     uint   `json:"cycle"`
	Algorithm string `json:"algorithm"`
	Delay     uint   `json:"delay"`
}

// DerivedKeyWithPath 根据BIP32的规则获取子密钥，例如：m/<purpose>'/*
// @param path string
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
// @param curveType string
// ECC_CURVE_SECP256K1
// ECC_CURVE_SECP256R1
// ECC_CURVE_ED25519
func (k *HDKey) DerivedKeyWithPath(path string, curveType uint32) (*owkeychain.ExtendedKey, error) {
	return owkeychain.DerivedPrivateKeyWithPath(k.seed, path, curveType)
}

//func (k *HDKey) DerivedKeyWithPath2(path string, curveType  uint32) (*hdkeychain.ExtendedKey, error) {
//	return getDerivedKeyWithPath(k.seed, path)
//}
//
//// newKeyFromBIP32 创建根私钥
//func newKeyFromBIP32(seed []byte) (*hdkeychain.ExtendedKey, error) {
//	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
//	if len(seed) < hdkeychain.MinSeedBytes || len(seed) > hdkeychain.MaxSeedBytes {
//		return nil, hdkeychain.ErrInvalidSeedLen
//	}
//
//	// First take the HMAC-SHA512 of the master key and the seed data:
//	//   I = HMAC-SHA512(Key = "Bitcoin seed", Data = S)
//	hmac512 := hmac.New(sha512.New, masterKey)
//	hmac512.Write(seed)
//	lr := hmac512.Sum(nil)
//
//	// Split "I" into two 32-byte sequences Il and Ir where:
//	//   Il = master secret key
//	//   Ir = master chain code
//	secretKey := lr[:len(lr)/2]
//	chainCode := lr[len(lr)/2:]
//
//	// Ensure the key in usable.
//	secretKeyNum := new(big.Int).SetBytes(secretKey)
//	if secretKeyNum.Cmp(btcec.S256().N) >= 0 || secretKeyNum.Sign() == 0 {
//		return nil, hdkeychain.ErrUnusableSeed
//	}
//
//	parentFP := []byte{0x00, 0x00, 0x00, 0x00}
//	hdPrivateKeyID := [4]byte{0x04, 0x88, 0xad, 0xe4}
//	return hdkeychain.NewExtendedKey(hdPrivateKeyID[:], secretKey, chainCode,
//		parentFP, 0, 0, true), nil
//}
//
//func getDerivedKeyWithPath(seed []byte, path string) (*hdkeychain.ExtendedKey, error) {
//
//	var (
//		err error
//	)
//
//	//if len(path) == 0 {
//	//	return nil, ErrInvalidDerivedPath
//	//}
//
//	key, err := newKeyFromBIP32(seed)
//	if err != nil {
//		return nil, err
//	}
//
//	if path == "m" || path == "/" || path == "" {
//		// 直接返回当前根
//		return key, nil
//	}
//
//	// strip "m/" from the beginning.
//	if strings.Index(path, "m/") == 0 {
//		path = path[2:]
//	}
//
//	derivedKey := key
//
//	// m/<purpose>'/<coin type>' 分解路径
//	elements := strings.Split(path, "/")
//	//log.Println(elements)
//	for i, elem := range elements {
//		if len(elem) == 0 {
//			continue
//		}
//		var value common.String
//		hardened := false
//		if strings.Index(elem, "'") == len(elem)-1 {
//			hardened = true
//			elem = elem[0 : len(elem)-1]
//		}
//
//		value = common.NewString(elem)
//		if i >= 0 && value.String() == elem {
//			if hardened {
//				derivedKey, err = derivedKey.Child(hdkeychain.HardenedKeyStart + value.UInt32())
//			} else {
//				derivedKey, err = derivedKey.Child(value.UInt32())
//			}
//			if err != nil {
//				return nil, err
//			}
//		} else {
//			return nil, ErrInvalidDerivedPath
//		}
//	}
//
//	return derivedKey, err
//}

//Mnemonic 密钥助记词
//func (k *HDKey) Mnemonic() string {
//	mnemonic, _ := bip39.NewMnemonic(k.seed)
//	return mnemonic
//}

//FileName 文件名
func (k *HDKey) FileName() string {
	return KeyFileName(k.Alias, k.KeyID)
}

//Seed 密钥种子
func (k *HDKey) Seed() []byte {
	return k.seed
}

// EncryptKey encrypts a key using the specified scrypt parameters into a json
// blob that can be decrypted later on.
func EncryptKey(hdkey *HDKey, auth string, scryptN, scryptP int) ([]byte, error) {

	authArray := []byte(auth)

	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptR, scryptP, scryptDKLen)
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:16]

	keyBytes := hdkey.seed

	iv := make([]byte, aes.BlockSize) // 16
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
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

	encryptedHDKeyJSON := encryptedHDKeyJSON{
		Alias:    hdkey.Alias,
		KeyID:    hdkey.KeyID,
		Crypto:   cryptoStruct,
		RootPath: hdkey.RootPath,
		Version:  version,
	}
	return json.MarshalIndent(encryptedHDKeyJSON, "", "\t")
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
		seed []byte
		err  error
	)
	k := new(encryptedHDKeyJSON)
	if err := json.Unmarshal(keyjson, k); err != nil {
		return nil, err
	}

	seed, err = decryptHDKey(k, auth)
	// Handle any decryption errors and return the key
	if err != nil {
		return nil, err
	}

	keyID := computeKeyID(seed)

	return &HDKey{
		Alias:    k.Alias,
		KeyID:    keyID,
		RootPath: k.RootPath,
		seed:     seed,
	}, nil
}

// decryptHDKey 解密HDKey的文件内容
func decryptHDKey(keyProtected *encryptedHDKeyJSON, auth string) (keyBytes []byte, err error) {

	if keyProtected.Crypto.Cipher != "aes-128-ctr" {
		return nil, fmt.Errorf("Cipher not supported: %v", keyProtected.Crypto.Cipher)
	}

	mac, err := hex.DecodeString(keyProtected.Crypto.MAC)
	if err != nil {
		return nil, err
	}

	iv, err := hex.DecodeString(keyProtected.Crypto.CipherParams.IV)
	if err != nil {
		return nil, err
	}

	cipherText, err := hex.DecodeString(keyProtected.Crypto.CipherText)
	if err != nil {
		return nil, err
	}

	derivedKey, err := getKDFKey(keyProtected.Crypto, auth)
	if err != nil {
		return nil, err
	}

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, ErrDecrypt
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, err
	}

	return plainText, err
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

// NewHDKey 通过userkey，私钥种子，根私钥标识符，账户路径，创建HDKey
func NewHDKey(seed []byte, alias, rootPath string) (*HDKey, error) {

	keyID := computeKeyID(seed)

	//实例化密钥
	hdkey := &HDKey{
		Alias:    alias,
		KeyID:    keyID,
		RootPath: rootPath,
		seed:     seed,
	}

	return hdkey, nil
}

// GenerateSeed returns a cryptographically secure random seed that can be used
// as the input for the NewMaster function to generate a new master node.
//
// The length is in bytes and it must be between 16 and 64 (128 to 512 bits).
// The recommended length is 32 (256 bits) as defined by the RecommendedSeedLen
// constant.
func GenerateSeed(length uint8) ([]byte, error) {
	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
	if length < MinSeedBytes || length > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}

	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

//writeKeyFile 写入HDKey结构内容到文件
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

//computeKeyID 计算HDKey的KeyID
func computeKeyID(seed []byte) string {

	//seed 通过hmac-sha256 两次 RIPEMD160 一次 得到keyID

	hmac256 := hmac.New(sha3.New256, masterKey)
	hmac256.Write(seed)
	keyID := hmac256.Sum(nil)

	hmac256 = hmac.New(sha3.New256, masterKey)
	hmac256.Write(keyID)
	keyID = hmac256.Sum(nil)

	keyID = owcrypt.Hash(keyID, 0, owcrypt.HASH_ALG_RIPEMD160)

	return owkeychain.Base58checkEncode(keyID, KeyIDVer)
}

// keyFileName implements the naming convention for keyfiles:
// wallet--<alias>-<rootId>
func KeyFileName(alias, rootId string) string {
	//ts := time.Now().UTC()
	return fmt.Sprintf("%s-%s", alias, rootId)
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
