package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/mr-tron/base58/base58"
	"testing"
)

func TestHmacSHA1(t *testing.T) {
	var (
		appkey     = "2d68067484a20f1a346b3cf28a898ed7f5736f5bacf0fe60449da95efdb97ad4"
		appsecret  = "0dd1e322907ad7f55deaa35fec2aac97cae7931454d734364bc63f3e9b9f993a"
		timestamp  = "1506565393"
		period     = "3600"
		ciphertext []byte
	)

	ciphertext = HmacSHA1(appsecret, []byte(appkey+timestamp+period))

	fmt.Println("ciphertext = ", string(ciphertext))
}

func TestAESDecrypt(t *testing.T) {
	entext := "HQTqsjyaQUVjCT2gUoUvxSNAwUXic3qdNUPG18gZHYJYehtAULGnCjY2eYyt"
	keytext := "CxyAogk73eLTtQb8hmXvCm7ehb1tyfhv3NC6iD2nV2B9"
	bit, _ := base58.Decode(entext)
	key, _ := base58.Decode(keytext)
	fmt.Printf("key hex(%d): %s\n", len(key), hex.EncodeToString(key))
	fmt.Printf("bit hex(%d): %s\n", len(bit), hex.EncodeToString(bit))
	plaintext, _ := AESDecrypt(bit, key)

	t.Logf("plaintext: %s", string(plaintext))
}

func TestEncryptJSON(t *testing.T) {

	keytext := "CxyAogk73eLTtQb8hmXvCm7ehb1tyfhv3NC6iD2nV2B9"
	key, _ := base58.Decode(keytext)

	params := map[string]interface{}{
		"name": "chance",
		"age":  18,
	}

	plainText, _ := json.Marshal(params)
	//7b22616765223a31382c226e616d65223a226368616e6365227d
	//7b22616765223a31382c226e616d65223a226368616e6365227d
	fmt.Printf("plainText hex(%d): %s\n", len(plainText), hex.EncodeToString(plainText))
	chipText, _ := AESEncrypt(plainText, key)
	fmt.Printf("chipText hex(%d): %s\n", len(chipText), hex.EncodeToString(chipText))
	base := base58.Encode(chipText)

	fmt.Printf("base58 = %s\n", base)
	//1bf9d9ff6bbaf559e6b74479b0e770b480de59c5aa9ff044478834a0b1ec6986
	enbase, _ := base58.Decode(base)
	fmt.Printf("enbase hex(%d): %s\n", len(enbase), hex.EncodeToString(enbase))
	raw, _ := AESDecrypt(enbase, key)
	fmt.Printf("raw = %s\n", string(raw))
}

func TestMD5(t *testing.T) {
	hash := MD5([]byte("skaljfls2"))
	fmt.Println(hex.EncodeToString(hash))

	hash2 := GetMD5("skaljfls2")
	fmt.Println(hash2)
}

func TestHmac(t *testing.T) {
	//key := []byte("9988776655")
	data := []byte("123456")
	hash := owcrypt.Hmac(nil, data, owcrypt.HMAC_SHA256_ALG)
	fmt.Printf("hash: %s\n", hex.EncodeToString(hash))
}

func TestBlake2b(t *testing.T) {
	data, _ := hex.DecodeString("")
	hash := owcrypt.Hash(data, 32, owcrypt.HASH_ALG_BLAKE2B)
	fmt.Printf("hash: %s\n", hex.EncodeToString(hash))
	//ca2cc09e84e78d7bbcd58004721aec0526904e0379508e1e87716d85597e1862
	//ca2cc09e84e78d7bbcd58004721aec0526904e0379508e1e87716d85597e1862
}

func TestED25519(t *testing.T) {
	key, _ := hex.DecodeString("")
	pubkey, _ := owcrypt.GenPubkey(key, owcrypt.ECC_CURVE_ED25519_NORMAL)
	signature, _ := hex.DecodeString("")
	msg, _ := hex.DecodeString("")
	fmt.Printf("pubkey: %s \n", hex.EncodeToString(pubkey))
	flag := owcrypt.Verify(pubkey, nil, msg, signature, owcrypt.ECC_CURVE_ED25519_NORMAL)
	fmt.Printf("flag: %d \n", flag)
}

func TestBase64(t *testing.T) {
	base64Str := ""
	bs, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		t.Errorf("base64 decode failed")
		return
	}
	prv := bs[16:]
	fmt.Printf("hex: %s \n", hex.EncodeToString(prv))
}

func TestSECP256R1_sign(t *testing.T) {
	key, _ := hex.DecodeString("")
	pubkey, _ := owcrypt.GenPubkey(key, owcrypt.ECC_CURVE_SECP256R1)
	pubkey = owcrypt.PointCompress(pubkey, owcrypt.ECC_CURVE_SECP256R1)
	fmt.Printf("pubkey: %s \n", hex.EncodeToString(pubkey))
	msg, _ := hex.DecodeString("111111111111111111111111111111111111111111111111")
	sig, _, ret := owcrypt.Signature(key, nil, msg, owcrypt.ECC_CURVE_SECP256R1)
	fmt.Printf("ret: %d \n", ret)
	fmt.Printf("sig: %s \n", hex.EncodeToString(sig))
}

func TestSECP256R1_verify(t *testing.T) {
	prvkey, _ := hex.DecodeString("")
	pubkey, _ := owcrypt.GenPubkey(prvkey, owcrypt.ECC_CURVE_SECP256R1)
	//pubkey, _ := hex.DecodeString("0267b7dca2a5620195bc84fbc9ffe331dfe3c4ee01545cd814789c4b636e5f5eb0")
	compressedPubkey := owcrypt.PointCompress(pubkey, owcrypt.ECC_CURVE_SECP256R1)
	fmt.Printf("pubkey: %s \n", hex.EncodeToString(compressedPubkey))
	//03fd78eac9fd3216191f884e47378e940fba0919adaa92820c8f82a099d6c40474
	//msg, _ := hex.DecodeString("6cec7463040014002c012c011400000400")
	//msg_hash := owcrypt.Hash(msg, 0, owcrypt.HASH_ALG_SHA256)
	msg_hash, _ := hex.DecodeString("1a15624e8521c8b4575cd6f2c33a6e71c78de695bed596ac01584c31dbf18c32")
	fmt.Printf("hash: %s \n", hex.EncodeToString(msg_hash))
	//signature, _, _ := owcrypt.Signature(prvkey, nil, msg_hash, owcrypt.ECC_CURVE_SECP256R1)
	//13a4030ffc51f7b3d9dafa60f4a567a34911972fb48103fabd56766e9401373b492e5522c63c431468e8d7bc5d655edc9c264dd768bfdf79232616d1bb97d3ff
	signature, _ := hex.DecodeString("e51f7701d9e1c0c10d7512b7d3d2781e661c15cd1d74ca7ac0de9d9f5d71e03c18b4118310bcfdac346bf0b5aa66744223eb557b1dec68a24c88b2ce1f887e78")
	fmt.Printf("signature: %s \n", hex.EncodeToString(signature))

	ret := owcrypt.Verify(pubkey, nil, msg_hash, signature, owcrypt.ECC_CURVE_SECP256R1)
	fmt.Printf("ret: %d \n", ret)
}

func TestSECP256K1(t *testing.T) {
	key, _ := hex.DecodeString("")
	pubkey, _ := owcrypt.GenPubkey(key, owcrypt.ECC_CURVE_SECP256K1)
	//pubkey = owcrypt.PointCompress(pubkey, owcrypt.ECC_CURVE_SECP256K1)
	fmt.Printf("pubkey: %s \n", hex.EncodeToString(pubkey))
	msg, _ := hex.DecodeString("b68a5c3e0fdffecc81459c522bfcdeeb787da05cfbcd59574337d7d2e376967e")
	//sig, _, _ := owcrypt.Signature(key, nil, msg, owcrypt.ECC_CURVE_SECP256K1)
	sig, _ := hex.DecodeString("6a6f5c658ef5c01a75d1474c91c4ef7bb8970dd156906fa7837e84be6b3bf5a61b6bdbaa9bcd2365cd29e919c50dafe68b3cedcc6e38feeddbac6bd4053de7ad")
	//sig = append(sig, v)
	ret := owcrypt.Verify(pubkey, nil, msg, sig, owcrypt.ECC_CURVE_SECP256K1)
	fmt.Printf("ret: %d \n", ret)
	//fmt.Printf("sig: %s \n", hex.EncodeToString(sig))
}
