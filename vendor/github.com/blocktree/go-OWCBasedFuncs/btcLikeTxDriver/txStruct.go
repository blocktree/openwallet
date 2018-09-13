package btcLikeTxDriver

import (
	"encoding/hex"
	"errors"

	"github.com/blocktree/go-OWCrypt"
)

type Transaction struct {
	Version  []byte
	Vins     []TxIn
	Vouts    []TxOut
	Witness  []TxWitness
	LockTime []byte
	//	HashType []byte
}

func newTransaction(vins []Vin, vouts []Vout, lockTime uint32, replaceable bool) (*Transaction, error) {
	txIn, err := newTxInForEmptyTrans(vins)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(txIn); i++ {
		txIn[i].setSequence(lockTime, replaceable)
	}

	txOut, err := newTxOutForEmptyTrans(vouts)
	if err != nil {
		return nil, err
	}

	version := uint32ToLittleEndianBytes(DefaultTxVersion)
	locktime := uint32ToLittleEndianBytes(lockTime)

	return &Transaction{version, txIn, txOut, nil, locktime}, nil
}

func (t Transaction) encodeToBytes() ([]byte, error) {
	ret := []byte{}

	ret = append(ret, t.Version...)
	if t.Witness != nil {
		ret = append(ret, SegWitSymbol, SegWitVersion)
	}

	ret = append(ret, byte(len(t.Vins)))

	for _, in := range t.Vins {
		ret = append(ret, in.txID...)
		ret = append(ret, in.vout...)
		if in.scriptPubkeySignature == nil {
			ret = append(ret, 0x00)
		} else {
			ret = append(ret, byte(len(in.scriptPubkeySignature)))
			ret = append(ret, in.scriptPubkeySignature...)
		}
		ret = append(ret, in.sequence...)
	}

	ret = append(ret, byte(len(t.Vouts)))

	for _, out := range t.Vouts {
		ret = append(ret, out.amount...)
		ret = append(ret, byte(len(out.lockScript)))
		ret = append(ret, out.lockScript...)
	}

	if t.Witness != nil {
		for _, w := range t.Witness {
			if w.Signature == nil {
				ret = append(ret, byte(0x00))
			} else {
				ret = append(ret, byte(0x02))
				ret = append(ret, w.encodeToScript(SigHashAll)...)
			}
		}
	}
	ret = append(ret, t.LockTime...)
	return ret, nil
}

func decodeFromBytes(txBytes []byte) Transaction {
	var rawTx Transaction
	index := int(0)
	segwit := false

	rawTx.Version = txBytes[index : index+4]
	index += 4

	if txBytes[index] == SegWitSymbol {
		if txBytes[index+1] != SegWitVersion {
			//errors.New("Wintess version invalid!")
		}
		segwit = true
		index += 2
	}

	numOfVins := txBytes[index]
	index++

	for i := byte(0); i < numOfVins; i++ {
		var tmpTxIn TxIn
		tmpTxIn.txID = txBytes[index : index+32]
		index += 32
		tmpTxIn.vout = txBytes[index : index+4]
		index += 4
		scriptLen := txBytes[index]
		index++
		if scriptLen == 0 {
			tmpTxIn.scriptPubkeySignature = nil
		} else {
			tmpTxIn.scriptPubkeySignature = txBytes[index : index+int(scriptLen)]
		}
		index += int(scriptLen)

		tmpTxIn.sequence = txBytes[index : index+4]
		index += 4

		rawTx.Vins = append(rawTx.Vins, tmpTxIn)
	}

	numOfVouts := txBytes[index]
	index++

	for i := byte(0); i < numOfVouts; i++ {
		var tmpTxOut TxOut
		tmpTxOut.amount = txBytes[index : index+8]
		index += 8
		lockScriptLen := txBytes[index]
		index++
		tmpTxOut.lockScript = txBytes[index : index+int(lockScriptLen)]
		index += int(lockScriptLen)

		rawTx.Vouts = append(rawTx.Vouts, tmpTxOut)
	}

	if segwit {
		for i := byte(0); i < numOfVins; i++ {
			if txBytes[index] == 0x00 {
				rawTx.Witness = append(rawTx.Witness, TxWitness{})
				index++
			} else if txBytes[index] == 0x02 {
				index++
				length := txBytes[index]
				witness, _ := decodeFromSegwitBytes(txBytes[index : index+int(length)+1+34])
				rawTx.Witness = append(rawTx.Witness, *witness)

				index += int(length) + 1 + 34
			} else {
				//TODO
				//for multisig
			}
		}
	}

	rawTx.LockTime = txBytes[index : index+4]
	index += 4

	return rawTx
}

func isScriptHash(script []byte) bool {
	if script[0] == OpCodeDup && script[1] == OpCodeHash160 && script[2] == 0x14 && script[23] == OpCodeEqualVerify && script[24] == OpCodeCheckSig {
		return false
	}
	return true
}

func calcSegwitHash(tx Transaction) ([]byte, []byte, []byte, error) {
	hashPrevouts := []byte{}
	hashSequence := []byte{}
	hashOutputs := []byte{}

	for _, vin := range tx.Vins {
		hashPrevouts = append(hashPrevouts, vin.txID...)
		hashPrevouts = append(hashPrevouts, vin.vout...)

		hashSequence = append(hashSequence, vin.sequence...)
	}

	for _, vout := range tx.Vouts {
		hashOutputs = append(hashOutputs, vout.amount...)
		hashOutputs = append(hashOutputs, byte(len(vout.lockScript)))
		hashOutputs = append(hashOutputs, vout.lockScript...)
	}
	return owcrypt.Hash(hashPrevouts, 0, owcrypt.HASh_ALG_DOUBLE_SHA256),
		owcrypt.Hash(hashSequence, 0, owcrypt.HASh_ALG_DOUBLE_SHA256),
		owcrypt.Hash(hashOutputs, 0, owcrypt.HASh_ALG_DOUBLE_SHA256),
		nil
}

func genScriptCodeFromRedeemScript(redeem string) ([]byte, error) {
	redeemBytes, err := hex.DecodeString(redeem)
	if err != nil {
		return nil, errors.New("Invalid redeem script!")
	}

	ret := []byte{}
	if redeemBytes[0] == 0x00 && redeemBytes[1] == 0x14 {
		ret = redeemBytes[2:]
		ret = append([]byte{byte(len(ret))}, ret...)
		ret = append([]byte{OpCodeDup, OpCodeHash160}, ret...)
		ret = append(ret, OpCodeEqualVerify, OpCodeCheckSig)
	} else {
		//TODO
		//for multi sig
	}

	return ret, nil
}

func (tx Transaction) calcSegwitHashForSig(unlockData TxUnlock, txid, vout, sequence []byte) ([]byte, error) {
	sigBytes := []byte{}

	sigBytes = append(sigBytes, tx.Version...)
	hashPrevouts, hashSequence, hashOutputs, err := calcSegwitHash(tx)
	if err != nil {
		return nil, err
	}
	sigBytes = append(sigBytes, hashPrevouts...)
	sigBytes = append(sigBytes, hashSequence...)

	sigBytes = append(sigBytes, txid...)
	sigBytes = append(sigBytes, vout...)

	scriptCode, err := genScriptCodeFromRedeemScript(unlockData.RedeemScript)
	if err != nil {
		return nil, err
	}

	sigBytes = append(sigBytes, byte(len(scriptCode)))
	sigBytes = append(sigBytes, scriptCode...)

	sigBytes = append(sigBytes, uint64ToLittleEndianBytes(unlockData.Amount)...)
	sigBytes = append(sigBytes, sequence...)

	sigBytes = append(sigBytes, hashOutputs...)
	sigBytes = append(sigBytes, tx.LockTime...)

	return sigBytes, nil
}

func (t Transaction) getHashesForSig(unlockData []TxUnlock) ([][]byte, error) {
	hashes := [][]byte{}
	if len(t.Vins) != len(unlockData) {
		return nil, errors.New("The number of Keys and UTXOs are not match!")
	}

	for i := 0; i < len(unlockData); i++ {
		sigBytes := []byte{}
		for j := 0; j < len(unlockData); j++ {
			t.Vins[j].scriptPubkeySignature = nil
		}
		lockBytes, err := hex.DecodeString(unlockData[i].LockScript)
		if err != nil {
			return nil, errors.New("Invalid lockscript!")
		}
		if isScriptHash(lockBytes) {
			sigBytes, err = t.calcSegwitHashForSig(unlockData[i], t.Vins[i].txID, t.Vins[i].vout, t.Vins[i].sequence)
			if err != nil {
				return nil, err
			}

		} else {
			t.Vins[i].scriptPubkeySignature = lockBytes

			sigBytes, err = t.encodeToBytes()
			if err != nil {
				return nil, err
			}

		}
		sigBytes = append(sigBytes, uint32ToLittleEndianBytes(DefaultHashType)...)

		hash := owcrypt.Hash(sigBytes, 0, owcrypt.HASH_ALG_SHA256)

		hashes = append(hashes, hash)
	}

	return hashes, nil
}

func verifyHashes(hashes [][]byte, sigPub []SignaturePubkey) bool {

	for i := 0; i < len(sigPub); i++ {
		pubkey := owcrypt.PointDecompress(sigPub[i].Pubkey, owcrypt.ECC_CURVE_SECP256K1)[1:]
		if owcrypt.Verify(pubkey, nil, 0, hashes[i], 32, sigPub[i].Signature, owcrypt.ECC_CURVE_SECP256K1) != owcrypt.SUCCESS {
			return false
		}
	}
	return true
}
