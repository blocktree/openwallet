package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/blocktree/OpenWallet/log"
	base58 "github.com/itchyny/base58-go"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/scrypt"
	//	"github.com/ontio/ontology-crypto/ec"
	//	"golang.org/x/crypto/ed25519"
)

const (
	//P224 curve type
	P224 byte = 1
	//P256 curve type
	P256 byte = 2
	//P384 curve type
	P384 byte = 3
	//P521 curve type
	P521 byte = 4
	// SM2P256V1 curve label
	SM2P256V1 byte = 20
	// ED25519 curve label
	ED25519 byte = 25
)

const (
	PK_ECDSA KeyType = 0x12
	PK_SM2   KeyType = 0x13

//	PK_EDDSA KeyType = 0x14

//	PK_P256_E KeyType = 0x02
//	PK_P256_O KeyType = 0x03
)

//err_generate 错误title
const err_generate = "key pair generation failed, "

//KeyType key类型
type KeyType byte

// //PublicKey 公钥
// type PublicKey crypto.PublicKey

// //PrivateKey 私钥
// type PrivateKey interface {
// 	crypto.PrivateKey
// 	Public() crypto.PublicKey
// }

//GetCurve 获取曲线类型
func GetCurve(label byte) (elliptic.Curve, error) {
	switch label {
	case P224:
		return elliptic.P224(), nil
	case P256:
		return elliptic.P256(), nil
	case P384:
		return elliptic.P384(), nil
	case P521:
		return elliptic.P521(), nil
		//	case SM2P256V1:
		//		return sm2.SM2P256V1(), nil
	default:
		return nil, errors.New("unknown elliptic curve")
	}

}

const maxInt = int(^uint(0) >> 1)

var ErrTooLarge = errors.New("bytes.Buffer: too large")

type ZeroCopySink struct {
	buf []byte
}

func (self *ZeroCopySink) Bytes() []byte { return self.buf }

func (self *ZeroCopySink) WriteUint8(data uint8) {
	buf := self.NextBytes(1)
	buf[0] = data
}

func (self *ZeroCopySink) WriteByte(c byte) {
	self.WriteUint8(c)
}

func (self *ZeroCopySink) WriteUint32(data uint32) {
	buf := self.NextBytes(4)
	binary.LittleEndian.PutUint32(buf, data)
}

func (self *ZeroCopySink) NextBytes(n uint64) (data []byte) {
	m, ok := self.tryGrowByReslice(int(n))
	if !ok {
		m = self.grow(int(n))
	}
	data = self.buf[m:]
	return
}
func (self *ZeroCopySink) WriteUint16(data uint16) {
	buf := self.NextBytes(2)
	binary.LittleEndian.PutUint16(buf, data)
}

func (self *ZeroCopySink) WriteBytes(p []byte) {
	data := self.NextBytes(uint64(len(p)))
	copy(data, p)
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(bytes.ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

func (self *ZeroCopySink) grow(n int) int {
	// Try to grow by means of a reslice.
	if i, ok := self.tryGrowByReslice(n); ok {
		return i
	}

	l := len(self.buf)
	c := cap(self.buf)
	if c > maxInt-c-n {
		panic(ErrTooLarge)
	}
	// Not enough space anywhere, we need to allocate.
	buf := makeSlice(2*c + n)
	copy(buf, self.buf)
	self.buf = buf[:l+n]
	return l
}

func (self *ZeroCopySink) tryGrowByReslice(n int) (int, bool) {
	if l := len(self.buf); n <= cap(self.buf)-l {
		self.buf = self.buf[:l+n]
		return l, true
	}
	return 0, false
}

// ScryptParam contains the parameters used in scrypt function
type ScryptParam struct {
	P     int `json:"p"`
	N     int `json:"n"`
	R     int `json:"r"`
	DKLen int `json:"dkLen,omitempty"`
}

type ProgramBuilder struct {
	sink *ZeroCopySink
}

func (self *ProgramBuilder) PushOpCode(op OpCode) *ProgramBuilder {
	self.sink.WriteByte(byte(op))
	return self
}

type OpCode byte

const (
	// Constants
	PUSHBYTES1  OpCode = 0x01 // 0x01-0x4B The next opcode bytes is data to be pushed onto the stack
	PUSHBYTES75 OpCode = 0x4B
	PUSHDATA1   OpCode = 0x4C // The next byte contains the number of bytes to be pushed onto the stack.
	PUSHDATA2   OpCode = 0x4D // The next two bytes contain the number of bytes to be pushed onto the stack.
	PUSHDATA4   OpCode = 0x4E // The next four bytes contain the number of bytes to be pushed onto the stack.

	// Crypto
	CHECKSIG OpCode = 0xAC // The entire transaction's outputs inputs and script (from the most recently-executed CODESEPARATOR to the end) are hashed. The signature used by CHECKSIG must be a valid signature for this hash and public key. If it is 1 is returned 0 otherwise.
)

func (self *ProgramBuilder) PushBytes(data []byte) *ProgramBuilder {
	if len(data) == 0 {
		panic("push data error: data is nil")
	}

	if len(data) <= int(PUSHBYTES75)+1-int(PUSHBYTES1) {
		self.sink.WriteByte(byte(len(data)) + byte(PUSHBYTES1) - 1)
	} else if len(data) < 0x100 {
		self.sink.WriteByte(byte(PUSHDATA1))
		self.sink.WriteUint8(uint8(len(data)))
	} else if len(data) < 0x10000 {
		self.sink.WriteByte(byte(PUSHDATA2))
		self.sink.WriteUint16(uint16(len(data)))
	} else {
		self.sink.WriteByte(byte(PUSHDATA4))
		self.sink.WriteUint32(uint32(len(data)))
	}
	self.sink.WriteBytes(data)

	return self
}

const (
	compress_even = 2
	compress_odd  = 3
	nocompress    = 4
)

func EncodePublicKey(key *ecdsa.PublicKey, compressed bool) []byte {
	if key == nil {
		panic("invalid argument: public key is nil")
	}

	length := (key.Curve.Params().BitSize + 7) >> 3
	buf := make([]byte, (length*2)+1)
	x := key.X.Bytes()
	copy(buf[length+1-len(x):], x)
	if compressed {
		if key.Y.Bit(0) == 0 {
			buf[0] = compress_even
		} else {
			buf[0] = compress_odd
		}
		return buf[:length+1]
	} else {
		buf[0] = nocompress
		y := key.Y.Bytes()
		copy(buf[length*2+1-len(y):], y)
		return buf
	}
}

func SerializePublicKey(key *PublicKey) []byte {
	var buf bytes.Buffer
	// switch t := key.(type) {
	// case *ec.PublicKey:
	switch key.Algorithm {
	case ECDSA:
		// Take P-256 as a special case
		if key.Params().Name == elliptic.P256().Params().Name {
			return EncodePublicKey(key.PublicKey, true)
		}
		buf.WriteByte(byte(PK_ECDSA))
	case SM2:
		buf.WriteByte(byte(PK_SM2))
	}
	label, err := GetCurveLabel(key.Curve)
	if err != nil {
		panic(err)
	}
	buf.WriteByte(label)
	buf.Write(EncodePublicKey(key.PublicKey, true))
	// case ed25519.PublicKey:
	// 	buf.WriteByte(byte(PK_EDDSA))
	// 	buf.WriteByte(ED25519)
	// 	buf.Write([]byte(t))
	// default:
	// 	panic("unknown public key type")
	// }

	return buf.Bytes()
}

func GetCurveLabel(c elliptic.Curve) (byte, error) {
	return GetNamedCurveLabel(c.Params().Name)
}
func GetNamedCurveLabel(name string) (byte, error) {
	switch strings.ToUpper(name) {
	case strings.ToUpper(elliptic.P224().Params().Name):
		return P224, nil
	case strings.ToUpper(elliptic.P256().Params().Name):
		return P256, nil
	case strings.ToUpper(elliptic.P384().Params().Name):
		return P384, nil
	case strings.ToUpper(elliptic.P521().Params().Name):
		return P521, nil
	// case strings.ToUpper(sm2.SM2P256V1().Params().Name):
	// 	return SM2P256V1, nil
	default:
		return 0, errors.New("unsupported elliptic curve")
	}
}

func (self *ProgramBuilder) PushPubKey(pubkey *PublicKey) *ProgramBuilder {
	buf := SerializePublicKey(pubkey)
	return self.PushBytes(buf)
}

// GenerateKeyPair 根据t生成公钥私钥对.
// opts is the necessary parameter(s), which is defined by the key type:
//     ECDSA: a byte specifies the elliptic curve, which defined in package ec
//     SM2:   same as ECDSA
//     EdDSA: a byte specifies the curve, only ED25519 supported currently.
func GenerateKeyPair(t KeyType, opts interface{}) (*PrivateKey, *PublicKey, error) {
	switch t {
	case PK_ECDSA, PK_SM2:
		param, ok := opts.(byte)
		if !ok {
			return nil, nil, errors.New(err_generate + "invalid EC options, 1 byte curve label excepted")
		}
		c, err := GetCurve(param)
		if err != nil {
			return nil, nil, errors.New(err_generate + err.Error())
		}

		if t == PK_ECDSA {
			return GenerateECKeyPair(c, rand.Reader, ECDSA)
		} else {
			return GenerateECKeyPair(c, rand.Reader, SM2)
		}

	// case PK_EDDSA:
	// 	param, ok := opts.(byte)
	// 	if !ok {
	// 		return nil, nil, errors.New(err_generate + "invalid EdDSA option")
	// 	}

	// 	if param == ED25519 {
	// 		pub, pri, err := ed25519.GenerateKey(rand.Reader)
	// 		return pri, pub, err
	// 	} else {
	// 		return nil, nil, errors.New(err_generate + "unsupported EdDSA scheme")
	// 	}
	default:
		return nil, nil, errors.New(err_generate + "unknown algorithm")
	}
}

func GenerateECKeyPair(c elliptic.Curve, rand io.Reader, alg ECAlgorithm) (*PrivateKey, *PublicKey, error) {
	d, x, y, err := elliptic.GenerateKey(c, rand)
	if err != nil {
		return nil, nil, errors.New("Generate ec key pair failed, " + err.Error())
	}

	log.Debugf("private key length:%v", len(d))
	pri := PrivateKey{
		Algorithm: alg,
		PrivateKey: &ecdsa.PrivateKey{
			D: new(big.Int).SetBytes(d),
			PublicKey: ecdsa.PublicKey{
				X:     x,
				Y:     y,
				Curve: c,
			},
		},
	}
	pub := PublicKey{
		Algorithm: alg,
		PublicKey: &pri.PublicKey,
	}
	return &pri, &pub, nil
}

const ADDR_LEN = 20

type Address [ADDR_LEN]byte

func AddressFromVmCode(code []byte) Address {
	var addr Address
	temp := sha256.Sum256(code)
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:0])

	return addr
}

func (f *Address) ToBase58() string {
	data := append([]byte{23}, f[:]...)
	temp := sha256.Sum256(data)
	temps := sha256.Sum256(temp[:])
	data = append(data, temps[0:4]...)

	bi := new(big.Int).SetBytes(data).String()
	encoded, _ := base58.BitcoinEncoding.Encode([]byte(bi))
	return string(encoded)
}

func AddressFromPubKey(pubkey *PublicKey) Address {
	prog := ProgramFromPubKey(pubkey)

	return AddressFromVmCode(prog)
}

func ProgramFromPubKey(pubkey *PublicKey) []byte {
	sink := ZeroCopySink{}
	EncodeSinglePubKeyProgramInto(&sink, pubkey)
	return sink.Bytes()
}

func EncodeSinglePubKeyProgramInto(sink *ZeroCopySink, pubkey *PublicKey) {
	builder := ProgramBuilder{sink: sink}

	builder.PushPubKey(pubkey).PushOpCode(CHECKSIG)
}

// Encrypt the private key with the given password. The password is used to
// derive a key via scrypt function. AES with GCM mode is used for encryption.
// The first 12 bytes of the derived key is used as the nonce, and the last 32
// bytes is used as the encryption key.
func EncryptPrivateKey(pri *PrivateKey, addr string, pwd []byte) (*ProtectedKey, error) {
	return EncryptWithCustomScrypt(pri, addr, pwd, GetScryptParameters())
}

const (
	DEFAULT_N                  = 16384
	DEFAULT_R                  = 8
	DEFAULT_P                  = 8
	DEFAULT_DERIVED_KEY_LENGTH = 64
)

// Return the default parameters used in scrypt function
func GetScryptParameters() *ScryptParam {
	return &ScryptParam{
		N:     DEFAULT_N,
		R:     DEFAULT_R,
		P:     DEFAULT_P,
		DKLen: DEFAULT_DERIVED_KEY_LENGTH,
	}
}
func randomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type EncryptError struct {
	detail string
}

func (e *EncryptError) Error() string {
	return "encrypt private key error: " + e.detail
}

func NewEncryptError(msg string) *EncryptError {
	return &EncryptError{detail: msg}
}

type DecryptError EncryptError

func (e *DecryptError) Error() string {
	return "decrypt private key error: " + e.detail
}

func NewDecryptError(msg string) *DecryptError {
	return &DecryptError{detail: msg}
}

func EncryptWithCustomScrypt(pri *PrivateKey, addr string, pwd []byte, param *ScryptParam) (*ProtectedKey, error) {
	var res = ProtectedKey{
		Address: addr,
		EncAlg:  "aes-256-gcm",
	}

	salt, err := randomBytes(16)
	if err != nil {
		return nil, NewEncryptError(err.Error())
	}
	res.Salt = salt

	dkey, err := kdf(pwd, salt, param)
	if err != nil {
		return nil, NewEncryptError(err.Error())
	}
	nonce := dkey[:12]
	ekey := dkey[len(dkey)-32:]

	// Prepare the private key data for encryption
	var plaintext []byte
	// switch t := pri.(type) {
	// case *ec.PrivateKey:
	plaintext = pri.D.Bytes()
	switch pri.Algorithm {
	case ECDSA:
		res.Alg = "ECDSA"
	case SM2:
		res.Alg = "SM2"
	default:
		panic("unsupported ec algorithm")
	}
	res.Param = make(map[string]string)
	res.Param["curve"] = pri.Params().Name
	// case ed25519.PrivateKey:
	// 	plaintext = []byte(t)
	// 	res.Alg = "Ed25519"
	// default:
	// 	panic("unsupported key type")
	// }

	gcm, err := gcmCipher(ekey)
	if err != nil {
		return nil, NewEncryptError(err.Error())
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, []byte(addr))
	res.Key = ciphertext
	return &res, nil
}

func kdf(pwd []byte, salt []byte, param *ScryptParam) (dkey []byte, err error) {
	if param.DKLen < 32 {
		err = errors.New("derived key length too short")
		return
	}

	// Derive the encryption key
	dkey, err = scrypt.Key([]byte(pwd), salt, param.N, param.R, param.P, param.DKLen)
	return
}

func gcmCipher(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm, nil
}

type Identity struct {
	ID      string       `json:"ontid"`
	Label   string       `json:"label,omitempty"`
	Lock    bool         `json:"lock"`
	Control []Controller `json:"controls,omitempty"`
	Extra   interface{}  `json:"extra,omitempty"`
}

type Controller struct {
	ID     string `json:"id"`
	Public string `json:"publicKey,omitemtpy"`
	ProtectedKey
}

/* crypto object */
type Account struct {
	PrivateKey *PrivateKey
	PublicKey  *PublicKey
	Address    Address
	SigScheme  SignatureScheme
}

type unlockAccountInfo struct {
	acc        *Account
	unlockTime time.Time
	expiredAt  int //s
}

func FileExisted(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
