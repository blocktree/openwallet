package btcLikeTxDriver

import (
	"encoding/hex"
	"errors"
	"strings"
)

type Vin struct {
	TxID string
	Vout uint32
}

type Vout struct {
	Address string
	Amount  uint64
}

type TxUnlock struct {
	PrivateKey   []byte
	LockScript   string
	RedeemScript string
	Amount       uint64
	Address      string
}

const (
	DefaultTxVersion = uint32(2)
	DefaultHashType  = uint32(1)
)

func CreateEmptyRawTransaction(vins []Vin, vouts []Vout, lockTime uint32, replaceable bool) (string, error) {
	emptyTrans, err := newTransaction(vins, vouts, lockTime, replaceable)
	if err != nil {
		return "", err
	}

	txBytes, err := emptyTrans.encodeToBytes()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(txBytes), nil
}

func CreateRawTransactionHashForSig(txHex string, unlockData []TxUnlock) ([]string, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, errors.New("Invalid transaction hex string!")
	}
	emptyTrans, err := DecodeRawTransaction(txBytes)
	if err != nil {
		return nil, err
	}

	hashes, err := emptyTrans.getHashesForSig(unlockData)
	if err != nil {
		return nil, err
	}

	ret := []string{}

	for _, h := range hashes {
		ret = append(ret, hex.EncodeToString(h))
	}

	return ret, nil
}

func SignEmptyRawTransaction(txHex string, unlockData []TxUnlock) (string, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return "", errors.New("Invalid transaction hex string!")
	}
	emptyTrans, err := DecodeRawTransaction(txBytes)
	if err != nil {
		return "", err
	}

	hashes, err := emptyTrans.getHashesForSig(unlockData)
	if err != nil {
		return "", err
	}

	sigPub, err := calcSignaturePubkey(hashes, unlockData)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(sigPub); i++ {
		lockBytes, err := hex.DecodeString(unlockData[i].LockScript)
		if err != nil {
			return "", errors.New("Invalid lock script or redeem script!")
		}

		scriptType := checkScriptType(lockBytes)
		//if isScriptHash(lockBytes) {
		if scriptType == TypeP2SH || scriptType == TypeBech32 {

			if scriptType == TypeBech32 {
				emptyTrans.Vins[i].ScriptPubkeySignature = nil
			} else {
				redeemScript, err := hex.DecodeString(unlockData[i].RedeemScript)
				if err != nil {
					return "", errors.New("Invalid redeem script!")
				}
				redeemScript = append([]byte{byte(len(redeemScript))}, redeemScript...)
				emptyTrans.Vins[i].ScriptPubkeySignature = redeemScript
			}

			if emptyTrans.Witness == nil {
				for j := 0; j < i; j++ {
					emptyTrans.Witness = append(emptyTrans.Witness, TxWitness{})
				}
				emptyTrans.Witness = append(emptyTrans.Witness, TxWitness{sigPub[i].Signature, sigPub[i].Pubkey})
			}

		} else {
			emptyTrans.Vins[i].ScriptPubkeySignature = sigPub[i].encodeToScript(SigHashAll)
			if emptyTrans.Witness != nil {
				emptyTrans.Witness = append(emptyTrans.Witness, TxWitness{})
			}
		}
	}

	txBytes, err = emptyTrans.encodeToBytes()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(txBytes), nil
}

func SignRawTransactionHash(txHash []string, unlockData []TxUnlock) ([]SignaturePubkey, error) {
	hashes := [][]byte{}
	for _, h := range txHash {
		hash, err := hex.DecodeString(h)
		if err != nil {
			return nil, errors.New("Invalid transaction hash data!")
		}
		hashes = append(hashes, hash)
	}

	return calcSignaturePubkey(hashes, unlockData)
}

func InsertSignatureIntoEmptyTransaction(txHex string, sigPub []SignaturePubkey, unlockData []TxUnlock) (string, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return "", errors.New("Invalid transaction hex data!")
	}

	emptyTrans, err := DecodeRawTransaction(txBytes)
	if err != nil {
		return "", err
	}

	if len(emptyTrans.Vins) != len(unlockData) {
		return "", errors.New("The number of transaction inputs and the unlock data are not match!")
	}

	if isMultiSig(unlockData[0].LockScript, unlockData[0].RedeemScript) {
		redeemBytes, _ := hex.DecodeString(unlockData[0].RedeemScript)
		emptyTrans.Vins[0].ScriptPubkeySignature = redeemBytes
		for i := 0; i < len(sigPub); i++ {
			emptyTrans.Witness = append(emptyTrans.Witness, TxWitness{sigPub[i].Signature, sigPub[i].Pubkey})
		}
	} else {
		for i := 0; i < len(emptyTrans.Vins); i++ {

			if sigPub[i].Signature == nil || len(sigPub[i].Signature) != 64 {
				return "", errors.New("Invalid signature data!")
			}
			if sigPub[i].Pubkey == nil || len(sigPub[i].Pubkey) != 33 {
				return "", errors.New("Invalid pubkey data!")
			}

			// bech32 branch
			if unlockData[i].RedeemScript == "" && strings.Index(unlockData[i].LockScript, "0014") == 0 {
				unlockData[i].RedeemScript = unlockData[i].LockScript
				unlockData[i].LockScript = "00"
			}

			if unlockData[i].RedeemScript == "" {

				emptyTrans.Vins[i].ScriptPubkeySignature = sigPub[i].encodeToScript(SigHashAll)
				if emptyTrans.Witness != nil {
					emptyTrans.Witness = append(emptyTrans.Witness, TxWitness{})
				}
			} else {
				if emptyTrans.Witness == nil {
					for j := 0; j < i; j++ {
						emptyTrans.Witness = append(emptyTrans.Witness, TxWitness{})
					}
				}
				emptyTrans.Witness = append(emptyTrans.Witness, TxWitness{sigPub[i].Signature, sigPub[i].Pubkey})
				if unlockData[i].RedeemScript == "" {
					return "", errors.New("Missing redeem script for a P2SH input!")
				}

				if unlockData[i].LockScript == "00" {
					emptyTrans.Vins[i].ScriptPubkeySignature = nil
				} else {
					redeem, err := hex.DecodeString(unlockData[i].RedeemScript)
					if err != nil {
						return "", errors.New("Invlalid redeem script!")
					}
					redeem = append([]byte{byte(len(redeem))}, redeem...)
					emptyTrans.Vins[i].ScriptPubkeySignature = redeem
				}
			}
		}
	}

	txBytes, err = emptyTrans.encodeToBytes()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(txBytes), nil
}

func VerifyRawTransaction(txHex string, unlockData []TxUnlock) bool {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return false
	}

	signedTrans, err := DecodeRawTransaction(txBytes)
	if err != nil {
		return false
	}

	if len(signedTrans.Vins) != len(unlockData) {
		return false
	}

	var sigAndPub []SignaturePubkey
	if signedTrans.Witness == nil {
		for _, sp := range signedTrans.Vins {
			tmp, err := decodeFromScriptBytes(sp.ScriptPubkeySignature)
			if err != nil {
				return false
			}
			sigAndPub = append(sigAndPub, *tmp)
		}
	} else {
		for i := 0; i < len(signedTrans.Vins); i++ {
			if signedTrans.Witness[i].Signature == nil {
				tmp, err := decodeFromScriptBytes(signedTrans.Vins[i].ScriptPubkeySignature)
				if err != nil {
					return false
				}
				sigAndPub = append(sigAndPub, *tmp)
			} else {
				sigAndPub = append(sigAndPub, SignaturePubkey{signedTrans.Witness[i].Signature, signedTrans.Witness[i].Pubkey})
				if strings.Index(unlockData[i].LockScript, "0014") == 0 {
					continue
				}
				unlockData[i].RedeemScript = hex.EncodeToString(signedTrans.Vins[i].ScriptPubkeySignature[1:])
			}
		}
	}

	signedTrans.Witness = nil

	hashes, err := signedTrans.getHashesForSig(unlockData)
	if err != nil {
		return false
	}

	return verifyHashes(hashes, sigAndPub)
}
