/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package hdkeystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/awnumar/memguard"

	"github.com/blocktree/go-owcdrivers/owkeychain"
	"github.com/blocktree/go-owcrypt"
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

	//KeyID首字节的标识
	KeyIDVer = []byte{0x48}
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
	//账户的扩展ID
	KeyID string
	//种子，只用于生成钱包文件临时使用
	//seed []byte
	//锁定内存的加密种子对象, 写入钱包文件的时候会作为临时locker使用
	encryptedSeed *memguard.LockedBuffer
	aadCall       AADCall
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
	Cipher     string      `json:"cipher"`
	CipherText string      `json:"ciphertext"`
	KDF        string      `json:"kdf"`
	KDFParams  interface{} `json:"kdfparams"`
	MAC        string      `json:"mac"`
}

type argon2KDFParam struct {
	Memory  int    `json:"memory"`
	Time    int    `json:"time"`
	Threads int    `json:"threads"`
	Keylen  int    `json:"keylen"`
	Salt    string `json:"salt"`
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
// DerivedKeyWithPath 安全地派生子密钥
func (k *HDKey) DerivedKeyWithPath(path string, curveType uint32) (*owkeychain.ExtendedKey, error) {
	var derivedKey *owkeychain.ExtendedKey

	err := k.Seed(func(seed []byte) error {
		var err error
		derivedKey, err = owkeychain.DerivedPrivateKeyWithPath(seed, path, curveType)
		return err
	})

	return derivedKey, err
}

func DerivedLockerKeyWithPath(seed *memguard.LockedBuffer, path string, curveType uint32) (*memguard.LockedBuffer, error) {
	derivedKey, err := owkeychain.DerivedPrivateKeyWithPath(seed.Bytes(), path, curveType)
	if err != nil {
		return nil, err
	}
	prkBytes, err := derivedKey.GetPrivateKeyBytes()
	if err != nil {
		return nil, err
	}
	// 创建安全 buffer
	result := memguard.NewBufferFromBytes(prkBytes)
	// 安全清零临时明文（逐字节）
	ClearData(prkBytes)
	return result, err
}

func (k *HDKey) SetAADCall(call AADCall) {
	k.aadCall = call
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

// FileName 文件名
func (k *HDKey) FileName() string {
	return KeyFileName(k.Alias, k.KeyID)
}

// Seed 密钥种子，该方法应该不允许再使用
//func (k *HDKey) Seed() []byte {
//	return decryptSeed(k.encryptedSeed.Bytes())
//}

// Seed 安全地提供种子访问，通过回调确保明文及时清零
func (k *HDKey) Seed(fn func(seed []byte) error) error {
	// 1. 复制加密数据
	encrypted := make([]byte, k.encryptedSeed.Size())
	copy(encrypted, k.encryptedSeed.Bytes())
	defer ClearData(encrypted) // 使用 defer 确保清零

	// 2. 创建临时锁定缓冲区存储明文种子
	seedBuf := memguard.NewBuffer(SeedLen) // 根据实际 seed 长度调整

	defer seedBuf.Destroy()

	// 3. 直接解密到锁定内存
	if err := decryptSeedTo(seedBuf.Bytes(), encrypted, k.KeyID, k.aadCall); err != nil {
		return err
	}

	// 4. 调用回调
	return fn(seedBuf.Bytes())
}

// DestroySeed 主动清零加密内存的种子数据
func (k *HDKey) DestroySeed() {
	k.encryptedSeed.Destroy()
}

// NewHDKey 通过userkey，私钥种子，根私钥标识符，账户路径，创建HDKey
func NewHDKey(seed []byte, alias, rootPath string) (*HDKey, error) {

	keyID := computeKeyID(seed)

	//实例化密钥
	hdkey := &HDKey{
		Alias:    alias,
		KeyID:    keyID,
		RootPath: rootPath,
		//seed:     seed,
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

// GenerateLockedSeed 生成种子并直接放入锁定内存
func GenerateLockedSeed(size int) (*memguard.LockedBuffer, error) {
	return memguard.NewBufferRandom(size), nil
}

// writeKeyFile 写入HDKey结构内容到文件
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

// computeKeyID 计算HDKey的KeyID, 这个版本有加入masterKey可能造成确定性破坏攻击
//func computeKeyID(seed []byte) string {
//
//	//seed 通过hmac-sha256 两次 RIPEMD160 一次 得到keyID
//
//	hmac256 := hmac.New(sha3.New256, masterKey)
//	hmac256.Write(seed)
//	keyID := hmac256.Sum(nil)
//
//	hmac256 = hmac.New(sha3.New256, masterKey)
//	hmac256.Write(keyID)
//	keyID = hmac256.Sum(nil)
//
//	keyID = owcrypt.Hash(keyID, 0, owcrypt.HASH_ALG_RIPEMD160)
//
//	return owkeychain.Base58checkEncode(keyID, KeyIDVer)
//}

// computeKeyID 从种子生成唯一、不可逆、抗碰撞的钱包标识符
// 算法：RIPEMD160(SHA256(seed)) → Base58Check(version + hash)
func computeKeyID(seed []byte) string {
	// Step 1: SHA256(seed)
	sha256Hash := sha256.Sum256(seed)

	// Step 2: RIPEMD160(SHA256(seed))
	ripemd160Hash := owcrypt.Hash(sha256Hash[:], 0, owcrypt.HASH_ALG_RIPEMD160)

	// Step 3: Base58Check 编码（带版本字节）
	return owkeychain.Base58checkEncode(ripemd160Hash, KeyIDVer)
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

// ClearData 最终清空底层数组的函数
func ClearData(slices ...[]byte) {
	for _, s := range slices {
		for i := range s { // range 自动处理 len=0 的情况
			s[i] = 0
		}
	}
}
