package keystore

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/go-owcdrivers/owkeychain"
	owcrypt "github.com/blocktree/go-owcrypt"
)

func GenerateKeyPairWithHdKey(addressIndex int) (*PrivateKey, *PublicKey, error) {
	//种子
	seed, err := hdkeystore.GenerateSeed(32)
	if err != nil {
		//
		log.Errorf("generate seed failed, err=%v", err)
		return nil, nil, err
	}

	//2层 钱包私钥
	walletPath := fmt.Sprintf("%s/%d'", hdkeystore.OpenwCoinTypePath, 1)
	walletPriKey, err := owkeychain.DerivedPrivateKeyWithPath(seed, walletPath, owcrypt.ECC_CURVE_SECP256R1)
	if err != nil {
		log.Errorf("generate child key failed, err=", err)
		return nil, nil, err
	}

	log.Debugf("start:%v", common.FormatStruct(walletPriKey))

	walletPubKey := walletPriKey.GetPublicKey()

	start, err := walletPubKey.GenPublicChild(0)
	if err != nil {
		log.Errorf("pubkey.GenPublicChild failed, err = %v", err)
		return nil, nil, err
	}

	log.Debugf("start:%v", common.FormatStruct(start))

	derivedPubKey, err := start.GenPublicChild(uint32(addressIndex))
	if err != nil {
		log.Errorf("start.GenPublicChild failed, err = %v", err)
		return nil, nil, err
	}

	log.Debugf("derivedPubKey:%v", common.FormatStruct(derivedPubKey))

	derivedPath := fmt.Sprintf("%s/%d/%d", walletPath, 0, addressIndex)

	derivedPriKey, err := owkeychain.DerivedPrivateKeyWithPath(seed, derivedPath, owcrypt.ECC_CURVE_SECP256R1)
	if err != nil {
		log.Errorf("derived private key failed, err=%v", err)
		return nil, nil, err
	}

	c, err := GetCurve(P256)
	if err != nil {
		log.Errorf("get curve failed, err=%v", err)
		return nil, nil, err
	}

	priKeyBytes, err := derivedPriKey.GetPrivateKeyBytes()
	if err != nil {
		log.Errorf("get privated key bytes failed, err=%v", err)
		return nil, nil, err
	}

	if len(priKeyBytes) != 32 {
		log.Errorf("private key length failed, should be 32 instead of %v", len(priKeyBytes))
		return nil, nil, errors.New("private key length failed")
	}

	pubKeyBytes := derivedPubKey.GetPublicKeyBytes()
	pubKeyBytes = owcrypt.PointDecompress(pubKeyBytes, owcrypt.ECC_CURVE_SECP256R1)
	if len(pubKeyBytes) != 65 {
		log.Errorf("public key length failed, should be 64 instead of %v", len(pubKeyBytes))
		return nil, nil, errors.New("public key length failed")
	}

	pri := PrivateKey{
		Algorithm: ECDSA,
		PrivateKey: &ecdsa.PrivateKey{
			D: new(big.Int).SetBytes(priKeyBytes),
			PublicKey: ecdsa.PublicKey{
				X:     new(big.Int).SetBytes(pubKeyBytes[1:33]),
				Y:     new(big.Int).SetBytes(pubKeyBytes[33:]),
				Curve: c,
			},
		},
	}

	pub := PublicKey{
		Algorithm: ECDSA,
		PublicKey: &pri.PublicKey,
	}
	return &pri, &pub, nil
}
