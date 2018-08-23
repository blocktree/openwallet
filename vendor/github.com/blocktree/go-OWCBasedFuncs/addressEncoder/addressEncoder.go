package addressEncoder

import (
	"errors"

	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder/base32PolyMod"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder/bech32"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder/blake256"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder/eip55"
	"github.com/blocktree/go-OWCrypt"
)

var (
	ErrorInvalidHashLength = errors.New("Invalid hash length!")
	ErrorInvalidAddress    = errors.New("Invalid address!")
)

func calcChecksum(data []byte, chkType string) []byte {
	if chkType == "doubleSHA256" {
		return owcrypt.Hash(data, 0, owcrypt.HASh_ALG_DOUBLE_SHA256)[:4]
	}
	if chkType == "doubleBlake256" {
		return blake256.DoubleBlake256(data)[:4]
	}
	return nil
}

func verifyChecksum(data []byte, chkType string) bool {
	checksum := calcChecksum(data[:len(data)-4], chkType)
	for i := 0; i < 4; i++ {
		if checksum[i] != data[len(data)-4+i] {
			return false
		}
	}
	return true
}

func catData(data1 []byte, data2 []byte) []byte {
	if data2 == nil {
		return data1
	}
	return append(data1, data2...)
}

func recoverData(data, prefix, suffix []byte) ([]byte, error) {
	for i := 0; i < len(prefix); i++ {
		if data[i] != prefix[i] {
			return nil, ErrorInvalidAddress
		}
	}
	if suffix != nil {
		for i := 0; i < len(suffix); i++ {
			if data[len(data)-len(suffix)+i] != suffix[i] {
				return nil, ErrorInvalidAddress
			}
		}
	}

	if suffix == nil {
		return data[len(prefix):], nil
	}
	return data[len(prefix) : len(data)-len(suffix)], nil
}

func encodeData(data []byte, encodeType string, alphabet string) string {
	if encodeType == "base58" {
		return Base58Encode(data, NewBase58Alphabet(alphabet))
	}
	return ""
}

func decodeData(data, encodeType, alphabet, checkType string, prefix, suffix []byte) ([]byte, error) {
	if encodeType == "base58" {
		ret, err := Base58Decode(data, NewBase58Alphabet(alphabet))
		if err != nil {
			return nil, ErrorInvalidAddress
		}

		if verifyChecksum(ret, checkType) == false {
			return nil, ErrorInvalidAddress
		}
		return recoverData(ret[:len(ret)-4], prefix, suffix)
	}
	return nil, nil
}

func calcHash(data []byte, hashType string) []byte {
	if hashType == "h160" {
		return owcrypt.Hash(data, 0, owcrypt.HASH_ALG_HASH160)
	}
	if hashType == "blake2b160" {
		return owcrypt.Hash(data, 20, owcrypt.HASH_ALG_BLAKE2B)
	}
	return nil
}

func AddressEncode(hash []byte, addresstype AddressType) string {

	if len(hash) != addresstype.hashLen {
		hash = calcHash(hash, addresstype.hashType)
	}

	if addresstype.encodeType == "bech32" {
		return bech32.Encode(addresstype.checksumType, addresstype.alphabet, hash)
	}
	if addresstype.encodeType == "base32PolyMod" {
		return base32PolyMod.Encode(addresstype.checksumType, addresstype.alphabet, hash)
	}
	if addresstype.encodeType == "eip55" {
		return eip55.Eip55_encode(hash)
	}
	data := catData(catData(addresstype.prefix, hash), addresstype.suffix)
	return encodeData(catData(data, calcChecksum(data, addresstype.checksumType)), addresstype.encodeType, addresstype.alphabet)

}

func AddressDecode(address string, addresstype AddressType) ([]byte, error) {
	if addresstype.encodeType == "bech32" {
		ret, err := bech32.Decode(address, addresstype.alphabet)
		if err != nil {
			return nil, ErrorInvalidAddress
		}
		if len(ret) != addresstype.hashLen {
			return nil, ErrorInvalidHashLength
		}
		return ret, nil
	}
	if addresstype.encodeType == "base32PolyMod" {
		ret, err := base32PolyMod.Decode(address, addresstype.alphabet)
		if err != nil {
			return nil, ErrorInvalidAddress
		}
		if len(ret) != addresstype.hashLen {
			return nil, ErrorInvalidHashLength
		}
		return ret, nil
	}
	if addresstype.encodeType == "eip55" {
		ret, err := eip55.Eip55_decode(address)
		if err != nil {
			return nil, ErrorInvalidAddress
		}
		if len(ret) != addresstype.hashLen {
			return nil, ErrorInvalidHashLength
		}
		return ret, nil
	}
	data, err := decodeData(address, addresstype.encodeType, addresstype.alphabet, addresstype.checksumType, addresstype.prefix, addresstype.suffix)
	if err != nil {
		return nil, err
	}

	if len(data) != addresstype.hashLen {
		return nil, ErrorInvalidHashLength
	}

	return data, nil
}
