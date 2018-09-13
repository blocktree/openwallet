package btcLikeTxDriver

import "errors"

type TxWitness struct {
	Signature []byte
	Pubkey    []byte
}

func (w TxWitness) encodeToScript(sigType byte) []byte {
	r := w.Signature[:32]
	s := w.Signature[32:]

	if r[0]&0x80 == 0x80 {
		r = append([]byte{0x00}, r...)
	}
	if s[0]&0x80 == 0x80 {
		s = append([]byte{0}, s...)
	}

	r = append([]byte{byte(len(r))}, r...)
	r = append([]byte{0x02}, r...)
	s = append([]byte{byte(len(s))}, s...)
	s = append([]byte{0x02}, s...)

	rs := append(r, s...)
	rs = append([]byte{byte(len(rs))}, rs...)
	rs = append(rs, sigType)
	rs = append([]byte{0x30}, rs...)
	rs = append([]byte{byte(len(rs))}, rs...)

	pub := append([]byte{byte(len(w.Pubkey))}, w.Pubkey...)

	return append(rs, pub...)
}

func decodeFromSegwitBytes(script []byte) (*TxWitness, error) {
	var ret TxWitness
	index := 0
	sigLen := script[index]
	index++

	if script[index] != 0x30 {
		return nil, errors.New("Invalid signature data!")
	}
	index++

	rsLen := script[index]
	index++

	if script[index] != 0x02 {
		return nil, errors.New("Invalid signature data!")
	}
	index++

	rLen := script[index]
	index++

	if rLen == 0x21 {
		if script[index] != 0x00 && (script[index+1]&0x80 != 0x80) {
			return nil, errors.New("Invalid signature data!")
		}
		index++
	}

	ret.Signature = script[index : index+32]
	index += 32

	if script[index] != 0x02 {
		return nil, errors.New("Invalid signature data!")
	}
	index++

	sLen := script[index]
	index++

	if sLen == 0x21 {
		if script[index] != 0x00 && (script[index+1]&0x80 != 0x80) {
			return nil, errors.New("Invalid signature data!")
		}
		index++
	}

	ret.Signature = append(ret.Signature, script[index:index+32]...)
	index += 32

	if script[index] != SigHashAll {
		return nil, errors.New("Only sigAll supported!")
	}
	index++

	pubLen := script[index]
	index++
	if pubLen != 0x21 {
		return nil, errors.New("Only compressed pubkey is supported!")
	}

	ret.Pubkey = script[index : index+33]
	index += 33

	if (rLen+sLen+4 != rsLen) || (rsLen+3 != sigLen) || (sigLen+pubLen+2 != byte(len(script))) {
		return nil, errors.New("Invalid transaction data!")
	}

	if index != len(script) {
		return nil, errors.New("Invalid transaction data!")
	}
	return &ret, nil
}
