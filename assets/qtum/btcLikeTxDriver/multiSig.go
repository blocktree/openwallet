package btcLikeTxDriver

import (
	"encoding/hex"
	"errors"

	"github.com/blocktree/go-OWCrypt"
)

func CreateMultiSig(required byte, pubkeys [][]byte) (string, string, error) {
	if required < 1 {
		return "", "", errors.New("A multisignature address must require at least one key to redeem!")
	}
	if required > byte(len(pubkeys)) {
		return "", "", errors.New("Not enough keys supplied for a multisignature address to redeem!")
	}
	if len(pubkeys) > 16 {
		return "", "", errors.New("Number of keys involved in the multisignature address creation is too big!")
	}

	redeem := []byte{}

	redeem = append(redeem, OpCode_1+required-1)

	for _, k := range pubkeys {
		if len(k) != 33 && len(k) != 65 {
			return "", "", errors.New("Invalid pubkey data for multisignature address!")
		}
		redeem = append(redeem, byte(len(k)))
		redeem = append(redeem, k...)
	}

	redeem = append(redeem, OpCode_1+byte(len(pubkeys))-1)

	redeem = append(redeem, OpCheckMultiSig)

	if len(redeem) > MaxScriptElementSize {
		return "", "", errors.New("Redeem script exceeds size limit!")
	}

	redeemHash := owcrypt.Hash(redeem, 0, owcrypt.HASH_ALG_SHA256)
	redeemHash = append([]byte{0x00, 0x20}, redeemHash...)
	redeemHash = owcrypt.Hash(redeemHash, 0, owcrypt.HASH_ALG_HASH160)

	return EncodeCheck(P2SHPrefix, redeemHash), hex.EncodeToString(redeem), nil
}

func (t Transaction) isMultiSig() bool {
	if len(t.Vins) != 1 {
		return false
	}
	if t.Vins[0].ScriptPubkeySignature == nil {
		return false
	}
	if t.Vins[0].ScriptPubkeySignature[len(t.Vins[0].ScriptPubkeySignature)-1] != OpCheckMultiSig {
		return false
	}
	return true
}

func (t Contract) isMultiSig() bool {
	if len(t.Vins) != 1 {
		return false
	}
	if t.Vins[0].ScriptPubkeySignature == nil {
		return false
	}
	if t.Vins[0].ScriptPubkeySignature[len(t.Vins[0].ScriptPubkeySignature)-1] != OpCheckMultiSig {
		return false
	}
	return true
}

func isMultiSig(lockScript, redeemScript string) bool {
	if len(lockScript) != 0x17*2 {
		return false
	}

	lockBytes, err := hex.DecodeString(lockScript)
	if err != nil {
		return false
	}
	if !(lockBytes[0] == OpCodeHash160 && lockBytes[1] == 0x14 || lockBytes[22] == OpCodeEqual) {
		return false
	}

	redeemBytes, err := hex.DecodeString(redeemScript)
	if err != nil {
		return false
	}

	if len(redeemBytes) == 0 || redeemBytes[len(redeemBytes)-1] != OpCheckMultiSig {
		return false
	}
	return true

}

func calcRedeemHash(redeem []byte) []byte {
	redeemHash := owcrypt.Hash(redeem, 0, owcrypt.HASH_ALG_SHA256)

	return append([]byte{0x22, 0x00, 0x20}, redeemHash...)
}
