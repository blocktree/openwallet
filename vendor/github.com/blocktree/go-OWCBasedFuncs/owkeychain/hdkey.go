package owkeychain

import (
	"errors"
	"strconv"
	"strings"

	"github.com/blocktree/go-OWCrypt"
)

var (
	ErrInvalidDerivedPath = errors.New("Invalid DerivedPath")
	ErrKeyIsNotPrivate    = errors.New("The key is not private")
)

//接口仅接受绝对路径
func DerivedPrivateKeyWithPath(seed []byte, derivedPath string, curveType uint32) (*ExtendedKey, error) {

	//移除空格
	path := strings.Replace(derivedPath, " ", "", -1)

	if path == "m" || path == "/" || path == "" {
		return InitRootKeyFromSeed(seed, curveType) //根私钥
	}

	if strings.Index(path, "m/") != 0 {
		return nil, ErrInvalidDerivedPath
	}

	priKey, err := InitRootKeyFromSeed(seed, curveType)
	if err != nil {
		return nil, err
	}

	path = path[2:]
	elements := strings.Split(path, "/")

	for _, elem := range elements {
		var hdSerializes uint32
		if len(elem) == 0 {
			return nil, ErrInvalidDerivedPath
		}

		if strings.Index(elem, "'") == len(elem)-1 {
			elem = elem[0 : len(elem)-1]
			index, err := strconv.Atoi(elem)
			if err != nil {
				return nil, ErrInvalidDerivedPath
			}
			hdSerializes = uint32(index + HardenedKeyStart)
		} else {
			index, err := strconv.Atoi(elem)
			if err != nil {
				return nil, ErrInvalidDerivedPath
			}
			hdSerializes = uint32(index)
		}

		priKey, err = priKey.GenPrivateChild(hdSerializes)
		if err != nil {
			return nil, err
		}

	}
	return priKey, nil
}

func GetCoinRootPublicKey(seed []byte, coinType CoinType) (*ExtendedKey, error) {
	tmpPrikey, err := DerivedPrivateKeyWithPath(seed, openwalletPrePath, coinType.curveType)
	if err != nil {
		return nil, err
	}
	coinRootPublicKey, err := tmpPrikey.GenPublicChild(coinType.hdIndex)
	if err != nil {
		return nil, err
	}
	return coinRootPublicKey, nil
}

func DerivedPrivateKeyBytes(seed []byte, coinType CoinType, serializes uint32) ([]byte, error) {
	tmpPrikey, err := DerivedPrivateKeyWithPath(seed, openwalletPrePath, coinType.curveType)
	if err != nil {
		return nil, err
	}
	coinRootPrivateKey, err := tmpPrikey.GenPrivateChild(coinType.hdIndex)
	if err != nil {
		return nil, err
	}
	privateKey, err := coinRootPrivateKey.GenPrivateChild(serializes)
	if err != nil {
		return nil, err
	}
	return privateKey.key, nil
}

func (k *ExtendedKey) DerivedPublicKeyFromSerializes(serializes uint32) (*ExtendedKey, error) {
	return k.GenPublicChild(serializes)
}

//GetPublicKey 获取当前密钥对应的公钥
func (k *ExtendedKey) GetPublicKeyBytes() []byte {
	if k.isPrivate {
		return owcrypt.Point_mulBaseG(k.key, k.curveType)
	}
	return k.key
}

//GetPrivateKey 获取当前密钥对应的私钥数组
func (k *ExtendedKey) GetPrivateKeyBytes() ([]byte, error) {
	if k.isPrivate {
		return k.key, nil
	}
	return nil, ErrKeyIsNotPrivate
}
