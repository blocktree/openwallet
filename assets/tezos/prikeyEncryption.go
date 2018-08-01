package tezos

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"hash"

	"golang.org/x/crypto/nacl/secretbox"
	"math/rand"
	"time"
	"golang.org/x/crypto/ed25519"
)

var (
	SaltLen = 8
)

func base58checkEncode(data []byte, fix []byte) string {
	ctx := sha256.New()
	ctx.Write(fix[:])
	ctx.Write(data[:])
	hash := ctx.Sum(nil)
	ctx = sha256.New()
	ctx.Write(hash)
	hash = ctx.Sum(nil)

	fix_value := []byte{}
	fix_value = append(fix, data...)
	fix_value = append(fix_value, hash[:4]...)

	return Encode(fix_value[:], BitcoinAlphabet)
}

func base58checkDecodeNormal(data string, fix []byte) ([]byte, error) {
	value, err := Decode(data, BitcoinAlphabet)
	if err == ErrorInvalidBase58String {
		return nil, err
	}

	for i := 0; i < len(fix); i++ {
		if value[i] != fix[i] {
			err = ErrorInvalidBase58String
			return nil, err
		}
	}
	ctx := sha256.New()
	ctx.Write(value[:len(value)-4])
	hash := ctx.Sum(nil)
	ctx = sha256.New()
	ctx.Write(hash)
	hash = ctx.Sum(nil)

	for i := 0; i < 4; i++ {
		if hash[i] != value[len(value)-4+i] {
			err = ErrorInvalidBase58String
			return nil, err
		}
	}

	return value[len(fix) : len(value)-4], err
}
func base58checkDecode(data string, fix []byte) ([]byte, []byte, error) {
	value, err := Decode(data, BitcoinAlphabet)
	if err == ErrorInvalidBase58String {
		return nil, nil, err
	}

	for i := 0; i < len(fix); i++ {
		if value[i] != fix[i] {
			err = ErrorInvalidBase58String
			return nil, nil, err
		}
	}
	ctx := sha256.New()
	ctx.Write(value[:len(value)-4])
	hash := ctx.Sum(nil)
	ctx = sha256.New()
	ctx.Write(hash)
	hash = ctx.Sum(nil)

	for i := 0; i < 4; i++ {
		if hash[i] != value[len(value)-4+i] {
			err = ErrorInvalidBase58String
			return nil, nil, err
		}
	}

	return value[len(fix) : len(fix)+SaltLen], value[len(fix)+SaltLen : len(value)-4], err
}

func getKey(password, salt []byte, iterator, keyLen int, h func() hash.Hash) [32]byte {
	ret := [32]byte{}
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	U := make([]byte, hashLen)
	for block := 1; block <= numBlocks; block++ {
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)
		prf.Write(buf[:4])
		dk = prf.Sum(dk)
		T := dk[len(dk)-hashLen:]
		copy(U, T)
		for n := 2; n <= iterator; n++ {
			prf.Reset()
			prf.Write(U)
			U = U[:0]
			U = prf.Sum(U)
			for x := range U {
				T[x] ^= U[x]
			}
		}
	}
	for i := 0; i < 32; i++ {
		ret[i] = dk[i]
	}
	return ret
}

func Encrypt(password, key string) (string, error) {
	edsk := [4]byte{43, 246, 78, 7}
	edsk2 := [4]byte{13, 15, 58, 7}
	edesk := [5]byte{0x07, 0x5A, 0x3C, 0xB3, 0x29}
	nonce := [24]byte{}
	value := []byte{}
	err := errors.New("")

	if len(key) == 54 {
		value, err = base58checkDecodeNormal(key, edsk2[:])
	} else if len(key) == 98 {
		value, err = base58checkDecodeNormal(key, edsk[:])
	} else {
		err = ErrorInvalidBase58String
		return "", err
	}
	if err != nil {
		return "", err
	}

	//salt := [8]byte{0x0c, 0x4b, 0x58, 0x71, 0x29, 0xb3, 0xb5, 0x29} //自行替换成随机数
	var salt [8]byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i:=0; i<8; i++ {
		salt[i] = byte(r.Intn(255))
	}

	pswd := []byte(password)
	subKey := getKey(pswd[:], salt[:], 0x8000, 32, sha512.New)
	encryptedKey := secretbox.Seal(nonce[:], value[:32], &nonce, &subKey)

	splice := []byte{}
	splice = append(salt[:], encryptedKey[24:]...)
	ret := base58checkEncode(splice[:], edesk[:])
	return ret, nil
}

func Decrypt(password, encryptedKey string) (string, error) {
	edsk := []byte{43, 246, 78, 7}
	edesk := [5]byte{0x07, 0x5A, 0x3C, 0xB3, 0x29}
	nonce := [24]byte{}

	salt, encKey, err := base58checkDecode(encryptedKey, edesk[:])

	if err == ErrorInvalidBase58String {
		return "", err
	}
	pswd := []byte(password)
	subKey := getKey(pswd[:], salt[:], 0x8000, 32, sha512.New)

	key, ret := secretbox.Open(nil, encKey[:], &nonce, &subKey)
	testkey := ed25519.NewKeyFromSeed(key[:])

	if ret {
		decryptKey := base58checkEncode(testkey[:], edsk[:])
		return decryptKey, nil
	} else {
		return "", ErrorInvalidBase58String
	}

}
