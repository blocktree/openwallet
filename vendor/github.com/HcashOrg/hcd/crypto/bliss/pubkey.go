package bliss

import (
	dcrcrypto "github.com/HcashOrg/hcd/crypto"
	"github.com/HcashOrg/bliss"
)

type PublicKey struct {
	dcrcrypto.PublicKeyAdapter
	bliss.PublicKey
}

func (p PublicKey) GetType() int {
	return pqcTypeBliss
}

func (p PublicKey) Serialize() []byte {
	return p.PublicKey.Serialize()
}

func (p PublicKey) SerializeCompressed() []byte {
	return p.Serialize()
}

func (p PublicKey) SerializeUnCompressed() []byte {
	return p.Serialize()
}
