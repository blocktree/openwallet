package btcLikeTxDriver

import (
	"encoding/hex"
	"errors"

	"github.com/blocktree/go-OWCrypt"
)

const (
	TypeP2PKH  = 0
	TypeP2SH   = 1
	TypeBech32 = 2
)

type Transaction struct {
	Version  []byte
	Vins     []TxIn
	Vouts    []TxOut
	Witness  []TxWitness
	LockTime []byte
	//	HashType []byte
}

type Contract struct {
	Version   []byte
	Vins      []TxIn
	Vcontract TxContract
	Vouts     []TxOut
	Witness   []TxWitness
	LockTime  []byte
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
	if len(t.Vins) == 0 {
		return nil, errors.New("No input found in the transaction struct!")
	}

	if len(t.Vouts) == 0 {
		return nil, errors.New("No output found in the transaction struct!")
	}

	ret := []byte{}
	ret = append(ret, t.Version...)
	if t.isMultiSig() {
		if t.Witness == nil {
			return nil, errors.New("No witness data for a multisig transaction!")
		}
		ret = append(ret, SegWitSymbol, SegWitVersion)
		ret = append(ret, byte(len(t.Vins)))
		for _, in := range t.Vins {
			if in.TxID == nil || len(in.TxID) != 32 || in.Vout == nil || len(in.Vout) != 4 {
				return nil, errors.New("Invalid transaction input!")
			}
			ret = append(ret, in.TxID...)
			ret = append(ret, in.Vout...)
			redeemHash := calcRedeemHash(in.ScriptPubkeySignature)
			ret = append(ret, byte(len(redeemHash)))
			ret = append(ret, redeemHash...)
			ret = append(ret, in.Sequence...)
		}
		ret = append(ret, byte(len(t.Vouts)))

		for _, out := range t.Vouts {
			if out.amount == nil || len(out.amount) != 8 || out.lockScript == nil {
				return nil, errors.New("Invalid transaction output!")
			}
			ret = append(ret, out.amount...)
			ret = append(ret, byte(len(out.lockScript)))
			ret = append(ret, out.lockScript...)
		}

		ret = append(ret, byte(0x04), 0x00)
		for _, w := range t.Witness {
			if w.Signature == nil {
				return nil, errors.New("Miss signature data for a multisig transaction!")
			} else {
				sig := w.encodeToScript(SigHashAll)
				sig = sig[:len(sig)-34]
				ret = append(ret, sig...)
			}
		}

		ret = append(ret, byte(len(t.Vins[0].ScriptPubkeySignature)))

		ret = append(ret, t.Vins[0].ScriptPubkeySignature...)
	} else {

		if t.Witness != nil {
			ret = append(ret, SegWitSymbol, SegWitVersion)
		}

		ret = append(ret, byte(len(t.Vins)))

		for _, in := range t.Vins {
			if in.TxID == nil || len(in.TxID) != 32 || in.Vout == nil || len(in.Vout) != 4 {
				return nil, errors.New("Invalid transaction input!")
			}
			ret = append(ret, in.TxID...)
			ret = append(ret, in.Vout...)
			if in.ScriptPubkeySignature == nil {
				ret = append(ret, 0x00)
			} else {
				ret = append(ret, byte(len(in.ScriptPubkeySignature)))
				ret = append(ret, in.ScriptPubkeySignature...)
			}
			ret = append(ret, in.Sequence...)
		}

		ret = append(ret, byte(len(t.Vouts)))

		for _, out := range t.Vouts {
			if out.amount == nil || len(out.amount) != 8 || out.lockScript == nil {
				return nil, errors.New("Invalid transaction output!")
			}
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
	}

	ret = append(ret, t.LockTime...)
	return ret, nil
}

func newQRC20TokenTransaction(vins []Vin, vcontract Vcontract, vout []Vout, lockTime uint32, replaceable bool) (*Contract, error) {
	txIn, err := newTxInForEmptyTrans(vins)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(txIn); i++ {
		txIn[i].setSequence(lockTime, replaceable)
	}

	txContract, err := newTxContractForEmptyTrans(vcontract)
	if err != nil {
		return nil, err
	}

	txOut, err := newTxOutForEmptyTrans(vout)
	if err != nil {
		return nil, err
	}

	version := uint32ToLittleEndianBytes(DefaultTxVersion)
	locktime := uint32ToLittleEndianBytes(lockTime)

	return &Contract{version, txIn, *txContract, txOut, nil, locktime}, nil
}

func (t Contract) encodeToBytes() ([]byte, error) {
	if len(t.Vins) == 0 {
		return nil, errors.New("No input found in the transaction struct!")
	}

	if len(t.Vouts) == 0 {
		return nil, errors.New("No output found in the transaction struct!")
	}

	ret := []byte{}
	ret = append(ret, t.Version...)
	if t.isMultiSig() {
		if t.Witness == nil {
			return nil, errors.New("No witness data for a multisig transaction!")
		}
		ret = append(ret, SegWitSymbol, SegWitVersion)
		ret = append(ret, byte(len(t.Vins)))
		for _, in := range t.Vins {
			if in.TxID == nil || len(in.TxID) != 32 || in.Vout == nil || len(in.Vout) != 4 {
				return nil, errors.New("Invalid transaction input!")
			}
			ret = append(ret, in.TxID...)
			ret = append(ret, in.Vout...)
			redeemHash := calcRedeemHash(in.ScriptPubkeySignature)
			ret = append(ret, byte(len(redeemHash)))
			ret = append(ret, redeemHash...)
			ret = append(ret, in.Sequence...)
		}

		//contract
		ret = append(ret, byte(0x02), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x63)
		ret = append(ret, t.Vcontract.vmVersion...)
		ret = append(ret, t.Vcontract.lenGasLimit...)
		ret = append(ret, t.Vcontract.gasLimit...)
		ret = append(ret, t.Vcontract.lenGasPrice...)
		ret = append(ret, t.Vcontract.gasPrice...)
		ret = append(ret,0x44)
		ret = append(ret, t.Vcontract.dataHex...)
		ret = append(ret,t.Vcontract.lenContract...)
		ret = append(ret, t.Vcontract.contractAddr...)
		ret = append(ret, t.Vcontract.opCall...)

		for _, out := range t.Vouts {
			if out.amount == nil || len(out.amount) != 8 || out.lockScript == nil {
				return nil, errors.New("Invalid transaction output!")
			}
			ret = append(ret, out.amount...)
			ret = append(ret, byte(len(out.lockScript)))
			ret = append(ret, out.lockScript...)
		}

		ret = append(ret, byte(0x04), 0x00)
		for _, w := range t.Witness {
			if w.Signature == nil {
				return nil, errors.New("Miss signature data for a multisig transaction!")
			} else {
				sig := w.encodeToScript(SigHashAll)
				sig = sig[:len(sig)-34]
				ret = append(ret, sig...)
			}
		}

		ret = append(ret, byte(len(t.Vins[0].ScriptPubkeySignature)))

		ret = append(ret, t.Vins[0].ScriptPubkeySignature...)
	} else {

		if t.Witness != nil {
			ret = append(ret, SegWitSymbol, SegWitVersion)
		}

		ret = append(ret, byte(len(t.Vins)))

		for _, in := range t.Vins {
			if in.TxID == nil || len(in.TxID) != 32 || in.Vout == nil || len(in.Vout) != 4 {
				return nil, errors.New("Invalid transaction input!")
			}
			ret = append(ret, in.TxID...)
			ret = append(ret, in.Vout...)
			if in.ScriptPubkeySignature == nil {
				ret = append(ret, 0x00)
			} else {
				ret = append(ret, byte(len(in.ScriptPubkeySignature)))
				ret = append(ret, in.ScriptPubkeySignature...)
			}
			ret = append(ret, in.Sequence...)
		}

		//contract
		ret = append(ret, byte(0x02), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x63)
		ret = append(ret, t.Vcontract.vmVersion...)
		ret = append(ret, t.Vcontract.lenGasLimit...)
		ret = append(ret, t.Vcontract.gasLimit...)
		ret = append(ret, t.Vcontract.lenGasPrice...)
		ret = append(ret, t.Vcontract.gasPrice...)
		ret = append(ret,0x44)
		ret = append(ret, t.Vcontract.dataHex...)
		ret = append(ret,t.Vcontract.lenContract...)
		ret = append(ret, t.Vcontract.contractAddr...)
		ret = append(ret, t.Vcontract.opCall...)

		for _, out := range t.Vouts {
			if out.amount == nil || len(out.amount) != 8 || out.lockScript == nil {
				return nil, errors.New("Invalid transaction output!")
			}
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
	}

	ret = append(ret, t.LockTime...)
	return ret, nil
}

func DecodeRawTransaction(txBytes []byte) (*Transaction, error) {
	limit := len(txBytes)
	if limit == 0 {
		return nil, errors.New("Invalid transaction data length!")
	}
	var rawTx Transaction
	index := int(0)
	segwit := false

	if index+4 > limit {
		return nil, errors.New("Invalid transaction data length!")
	}
	rawTx.Version = txBytes[index : index+4]
	index += 4

	if index+2 > limit {
		return nil, errors.New("Invalid transaction data length!")
	}
	if txBytes[index] == SegWitSymbol {
		if txBytes[index+1] != SegWitVersion {
			return nil, errors.New("Invalid witness symbol!")
		}
		segwit = true
		index += 2
	}

	if index+1 > limit {
		return nil, errors.New("Invalid transaction data length!")
	}
	numOfVins := txBytes[index]
	index++

	for i := byte(0); i < numOfVins; i++ {
		var tmpTxIn TxIn

		if index+32 > limit {
			return nil, errors.New("Invalid transaction data length!")
		}
		tmpTxIn.TxID = txBytes[index : index+32]
		index += 32

		if index+4 > limit {
			return nil, errors.New("Invalid transaction data length!")
		}
		tmpTxIn.Vout = txBytes[index : index+4]
		index += 4

		if index+1 > limit {
			return nil, errors.New("Invalid transaction data length!")
		}
		scriptLen := txBytes[index]
		index++
		if scriptLen == 0 {
			tmpTxIn.ScriptPubkeySignature = nil
		} else {
			if index+int(scriptLen) > limit {
				return nil, errors.New("Invalid transaction data length!")
			}
			tmpTxIn.ScriptPubkeySignature = txBytes[index : index+int(scriptLen)]
		}
		index += int(scriptLen)

		if index+4 > limit {
			return nil, errors.New("Invalid transaction data length!")
		}
		tmpTxIn.Sequence = txBytes[index : index+4]
		index += 4

		rawTx.Vins = append(rawTx.Vins, tmpTxIn)
	}

	if index+1 > limit {
		return nil, errors.New("Invalid transaction data length!")
	}
	numOfVouts := txBytes[index]
	index++

	for i := byte(0); i < numOfVouts; i++ {
		var tmpTxOut TxOut

		if index+8 > limit {
			return nil, errors.New("Invalid transaction data length!")
		}
		tmpTxOut.amount = txBytes[index : index+8]
		index += 8

		if index+1 > limit {
			return nil, errors.New("Invalid transaction data length!")
		}
		lockScriptLen := txBytes[index]
		index++

		if index+int(lockScriptLen) > limit {
			return nil, errors.New("Invalid transaction data length!")
		}
		tmpTxOut.lockScript = txBytes[index : index+int(lockScriptLen)]
		index += int(lockScriptLen)

		rawTx.Vouts = append(rawTx.Vouts, tmpTxOut)
	}

	if segwit {
		for i := byte(0); i < numOfVins; i++ {
			if index+1 > limit {
				return nil, errors.New("Invalid transaction data length!")
			}
			if txBytes[index] == 0x00 {
				rawTx.Witness = append(rawTx.Witness, TxWitness{})
				index++
			} else if txBytes[index] == 0x02 {
				index++

				if index+1 > limit {
					return nil, errors.New("Invalid transaction data length!")
				}
				length := txBytes[index]

				if index+int(length)+1+34 > limit {
					return nil, errors.New("Invalid transaction data length!")
				}
				witness, _ := decodeFromSegwitBytes(txBytes[index : index+int(length)+1+34])
				rawTx.Witness = append(rawTx.Witness, *witness)

				index += int(length) + 1 + 34
			} else {
				//TODO
				//for multisig
			}
		}
	}

	if index+4 > limit {
		return nil, errors.New("Invalid transaction data length!")
	}
	rawTx.LockTime = txBytes[index : index+4]
	index += 4

	if index != limit {
		return nil, errors.New("Too much transaction data!")
	}
	return &rawTx, nil
}

func isScriptHash(script []byte) bool {
	if script[0] == OpCodeDup && script[1] == OpCodeHash160 && script[2] == 0x14 && script[23] == OpCodeEqualVerify && script[24] == OpCodeCheckSig {
		return false
	}
	return true
}

func checkScriptType(script []byte) int {
	if script[0] == OpCodeDup && script[1] == OpCodeHash160 && script[2] == 0x14 && script[23] == OpCodeEqualVerify && script[24] == OpCodeCheckSig {
		return TypeP2PKH
	} else if script[0] == OpCodeHash160 && script[1] == 0x14 && script[22] == OpCodeEqual {
		return TypeP2SH
	} else if script[0] == 0x00 && script[1] == 0x14 {
		return TypeBech32
	} else {
		return -1
	}
}

func calcSegwitHash(tx Transaction) ([]byte, []byte, []byte, error) {
	hashPrevouts := []byte{}
	hashSequence := []byte{}
	hashOutputs := []byte{}

	for _, vin := range tx.Vins {
		hashPrevouts = append(hashPrevouts, vin.TxID...)
		hashPrevouts = append(hashPrevouts, vin.Vout...)

		hashSequence = append(hashSequence, vin.Sequence...)
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

		if len(ret) != 0x14 {
			return nil, errors.New("Invalid redeem script!")
		}
		ret = append([]byte{byte(len(ret))}, ret...)
		ret = append([]byte{OpCodeDup, OpCodeHash160}, ret...)
		ret = append(ret, OpCodeEqualVerify, OpCodeCheckSig)
	} else {
		//TODO
		//for multi sig
		ret = redeemBytes
	}

	return ret, nil
}

func (tx Transaction) calcSegwitBytesForSig(unlockData TxUnlock, txid, vout, sequence []byte) ([]byte, error) {
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
	if t.Vins == nil || len(t.Vins) != len(unlockData) {
		return nil, errors.New("The number of Keys and UTXOs are not match!")
	}

	for i := 0; i < len(unlockData); i++ {
		sigBytes := []byte{}
		for j := 0; j < len(unlockData); j++ {
			t.Vins[j].ScriptPubkeySignature = nil
		}
		lockBytes, err := hex.DecodeString(unlockData[i].LockScript)
		if err != nil {
			return nil, errors.New("Invalid lockscript!")
		}

		if lockBytes == nil || len(lockBytes) == 0 || (len(lockBytes) != 22 && len(lockBytes) != 23 && len(lockBytes) != 25) {
			return nil, errors.New("Check the lockscript data!")
		}

		scriptType := checkScriptType(lockBytes)
		if scriptType == TypeP2SH || scriptType == TypeBech32 {
			if scriptType == TypeBech32 {
				unlockData[i].RedeemScript = unlockData[i].LockScript
			}
			sigBytes, err = t.calcSegwitBytesForSig(unlockData[i], t.Vins[i].TxID, t.Vins[i].Vout, t.Vins[i].Sequence)
			if err != nil {
				return nil, err
			}
		} else if scriptType == TypeP2PKH {
			t.Vins[i].ScriptPubkeySignature = lockBytes

			sigBytes, err = t.encodeToBytes()
			if err != nil {
				return nil, err
			}

		} else {
			return nil, errors.New("Unknown type of lockscript!")
		}

		sigBytes = append(sigBytes, uint32ToLittleEndianBytes(DefaultHashType)...)

		hash := owcrypt.Hash(sigBytes, 0, owcrypt.HASh_ALG_DOUBLE_SHA256)

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
