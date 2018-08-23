package owkeychain

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/blocktree/go-OWCrypt"
)

const (
	// 用于产生根密钥的种子长度
	RecommendedSeedLen = 32
	// 强化子密钥索引号起始值
	HardenedKeyStart = 0x80000000
	// 扩展深度限制
	maxUint8 = 1<<8 - 1
)

var (
	// 公钥不能进行强化扩展
	ErrDeriveHardFromPublic = errors.New("cannot derive a hardened key from a public key")

	// 超过最大扩展深度
	ErrDeriveBeyondMaxDepth = errors.New("cannot derive a key with more than 255 indices in its path")

	// 公钥无法扩展子私钥
	ErrNotPrivExtKey = errors.New("unable to create private keys from a public extended key")

	// 该索引号无法扩展，应尝试下一个
	ErrInvalidChild = errors.New("the extended key at this index is invalid")

	// 种子长度错误
	ErrInvalidSeedLen = fmt.Errorf("seed length should be 256")

	// reserve
	ErrBadChecksum = errors.New("bad extended key checksum")

	// reserve
	ErrInvalidKeyLen = errors.New("the provided serialized extended key length is invalid")
)

//ExtendedKey 扩展密钥结构体
type ExtendedKey struct {
	depth      uint8  //深度
	parentFP   []byte //父密钥指纹
	serializes uint32 //序列号
	chainCode  []byte //链码
	key        []byte //密钥数据
	isPrivate  bool   //当前密钥的私钥标记
	curveType  uint32 //曲线类型
}

//NewExtendedKey 初始化密钥结构体
func NewExtendedKey(key, chainCode, parentFP []byte, depth uint8,
	serializes uint32, isPrivate bool, curveType uint32) *ExtendedKey {

	return &ExtendedKey{
		depth:      depth,
		parentFP:   parentFP,
		serializes: serializes,
		chainCode:  chainCode,
		key:        key,
		isPrivate:  isPrivate,
		curveType:  curveType,
	}
}

func getI(data, key []byte, serializes, typeChoose uint32) []byte {
	tmp := [4]byte{}
	hmac512 := hmac.New(sha512.New, key)
	binary.BigEndian.PutUint32(tmp[:], serializes)
	if len(data) == 32 {
		hmac512.Write([]byte{0})
	}
	hmac512.Write(data)
	hmac512.Write(tmp[:])
	return hmac512.Sum(nil)
}

func inverse(data []byte) []byte {
	ret := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		ret[i] = data[len(data)-1-i]
	}
	return ret
}

func getPriChildViaPriParent(il, prikey []byte, typeChoose uint32) ([]byte, error) {
	priChild := []byte{}
	if typeChoose == owcrypt.ECC_CURVE_ED25519 {
		ilNum := new(big.Int).SetBytes(inverse(il[:28]))
		kpr := new(big.Int).SetBytes(inverse(prikey))
		num8 := new(big.Int).SetBytes([]byte{8})
		curveOrder := new(big.Int).SetBytes(owcrypt.GetCurveOrder(typeChoose))
		ilNum.Mul(ilNum, num8)
		ilNum.Add(ilNum, kpr)
		check := new(big.Int).Mod(ilNum, curveOrder)
		if check.Sign() == 0 {
			return nil, ErrInvalidChild
		}
		priChild = inverse(ilNum.Bytes())
	} else {
		ilNum := new(big.Int).SetBytes(il)
		curveOrder := new(big.Int).SetBytes(owcrypt.GetCurveOrder(typeChoose))
		if ilNum.Cmp(curveOrder) >= 0 || ilNum.Sign() == 0 {
			return nil, ErrInvalidChild
		}
		kpr := new(big.Int).SetBytes(prikey)
		ilNum.Add(ilNum, kpr)
		ilNum.Mod(ilNum, curveOrder)
		if ilNum.Sign() == 0 {
			return nil, ErrInvalidChild
		}
		priChild = ilNum.Bytes()
	}
	return priChild, nil
}

func getPubChildViaPubParent(il, pubkey []byte, typeChoose uint32) ([]byte, error) {
	if typeChoose == owcrypt.ECC_CURVE_ED25519 {
		ilNum := new(big.Int).SetBytes(inverse(il[:28]))
		num8 := new(big.Int).SetBytes([]byte{8})
		ilNum.Mul(ilNum, num8)
		il2 := inverse(ilNum.Bytes())
		point, isinfinity := owcrypt.Point_mulBaseG_add(pubkey, il2, typeChoose)
		if isinfinity {
			return nil, ErrInvalidChild
		}
		return point, nil
	}
	ilNum := new(big.Int).SetBytes(il)
	curveOrder := new(big.Int).SetBytes(owcrypt.GetCurveOrder(typeChoose))
	if ilNum.Cmp(curveOrder) >= 0 || ilNum.Sign() == 0 {
		return nil, ErrInvalidChild
	}
	parentPubPoint := owcrypt.PointDecompress(pubkey, typeChoose)
	point, isinfinity := owcrypt.Point_mulBaseG_add(parentPubPoint[1:], il, typeChoose)
	if isinfinity {
		return nil, ErrInvalidChild
	}
	point = owcrypt.PointCompress(point, typeChoose)
	return point, nil
}

func getFP(key []byte, isPrivate bool, typeChoose uint32) []byte {
	fingerPrint := []byte{}
	if !isPrivate {
		fingerPrint = owcrypt.Hash(key, 0, owcrypt.HASH_ALG_HASH160)[:4]
	} else {
		pubkey := owcrypt.Point_mulBaseG(key, typeChoose)
		fingerPrint = owcrypt.Hash(pubkey, 0, owcrypt.HASH_ALG_HASH160)[:4]
	}
	return fingerPrint
}

//GenPrivateChild 通过k扩展子私钥
func (k *ExtendedKey) GenPrivateChild(serializes uint32) (*ExtendedKey, error) {
	i := []byte{}
	childChainCode := []byte{}
	//越过最大深度限制
	if k.depth == maxUint8 {
		return nil, ErrDeriveBeyondMaxDepth
	}
	//不能从父公钥扩展子私钥
	if !k.isPrivate {
		return nil, ErrNotPrivExtKey
	}

	if serializes >= HardenedKeyStart { //强化扩展
		i = getI(k.key, k.chainCode, serializes, k.curveType)
	} else { //普通扩展
		point := owcrypt.Point_mulBaseG(k.key, k.curveType)
		i = getI(point[:], k.chainCode, serializes, k.curveType)
	}

	childKey, err := getPriChildViaPriParent(i[:32], k.key, k.curveType)

	if err != nil {
		return nil, err
	}

	childChainCode = i[32:]

	parentFP := getFP(k.key, k.isPrivate, k.curveType)
	return NewExtendedKey(childKey, childChainCode, parentFP, k.depth+1,
		serializes, true, k.curveType), nil
}

//GenPublicChild 通过k扩展子公钥
func (k *ExtendedKey) GenPublicChild(serializes uint32) (*ExtendedKey, error) {
	if !k.isPrivate {
		if serializes >= HardenedKeyStart { //不能从父公钥强化扩展
			return nil, ErrDeriveHardFromPublic
		}
		i := getI(k.key, k.chainCode, serializes, k.curveType)
		childKey, err := getPubChildViaPubParent(i[:32], k.key, k.curveType)
		if err != nil {
			return nil, err
		}
		childChainCode := i[len(i)/2:]

		parentFP := getFP(k.key, false, k.curveType)
		return NewExtendedKey(childKey, childChainCode, parentFP, k.depth+1,
			serializes, false, k.curveType), nil

	}
	childPrikey, err := k.GenPrivateChild(serializes)
	if err != nil {
		return nil, err
	}
	childKey := owcrypt.Point_mulBaseG(childPrikey.key, childPrikey.curveType)
	return NewExtendedKey(childKey, childPrikey.chainCode, childPrikey.parentFP, k.depth+1,
		serializes, false, k.curveType), nil

}

//InitRootKeyFromSeed 通过种子获得根私钥
func InitRootKeyFromSeed(seed []byte, curveType uint32) (*ExtendedKey, error) {
	if len(seed) != RecommendedSeedLen {
		return nil, ErrInvalidSeedLen
	}
	ctx := sha512.New()
	ctx.Write(seed)
	i := ctx.Sum(nil)
	if curveType == owcrypt.ECC_CURVE_ED25519 {
		i[0] &= 248
		i[31] &= 63
		i[31] |= 64
	}
	rootParentFP := [4]byte{0, 0, 0, 0}
	return NewExtendedKey(i[:32], i[32:], rootParentFP[:], 0, 0, true, curveType), nil
}
