package keystore

import (
	"crypto"
	"crypto/ecdsa"
)

const (
	ECDSA ECAlgorithm = iota
	SM2
)

type ECAlgorithm byte

type PrivateKey struct {
	Algorithm ECAlgorithm
	*ecdsa.PrivateKey
}

func (this *PrivateKey) Public() crypto.PublicKey {
	return &PublicKey{Algorithm: this.Algorithm, PublicKey: &this.PublicKey}
}

type PublicKey struct {
	Algorithm ECAlgorithm
	*ecdsa.PublicKey
}

// ProtectedKey 存储编码后的私钥和相关的
type ProtectedKey struct {
	Address string            `json:"address"`
	EncAlg  string            `json:"enc-alg"`
	Key     []byte            `json:"key"`
	Alg     string            `json:"algorithm"`
	Salt    []byte            `json:"salt,omitempty"`
	Hash    string            `json:"hash,omitempty"`
	Param   map[string]string `json:"parameters,omitempty"`
}
