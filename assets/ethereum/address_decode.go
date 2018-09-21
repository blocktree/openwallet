package ethereum

import (
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	owcrypt "github.com/blocktree/go-OWCrypt"
)

type addressDecoder struct{}

//PrivateKeyToWIF 私钥转WIF
func (decoder *addressDecoder) PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error) {
	return "", nil

}

//PublicKeyToAddress 公钥转地址
func (decoder *addressDecoder) PublicKeyToAddress(pub []byte, isTestnet bool) (string, error) {

	cfg := addressEncoder.ETH_mainnetPublicAddress
	if isTestnet {
		cfg = addressEncoder.ETH_mainnetPublicAddress
	}

	//pkHash := btcutil.Hash160(pub)
	//address, err :=  btcutil.NewAddressPubKeyHash(pkHash, &cfg)
	//if err != nil {
	//	return "", err
	//}

	//log.Debug("public key:", common.ToHex(pub))
	publickKey := owcrypt.PointDecompress(pub, owcrypt.ECC_CURVE_SECP256K1)
	//log.Debug("after encode public key:", common.ToHex(publickKey))
	pkHash := owcrypt.Hash(publickKey[1:len(publickKey)], 0, owcrypt.HASH_ALG_KECCAK256)

	address := addressEncoder.AddressEncode(pkHash, cfg)

	return address, nil

}

//RedeemScriptToAddress 多重签名赎回脚本转地址
func (decoder *addressDecoder) RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error) {
	return "", nil

}

//WIFToPrivateKey WIF转私钥
func (decoder *addressDecoder) WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error) {
	return nil, nil

}
