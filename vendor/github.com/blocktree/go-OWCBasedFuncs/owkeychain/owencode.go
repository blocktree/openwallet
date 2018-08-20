package owkeychain

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"strings"

	"github.com/blocktree/go-OWCrypt"
)

var (
	ErrInvalidAddress = errors.New("address is invalid")
)

func Base58checkEncode(data []byte, fix []byte) string {
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

func Base58checkDecode(address string, fix []byte) ([]byte, error) {
	decodeBytes, err := Decode(address, BitcoinAlphabet)

	if err != nil {
		return nil, ErrInvalidAddress
	}
	for i := 0; i < len(fix); i++ {
		if fix[i] != decodeBytes[i] {
			return nil, ErrInvalidAddress
		}
	}

	checksum := owcrypt.Hash(decodeBytes[:len(decodeBytes)-4], 0, owcrypt.HASh_ALG_DOUBLE_SHA256)[:4]

	for i := 0; i < 4; i++ {
		if checksum[i] != decodeBytes[len(decodeBytes)-4+i] {
			return nil, ErrInvalidAddress
		}
	}
	return decodeBytes[len(fix) : len(decodeBytes)-4], nil
}

func uint32ToBytes(i uint32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}

func bytesToUInt32(buf []byte) uint32 {
	return uint32(binary.BigEndian.Uint32(buf))
}

func (k *ExtendedKey) OWEncode() string {
	data := []byte{}
	data = append(data[:], uint32ToBytes(k.curveType)...)
	data = append(data[:], []byte{k.depth}...)
	data = append(data[:], k.parentFP...)
	data = append(data[:], uint32ToBytes(k.serializes)...)
	data = append(data[:], k.chainCode...)
	if k.curveType == owcrypt.ECC_CURVE_ED25519 || k.isPrivate == true {
		data = append(data[:], []byte{0}...)
	}
	data = append(data[:], k.key...)

	if k.isPrivate {
		return Base58checkEncode(data, owprvPrefix)
	}
	return Base58checkEncode(data, owpubPrefix)
}

func OWDecode(data string) (*ExtendedKey, error) {
	privateFlag := true
	decodeBytes := []byte{}
	var err error
	if strings.Index(data, "owpub") == 0 {
		privateFlag = false
	} else if strings.Index(data, "owprv") == 0 {
		privateFlag = true
	} else {
		return nil, ErrInvalidAddress
	}
	if privateFlag {
		decodeBytes, err = Base58checkDecode(data, owprvPrefix)
		if err != nil {
			return nil, err
		}
	} else {
		decodeBytes, err = Base58checkDecode(data, owpubPrefix)
		if err != nil {
			return nil, err
		}
	}
	curveType := bytesToUInt32(decodeBytes[:4])
	depth := decodeBytes[4]
	parentFP := decodeBytes[5:9]
	serializes := bytesToUInt32(decodeBytes[9:13])
	chainCode := decodeBytes[13:45]
	if curveType == owcrypt.ECC_CURVE_ED25519 || privateFlag {
		if decodeBytes[45] != 0 {
			return nil, ErrInvalidAddress
		}
		key := decodeBytes[46:]

		return NewExtendedKey(key, chainCode, parentFP, depth, serializes, privateFlag, curveType), nil
	}
	key := decodeBytes[45:]
	return NewExtendedKey(key, chainCode, parentFP, depth, serializes, privateFlag, curveType), nil
}
